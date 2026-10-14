#pragma once

#include <arpa/inet.h>
#include <netinet/in.h>
#include <strings.h>
#include <sys/stat.h>
#include <unistd.h>

#include <string>

namespace rest_tls {

inline bool is_https_url(const std::string& url)
{
    static const char scheme[] = "https://";
    return url.size() >= sizeof(scheme) - 1 &&
           strncasecmp(url.c_str(), scheme, sizeof(scheme) - 1) == 0;
}

inline bool get_url_host(const std::string& url, std::string& host)
{
    const std::string::size_type scheme_end = url.find("://");
    if (scheme_end == std::string::npos) {
        return false;
    }

    const std::string::size_type authority_start = scheme_end + 3;
    const std::string::size_type authority_end = url.find_first_of("/?#", authority_start);
    std::string authority = url.substr(authority_start, authority_end - authority_start);
    if (authority.empty()) {
        return false;
    }

    const std::string::size_type userinfo_end = authority.rfind('@');
    if (userinfo_end != std::string::npos) {
        authority.erase(0, userinfo_end + 1);
    }

    if (authority.empty()) {
        return false;
    }

    if (authority[0] == '[') {
        const std::string::size_type bracket_end = authority.find(']');
        if (bracket_end == std::string::npos) {
            return false;
        }
        if (bracket_end + 1 < authority.size() && authority[bracket_end + 1] != ':') {
            return false;
        }
        host = authority.substr(1, bracket_end - 1);
    } else {
        const std::string::size_type port_start = authority.find(':');
        if (port_start != std::string::npos && authority.find(':', port_start + 1) != std::string::npos) {
            return false;
        }
        host = authority.substr(0, port_start);
    }

    return !host.empty();
}

inline bool is_loopback_url(const std::string& url)
{
    std::string host;
    if (!get_url_host(url, host)) {
        return false;
    }

    if (strcasecmp(host.c_str(), "localhost") == 0) {
        return true;
    }

    struct in_addr ipv4;
    if (inet_pton(AF_INET, host.c_str(), &ipv4) == 1) {
        return (ntohl(ipv4.s_addr) >> 24) == 127;
    }

    struct in6_addr ipv6;
    if (inet_pton(AF_INET6, host.c_str(), &ipv6) != 1) {
        return false;
    }

    return IN6_IS_ADDR_LOOPBACK(&ipv6) ||
           (IN6_IS_ADDR_V4MAPPED(&ipv6) && ipv6.s6_addr[12] == 127);
}

inline bool is_readable_ca_file(const char *path)
{
    /* The path must remain operator-controlled between this check and libcurl's open. */
    struct stat file_stat;
    return path && path[0] != '\0' && stat(path, &file_stat) == 0 &&
           S_ISREG(file_stat.st_mode) && access(path, R_OK) == 0;
}

}  // namespace rest_tls
