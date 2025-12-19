# observabilityd (sonic-mgmt-framework)

Small daemon that consumes CONFIG_DB updates for:

- `OBSERVABILITY_CONNECTOR|splunk` (`url`, `token`)
- `OBSERVABILITY_CONNECTOR|datadog` (`url`, `api_key`, `app_key`)

and posts a JSON event to the configured backend whenever the connector config changes.

## How it watches ConfigDB

- **Primary**: Redis keyspace notifications (pattern subscribe on `__keyspace@4__:OBSERVABILITY_CONNECTOR|*`)
- **Fallback**: periodic polling (default `--poll=5s`)

Keyspace notifications require Redis `notify-keyspace-events` to include keyspace/hash events (e.g. `KEA` or at least `Kh`).

## Build

```bash
cd sonic-mgmt-framework/observabilityd
go mod tidy
go build -o observabilityd .
```

## Run

```bash
./observabilityd --redis-sock /var/run/redis/redis.sock --db 4 --poll 5s
```

If you want TCP redis:

```bash
./observabilityd --redis-addr 127.0.0.1:6379 --db 4
```

## Notes

- `splunk url` is expected to be a full HTTP endpoint (e.g. Splunk HEC).
- `datadog url` is treated as a generic HTTP JSON endpoint; the daemon sends `DD-API-KEY` and `DD-APPLICATION-KEY` headers.


