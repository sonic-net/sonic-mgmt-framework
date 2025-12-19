package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	configDBDefault = 4
	stateDBDefault  = 6
	tablePrefix     = "OBSERVABILITY_CONNECTOR|"
)

type Provider string

const (
	ProviderSplunk  Provider = "splunk"
	ProviderDatadog Provider = "datadog"
)

type ConnectorConfig struct {
	Provider Provider
	URL      string
	// Splunk
	Token string
	// Datadog
	APIKey string
	AppKey string
}

func (c ConnectorConfig) isConfigured() bool {
	return strings.TrimSpace(c.URL) != "" && (c.Provider == ProviderSplunk && c.Token != "" ||
		c.Provider == ProviderDatadog && c.APIKey != "" && c.AppKey != "")
}

func (c ConnectorConfig) redact() map[string]string {
	out := map[string]string{
		"provider": string(c.Provider),
		"url":      c.URL,
	}
	switch c.Provider {
	case ProviderSplunk:
		out["token"] = maskSecret(c.Token)
	case ProviderDatadog:
		out["api_key"] = maskSecret(c.APIKey)
		out["app_key"] = maskSecret(c.AppKey)
	}
	return out
}

type Daemon struct {
	rdb          *redis.Client
	stateRdb     *redis.Client
	httpClient   *http.Client
	pollInterval time.Duration
	healthEvery  time.Duration

	lastHash string
	lastCfg  map[Provider]ConnectorConfig
}

func NewDaemon(rdb *redis.Client, stateRdb *redis.Client, pollInterval, healthEvery time.Duration) *Daemon {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Daemon{
		rdb:          rdb,
		stateRdb:     stateRdb,
		httpClient:   &http.Client{Timeout: 10 * time.Second, Transport: t},
		pollInterval: pollInterval,
		healthEvery:  healthEvery,
		lastCfg:      map[Provider]ConnectorConfig{},
	}
}

func main() {
	var redisSock string
	var redisAddr string
	var db int
	var stateDB int
	var poll time.Duration
	var healthEvery time.Duration
	var logJSON bool

	flag.StringVar(&redisSock, "redis-sock", "/var/run/redis/redis.sock", "Redis unix socket path (CONFIG_DB)")
	flag.StringVar(&redisAddr, "redis-addr", "", "Redis TCP addr (host:port). If set, unix socket is not used.")
	flag.IntVar(&db, "db", configDBDefault, "Redis DB number for CONFIG_DB (default 4)")
	flag.IntVar(&stateDB, "state-db", stateDBDefault, "Redis DB number for STATE_DB (default 6)")
	flag.DurationVar(&poll, "poll", 5*time.Second, "Polling interval fallback (also used for periodic refresh)")
	flag.DurationVar(&healthEvery, "health-every", 60*time.Second, "Healthcheck interval")
	flag.BoolVar(&logJSON, "log-json", false, "Emit structured JSON logs")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	if logJSON {
		logger.SetFlags(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts := &redis.Options{DB: db}
	if redisAddr != "" {
		opts.Addr = redisAddr
	} else {
		opts.Network = "unix"
		opts.Addr = redisSock
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		fatal(logger, logJSON, "redis_ping_failed", map[string]any{"error": err.Error()})
	}

	stateOpts := &redis.Options{DB: stateDB}
	if redisAddr != "" {
		stateOpts.Addr = redisAddr
	} else {
		stateOpts.Network = "unix"
		stateOpts.Addr = redisSock
	}

	stateRdb := redis.NewClient(stateOpts)
	if err := stateRdb.Ping(ctx).Err(); err != nil {
		fatal(logger, logJSON, "state_redis_ping_failed", map[string]any{"error": err.Error()})
	}

	d := NewDaemon(rdb, stateRdb, poll, healthEvery)
	info(logger, logJSON, "daemon_start", map[string]any{
		"db":          db,
		"state_db":    stateDB,
		"redis_addr":  redisAddr,
		"redis_sock":  redisSock,
		"poll":        poll.String(),
		"healthEvery": healthEvery.String(),
		"tablePrefix": tablePrefix,
	})

	go d.runKeyspaceSubscriber(ctx, logger, logJSON, db)
	go d.runHealthLoop(ctx, logger, logJSON)

	if err := d.runPollLoop(ctx, logger, logJSON); err != nil && !errorsIsCtxCanceled(err, ctx) {
		fatal(logger, logJSON, "poll_loop_failed", map[string]any{"error": err.Error()})
	}

	info(logger, logJSON, "daemon_exit", nil)
}

func (d *Daemon) runPollLoop(ctx context.Context, logger *log.Logger, logJSON bool) error {
	t := time.NewTicker(d.pollInterval)
	defer t.Stop()

	if err := d.syncOnce(ctx, logger, logJSON, "startup"); err != nil {
		warn(logger, logJSON, "sync_failed", map[string]any{"error": err.Error()})
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := d.syncOnce(ctx, logger, logJSON, "poll"); err != nil {
				warn(logger, logJSON, "sync_failed", map[string]any{"error": err.Error()})
			}
		}
	}
}

func (d *Daemon) runKeyspaceSubscriber(ctx context.Context, logger *log.Logger, logJSON bool, db int) {
	pattern := fmt.Sprintf("__keyspace@%d__:%s*", db, tablePrefix)
	pubsub := d.rdb.PSubscribe(ctx, pattern)
	defer func() { _ = pubsub.Close() }()

	ch := pubsub.Channel()
	info(logger, logJSON, "keyspace_subscribe", map[string]any{"pattern": pattern})

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok || msg == nil {
				return
			}
			info(logger, logJSON, "keyspace_event", map[string]any{
				"channel": msg.Channel,
				"event":   msg.Payload,
			})
			if err := d.syncOnce(ctx, logger, logJSON, "keyspace"); err != nil {
				warn(logger, logJSON, "sync_failed", map[string]any{"error": err.Error()})
			}
		}
	}
}

func (d *Daemon) syncOnce(ctx context.Context, logger *log.Logger, logJSON bool, reason string) error {
	cfgs, hash, err := d.readConfigs(ctx)
	if err != nil {
		return err
	}
	if hash == d.lastHash {
		return nil
	}

	prev := d.lastCfg
	d.lastHash = hash
	d.lastCfg = cfgs

	info(logger, logJSON, "config_changed", map[string]any{
		"reason": reason,
		"hash":   hash,
		"cfgs": map[string]any{
			"splunk":  cfgs[ProviderSplunk].redact(),
			"datadog": cfgs[ProviderDatadog].redact(),
		},
	})

	evt := map[string]any{
		"event_type": "connector_config_changed",
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
		"configs": map[string]any{
			"splunk":  cfgs[ProviderSplunk].redact(),
			"datadog": cfgs[ProviderDatadog].redact(),
		},
		"previous_configured": map[string]bool{
			"splunk":  prev[ProviderSplunk].isConfigured(),
			"datadog": prev[ProviderDatadog].isConfigured(),
		},
	}

	if cfgs[ProviderSplunk].isConfigured() {
		if err := d.postToSplunk(ctx, cfgs[ProviderSplunk], evt); err != nil {
			warn(logger, logJSON, "post_failed", map[string]any{"provider": "splunk", "error": err.Error()})
		} else {
			info(logger, logJSON, "post_ok", map[string]any{"provider": "splunk"})
		}
	}
	if cfgs[ProviderDatadog].isConfigured() {
		if err := d.postToDatadog(ctx, cfgs[ProviderDatadog], evt); err != nil {
			warn(logger, logJSON, "post_failed", map[string]any{"provider": "datadog", "error": err.Error()})
		} else {
			info(logger, logJSON, "post_ok", map[string]any{"provider": "datadog"})
		}
	}

	return nil
}

func (d *Daemon) runHealthLoop(ctx context.Context, logger *log.Logger, logJSON bool) {
	t := time.NewTicker(d.healthEvery)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cfgs, _, err := d.readConfigs(ctx)
			if err != nil {
				warn(logger, logJSON, "health_read_config_failed", map[string]any{"error": err.Error()})
				continue
			}

			for _, p := range []Provider{ProviderSplunk, ProviderDatadog} {
				cfg := cfgs[p]
				if !cfg.isConfigured() {
					continue
				}

				ok, detail := d.healthCheckEndpoint(ctx, cfg)
				status := "OK"
				if !ok {
					status = "FAIL"
					if detail != "" {
						status = status + " (" + detail + ")"
					}
				}

				if err := d.writeHealthToStateDB(ctx, p, status); err != nil {
					warn(logger, logJSON, "health_state_write_failed", map[string]any{"provider": string(p), "error": err.Error()})
				} else {
					info(logger, logJSON, "health_state_written", map[string]any{"provider": string(p), "status": status})
				}
			}
		}
	}
}

func (d *Daemon) healthCheckEndpoint(ctx context.Context, cfg ConnectorConfig) (bool, string) {
	var url string
	var method string

	switch cfg.Provider {
	case ProviderSplunk:
		// Splunk healthcheck: GET /services/collector/health
		url = strings.TrimSuffix(cfg.URL, "/") + "/services/collector/health"
		method = http.MethodGet
	case ProviderDatadog:
		// Datadog: use HEAD on base URL
		url = cfg.URL
		method = http.MethodHead
	default:
		return false, "unknown_provider"
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return false, "bad_url"
	}

	switch cfg.Provider {
	case ProviderSplunk:
		req.Header.Set("Authorization", "Splunk "+cfg.Token)
	case ProviderDatadog:
		req.Header.Set("DD-API-KEY", cfg.APIKey)
		req.Header.Set("DD-APPLICATION-KEY", cfg.AppKey)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return false, "conn_error"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, ""
	}
	return false, fmt.Sprintf("http_%d", resp.StatusCode)
}

func (d *Daemon) writeHealthToStateDB(ctx context.Context, provider Provider, status string) error {
	key := tablePrefix + string(provider) // OBSERVABILITY_CONNECTOR|<provider>
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := d.stateRdb.HSet(ctx, key, map[string]any{
		"LAST_HEALTH_TIME":   now,
		"LAST_HEALTH_STATUS": status,
	}).Result()
	return err
}

func (d *Daemon) readConfigs(ctx context.Context) (map[Provider]ConnectorConfig, string, error) {
	cfgs := map[Provider]ConnectorConfig{
		ProviderSplunk:  {Provider: ProviderSplunk},
		ProviderDatadog: {Provider: ProviderDatadog},
	}

	for _, p := range []Provider{ProviderSplunk, ProviderDatadog} {
		key := tablePrefix + string(p)
		m, err := d.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, "", err
		}
		if len(m) == 0 {
			continue
		}
		cfg := cfgs[p]
		cfg.URL = strings.TrimSpace(m["url"])
		switch p {
		case ProviderSplunk:
			cfg.Token = m["token"]
		case ProviderDatadog:
			cfg.APIKey = m["api_key"]
			cfg.AppKey = m["app_key"]
		}
		cfgs[p] = cfg
	}

	h := sha256.New()
	for _, p := range []Provider{ProviderSplunk, ProviderDatadog} {
		cfg := cfgs[p]
		lines := []string{
			"provider=" + string(cfg.Provider),
			"url=" + cfg.URL,
			"token=" + cfg.Token,
			"api_key=" + cfg.APIKey,
			"app_key=" + cfg.AppKey,
		}
		sort.Strings(lines)
		for _, ln := range lines {
			_, _ = h.Write([]byte(ln))
			_, _ = h.Write([]byte{'\n'})
		}
	}
	return cfgs, hex.EncodeToString(h.Sum(nil)), nil
}

func (d *Daemon) postToSplunk(ctx context.Context, cfg ConnectorConfig, evt map[string]any) error {
	body := map[string]any{
		"time":  float64(time.Now().UTC().Unix()),
		"event": evt,
	}
	b, _ := json.Marshal(body)

	// Splunk events: POST /services/collector/events
	url := strings.TrimSuffix(cfg.URL, "/") + "/services/collector/events"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+cfg.Token)

	return doRequest(d.httpClient, req)
}

func (d *Daemon) postToDatadog(ctx context.Context, cfg ConnectorConfig, evt map[string]any) error {
	b, _ := json.Marshal(evt)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", cfg.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", cfg.AppKey)

	return doRequest(d.httpClient, req)
}

func doRequest(c *http.Client, req *http.Request) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
}

func maskSecret(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return strings.Repeat("*", len(v))
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}

func errorsIsCtxCanceled(err error, ctx context.Context) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	return ctx != nil && ctx.Err() != nil
}

func info(logger *log.Logger, jsonMode bool, msg string, fields map[string]any) {
	logKV(logger, jsonMode, "info", msg, fields)
}

func warn(logger *log.Logger, jsonMode bool, msg string, fields map[string]any) {
	logKV(logger, jsonMode, "warn", msg, fields)
}

func fatal(logger *log.Logger, jsonMode bool, msg string, fields map[string]any) {
	logKV(logger, jsonMode, "error", msg, fields)
	os.Exit(1)
}

func logKV(logger *log.Logger, jsonMode bool, level, msg string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["level"] = level
	fields["msg"] = msg
	fields["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	if jsonMode {
		b, _ := json.Marshal(fields)
		logger.Print(string(b))
		return
	}
	logger.Printf("%s: %s %v", level, msg, fields)
}
