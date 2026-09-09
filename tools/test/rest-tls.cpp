#include <cstdio>
#include <cstdlib>
#include <string>
#include <unistd.h>

#include "../../CLI/klish/patches/klish-2.1.4/plugins/clish/rest_tls.h"

static int failures = 0;

static void expect_loopback(const char *url, bool expected)
{
    const bool actual = rest_tls::is_loopback_url(url);
    if (actual != expected) {
        std::fprintf(stderr, "is_loopback_url(%s): expected %d, got %d\n",
                     url, expected, actual);
        ++failures;
    }
}

static void expect_readable_ca(const char *path, bool expected)
{
    const bool actual = rest_tls::is_readable_ca_file(path);
    if (actual != expected) {
        std::fprintf(stderr, "is_readable_ca_file(%s): expected %d, got %d\n",
                     path ? path : "null", expected, actual);
        ++failures;
    }
}

int main()
{
    expect_loopback("https://localhost", true);
    expect_loopback("HTTPS://LOCALHOST:443/restconf", true);
    expect_loopback("https://127.0.0.1", true);
    expect_loopback("https://127.255.255.254:8443", true);
    expect_loopback("https://[::1]", true);
    expect_loopback("https://[::1]:8443/restconf", true);
    expect_loopback("https://[::ffff:127.0.0.1]", true);
    expect_loopback("https://[::ffff:127.255.255.254]:8443", true);

    expect_loopback("https://localhost.example.com", false);
    expect_loopback("https://127.0.0.1.example.com", false);
    expect_loopback("https://128.0.0.1", false);
    expect_loopback("https://[::2]", false);
    expect_loopback("https://[::ffff:128.0.0.1]", false);
    expect_loopback("https://example.com", false);
    expect_loopback("https://localhost@remote.example", false);
    expect_loopback("https://[::1.example.com", false);
    expect_loopback("not-a-url", false);

    char ca_path[] = "/tmp/rest-tls-ca-XXXXXX";
    const int ca_fd = mkstemp(ca_path);
    if (ca_fd == -1) {
        std::perror("mkstemp");
        return EXIT_FAILURE;
    }
    close(ca_fd);

    expect_readable_ca(ca_path, true);
    expect_readable_ca(NULL, false);
    expect_readable_ca("", false);
    expect_readable_ca("/path/that/does/not/exist", false);
    expect_readable_ca("/tmp", false);

    unlink(ca_path);
    return failures == 0 ? EXIT_SUCCESS : EXIT_FAILURE;
}
