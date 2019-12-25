/**
 * This is an NSS module. It allows applications running in containers to
 * access the user account information on the host. A service on the host
 * called "hamd" for Host Account Management Daemon is contacted over DBus
 * to retrieve user account information.
 *
 * This code is compiled as libnss_ham.so.2. In order to use it one must
 * use the keyword "ham" in the NSS configuration file (i.e.
 * /etc/nsswitch.conf) as follows:
 *
 * /etc/nsswitch.conf
 * ==================
 *
 *     passwd:   compat ham    <- To support getpwnam() and getpwuid()
 *     group:    compat ham    <- To support getgrnam() and getgrgid()
 *     shadow:   compat ham    <- To support getspnam()

 * The Naming Scheme of the NSS Modules
 * ====================================
 *
 * The name of each function consists of various parts:
 *
 *     _nss_[service]_[function]
 *
 * [service] of course corresponds to the name of the module this function
 * is found in. The [function] part is derived from the interface function
 * in the C library itself. If the user calls the function gethostbyname()
 * and the [service] used is "files" the function
 *
 *     _nss_files_gethostbyname_r
 *
 * in the module
 *
 *     libnss_files.so.2
 *
 * is used. You see, what is explained above in not the whole truth. In
 * fact the NSS modules only contain reentrant versions of the lookup
 * functions. I.e., if the user would call the gethostbyname_r function
 * this also would end in the above function. For all user interface
 * functions the C library maps this call to a call to the reentrant
 * function. For reentrant functions this is trivial since the interface is
 * (nearly) the same. For the non-reentrant version the library keeps
 * internal buffers which are used to replace the user supplied
 * buffer.
 *
 * I.e., the reentrant functions can have counterparts. No service module
 * is forced to have functions for all databases and all kinds to access
 * them. If a function is not available it is simply treated as if the
 * function would return unavail (see Actions in the NSS configuration).
 *
 * Error codes
 * ===========
 *
 * In case the interface function has to return an error it is important
 * that the correct error code is stored in *errnop. Some return status
 * values have only one associated error code, others have more.
 *
 * enum nss_status      errno   Description
 * ===================  ======= ==========================================
 * NSS_STATUS_TRYAGAIN  EAGAIN  One of the functions used ran temporarily
 *                              out of resources or a service is
 *                              currently not available.
 *
 *                      ERANGE  The provided buffer is not large enough.
 *                              The function should be called again with a
 *                              larger buffer.
 *
 * NSS_STATUS_UNAVAIL   ENOENT  A necessary input file cannot be found.
 *
 * NSS_STATUS_NOTFOUND  ENOENT  The requested entry is not available.
 *
 * NSS_STATUS_NOTFOUND  SUCCESS There are no entries. Use this to avoid
 *                              returning errors for inactive services
 *                              which may be enabled at a later time. This
 *                              is not the same as the service being
 *                              temporarily unavailable.
 */

#ifndef _GNU_SOURCE
#   define _GNU_SOURCE
#endif
#include <nss.h>                    /* NSS_STATUS_SUCCESS, NSS_STATUS_NOTFOUND */
#include <pwd.h>                    /* uid_t, struct passwd */
#include <grp.h>                    /* gid_t, struct group */
#include <shadow.h>                 /* struct spwd */
#include <stddef.h>                 /* NULL */
#include <sys/stat.h>
#include <sys/types.h>              /* getpid() */
#include <unistd.h>                 /* getpid() */
#include <errno.h>                  /* errno */
#include <string.h>                 /* strdup() */
#include <fcntl.h>                  /* open(), O_RDONLY */
#include <stdio.h>                  /* fopen(), fclose() */
#include <limits.h>                 /* LINE_MAX */
#include <stdint.h>                 /* uint32_t */
#include <syslog.h>                 /* syslog() */
#include <string>                   /* std::string */
#include <vector>                   /* std::vector */
#include <numeric>                  /* std::accumulate() */
#include <dbus-c++/dbus.h>          /* DBus */

#include "../shared/utils.h"        /* strneq(), startswith(), cpy2buf(), join() */
#include "../shared/dbus-address.h" /* DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE */

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-but-set-variable"   /* SUPPRESS: warning: variable 'ri' set but not used [-Wunused-but-set-variable] */
#include "../shared/org.SONiC.HostAccountManagement.dbus-proxy.h"
#pragma GCC diagnostic pop

class name_service_proxy_c : public ham::name_service_proxy,
                             public DBus::IntrospectableProxy,
                             public DBus::ObjectProxy
{
public:
    name_service_proxy_c(DBus::Connection &conn, const char * dbus_bus_name_p, const char * dbus_obj_name_p) :
    DBus::ObjectProxy(conn, dbus_obj_name_p, dbus_bus_name_p)
    {
    }
};

#define SYSLOG(LEVEL, args...) \
do \
{ \
    if (verbose) \
    { \
        if (log_p != nullptr) fprintf(log_p, args); \
        else                  syslog(LEVEL, args);  \
    } \
} while (0)


/**
 * @brief Extract cmdline from /proc/self/cmdline. This is only needed for
 *        debugging purposes
 */
static char * cmdline_p = program_invocation_name;
static void read_cmdline()
{
    const char * fn = "/proc/self/cmdline";
    struct stat  st;
    if (stat(fn, &st) != 0)
        return;

    int fd = open(fn, O_RDONLY);
    if (fd == -1)
        return;

    char      buffer[LINE_MAX];
    size_t    sz = sizeof(buffer);
    size_t    n  = 0;
    buffer[0] = '\0';
    for(;;)
    {
        ssize_t r = read(fd, &buffer[n], sz-n);
        if (r == -1)
        {
            if (errno == EINTR) continue;
            break;
        }
        n += r;
        if (n == sz) break; // filled the buffer
        if (r == 0)  break; // EOF
    }
    close(fd);

    if (n)
    {
        if (n == sz) n--;
        buffer[n] = '\0';
        size_t i = n;
        while (i--)
        {
            int c = buffer[i];
            if ((c < ' ') || (c > '~')) buffer[i] = ' ';
        }

        // Delete trailing spaces, tabs, and newline chars.
        char * p = &buffer[strcspn(buffer, "\n\r")];
        while ((p >= buffer) && (*p == ' ' || *p == '\t' || *p == '\0'))
        {
            *p-- = '\0';
        }
    }

    cmdline_p = strdup(buffer);
}

/**
 * @brief Get configuration parameters for this module. Parameters are
 *        retrieved from the file "/etc/libnss_ham.conf"
 */
static FILE * log_p     = nullptr;
static bool   verbose   = false;
static void read_config()
{
    FILE * file = fopen("/etc/libnss_ham.conf", "re");
    if (file)
    {
        #define WHITESPACE " \t\n\r"
        char    line[LINE_MAX];
        char  * p;
        char  * s;
        while (NULL != (p = fgets(line, sizeof line, file)))
        {
            p += strspn(p, WHITESPACE);            // Remove leading newline and spaces
            if (*p == '#' || *p == '\0') continue; // Skip comments and empty lines

            if (NULL != (s = startswith(p, "debug")))
            {
                s += strspn(s, " \t=");
                if (strneq(s, "yes", 3))
                    verbose = true;
            }
            else if (NULL != (s = startswith(p, "log")))
            {
                s += strspn(s, " \t=");
                s[strcspn(s, WHITESPACE)] = '\0'; // Remove trailing newline and spaces
                if (*s != '\0')
                {
                    if (log_p != nullptr)
                    {
                        fclose(log_p);
                        log_p = nullptr;
                    }
                    log_p = fopen(s, "w");
                }
            }
        }

        fclose(file);
    }

    if (verbose)
    {
        read_cmdline();
    }
}

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Initalize module singletons on entry.
 */
void __attribute__((constructor)) __module_enter(void)
{
    read_config();
}

/**
 * @brief Module clean up on exit
 */
void __attribute__((destructor)) __module_exit(void)
{
    if ((cmdline_p != nullptr) && (cmdline_p != program_invocation_name))
    {
        free(cmdline_p);
        cmdline_p = nullptr;
    }

    if (log_p != nullptr)
    {
        fclose(log_p);
        log_p = nullptr;
    }
}

#ifdef __cplusplus
}
#endif

/**
 * @brief   Retrieve current system monotonic clock as a 64-bit count in
 *          nano-sec.
 */
uint64_t get_now_nsec()
{
    struct timespec now;
    clock_gettime(CLOCK_MONOTONIC, &now);
    return (now.tv_sec * 1000000000ULL) + now.tv_nsec;
}

/**
 * @brief   Fill structure pointed to by @result with the data contained in
 *          the structure pointed to by @ham_data_r.
 *
 * @details The string fields pointed to by the members of the passwd
 *          structure are stored in @buffer of size buflen. In case @buffer
 *          has insufficient memory to hold the strings of struct passwd,
 *          the ENOMEM error will be reported.
 *
 * @param ham_data_r  This is the data we received from hamd.
 * @param result      Pointer to destination where data gets copied
 * @param buffer      Pointer to memory where strings can be stored
 * @param buflen      Size of buffer
 * @param errnop      Pointer to where errno code can be written
 *
 * @return - If no entry was found, return NSS_STATUS_NOTFOUND and set
 *           errno=ENOENT.
 *         - If @buffer has insufficient memory to hold the strings of
 *           struct passwd then return NSS_STATUS_TRYAGAIN and set
 *           errno=ENOMEM.
 *         - Otherwise return NSS_STATUS_SUCCESS and set errno to 0.
 */
static enum nss_status pw_fill_result(const char           * fn_p,
                                      double                 duration_msec,
                                      const ::DBus::Struct <
                                         bool,        /* _1: success   */
                                         std::string, /* _2: pw_name   */
                                         std::string, /* _3: pw_passwd */
                                         uint32_t,    /* _4: pw_uid    */
                                         uint32_t,    /* _5: pw_gid    */
                                         std::string, /* _6: pw_gecos  */
                                         std::string, /* _7: pw_dir    */
                                         std::string  /* _8: pw_shell  */
                                      >                    & ham_data_r,
                                      struct passwd        * result,
                                      char                 * buffer,
                                      size_t                 buflen,
                                      int                  * errnop)
{
    bool success = ham_data_r._1;
    if (!success)
    {
        SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" exec time=%.3f ms. Failed!\n", fn_p, cmdline_p, duration_msec);
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    const std::string & pw_name_r   = ham_data_r._2;
    const std::string & pw_passwd_r = ham_data_r._3;
    uid_t               pw_uid      = ham_data_r._4;
    gid_t               pw_gid      = ham_data_r._5;
    const std::string & pw_gecos_r  = ham_data_r._6;
    const std::string & pw_dir_r    = ham_data_r._7;
    const std::string & pw_shell_r  = ham_data_r._8;

    SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" exec time=%.3f ms. Success! pw_name=\"%s\", pw_passwd=\"%s\", pw_uid=%u, pw_gid=%u, pw_gecos=\"%s\", pw_dir=%s, pw_shell=\"%s\"\n",
           fn_p, cmdline_p, duration_msec, pw_name_r.c_str(), pw_passwd_r.c_str(), pw_uid, pw_gid, pw_gecos_r.c_str(), pw_dir_r.c_str(), pw_shell_r.c_str());

    size_t name_l   = pw_name_r.length()   + 1; /* +1 to include NUL terminating char */
    size_t dir_l    = pw_dir_r.length()    + 1;
    size_t shell_l  = pw_shell_r.length()  + 1;
    size_t passwd_l = pw_passwd_r.length() + 1;
    size_t gecos_l  = pw_gecos_r.length()  + 1;
    if (buflen < (name_l + shell_l + dir_l + passwd_l + gecos_l))
    {
        SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" not enough memory for struct passwd data\n", fn_p, cmdline_p);

        if (errnop) *errnop = ENOMEM;
        return NSS_STATUS_TRYAGAIN;
    }

    result->pw_uid    = pw_uid;
    result->pw_gid    = pw_gid;
    result->pw_name   = buffer;
    result->pw_dir    = cpy2buf(result->pw_name,   pw_name_r.c_str(),   name_l);
    result->pw_shell  = cpy2buf(result->pw_dir,    pw_dir_r.c_str(),    dir_l);
    result->pw_passwd = cpy2buf(result->pw_shell,  pw_shell_r.c_str(),  shell_l);
    result->pw_gecos  = cpy2buf(result->pw_passwd, pw_passwd_r.c_str(), passwd_l);
    cpy2buf(result->pw_gecos, pw_gecos_r.c_str(), gecos_l);

    if (errnop) *errnop = 0;

    return NSS_STATUS_SUCCESS;
}

/**
 * @brief Fill structure pointed to by @result with the data contained in
 *        the structure pointed to by @ham_data_r.
 *
 * @details The string fields pointed to by the members of the group
 *          structure are stored in @buffer of size buflen. In case @buffer
 *          has insufficient memory to hold the strings of struct group,
 *          the ENOMEM error will be reported.
 *
 * @param ham_data_r   This is the data we received from hamd.
 * @param result       Pointer to destination where data gets copied
 * @param buffer       Pointer to memory where strings can be stored
 * @param buflen       Size of buffer
 * @param errnop       Pointer to where errno code can be written
 *
 * @return - If no entry was found, return NSS_STATUS_NOTFOUND and set
 *           errno=ENOENT.
 *         - If @buffer has insufficient memory to hold the strings of
 *           struct passwd then return NSS_STATUS_TRYAGAIN and set
 *           errno=ENOMEM.
 *         - Otherwise return NSS_STATUS_SUCCESS and set errno to 0.
 */
static enum nss_status gr_fill_result(const char                         * fn_p,
                                      double                               duration_msec,
                                      const ::DBus::Struct <
                                         bool,                      /* _1: success   */
                                         std::string,               /* _2: gr_name   */
                                         std::string,               /* _3: gr_passwd */
                                         uint32_t,                  /* _4: gr_gid    */
                                         std::vector< std::string > /* _5: gr_mem    */
                                      >                                  & ham_data_r,
                                      struct group                       * result,
                                      char                               * buffer,
                                      size_t                               buflen,
                                      int                                * errnop)
{
    bool success = ham_data_r._1;

    if (!success)
    {
        SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" exec time=%.3f ms. Failed!\n", fn_p, cmdline_p, duration_msec);
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    const std::string                & gr_name_r   = ham_data_r._2;
    const std::string                & gr_passwd_r = ham_data_r._3;
    gid_t                              gr_gid      = ham_data_r._4;
    const std::vector< std::string > & gr_mem_r    = ham_data_r._5;

    SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" exec time=%.3f ms. Success! gr_name=\"%s\", pw_passwd=\"%s\", gr_gid=%u, pw_mem=[\"%s\"]\n",
           fn_p, cmdline_p, duration_msec, gr_name_r.c_str(), gr_passwd_r.c_str(), gr_gid, join(gr_mem_r.begin(), gr_mem_r.end(), "\", \"").c_str());

    size_t name_l    = gr_name_r.length()   + 1; /* +1 to include NUL terminating char */
    size_t passwd_l  = gr_passwd_r.length() + 1;
    size_t array_l   = sizeof(char *) * (gr_mem_r.size() + 1); /* +1 for a nullptr sentinel */
    size_t strings_l = std::accumulate(gr_mem_r.begin(), gr_mem_r.end(), 0, [](size_t sum, const std::string& elem) {return sum + elem.length() + 1;}); /* +1 to include NUL terminating char */
    if (buflen < (name_l + passwd_l + array_l + strings_l))
    {
        SYSLOG(LOG_DEBUG, "%s() - cmdline=\"%s\" not enough memory for struct group data\n", fn_p, cmdline_p);

        if (errnop) *errnop = ENOMEM;
        return NSS_STATUS_TRYAGAIN;
    }

    result->gr_gid    = gr_gid;
    result->gr_mem    = (char  **)buffer;
    result->gr_mem[0] = buffer + array_l;
    for (unsigned i = 0; i < gr_mem_r.size(); i++)
        result->gr_mem[i+1] = cpy2buf(result->gr_mem[i], gr_mem_r[i].c_str(), gr_mem_r[i].length() + 1);

    result->gr_name   = result->gr_mem[gr_mem_r.size()];
    result->gr_mem[gr_mem_r.size()] = nullptr; // sentinel

    result->gr_passwd = cpy2buf(result->gr_name, gr_name_r.c_str(), name_l);
    cpy2buf(result->gr_passwd, gr_passwd_r.c_str(), passwd_l);

    if (errnop) *errnop = 0;

    return NSS_STATUS_SUCCESS;
}



#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Retrieve passwd info from Host Account Management daemon (hamd).
 *
 * @param name      User name.
 * @param result    Where to write the result
 * @param buffer    Buffer used as a temporary pool where we can save
 *                  strings.
 * @param buflen    Size of memory pointed to by buffer
 * @param errnop    Where to return the errno
 *
 * @return NSS_STATUS_SUCCESS, NSS_STATUS_NOTFOUND, or NSS_STATUS_TRYAGAIN.
 */
enum nss_status _nss_ham_getpwnam_r(const char    * name,
                                    struct passwd * result,
                                    char          * buffer,
                                    size_t          buflen,
                                    int           * errnop)
{
    SYSLOG(LOG_DEBUG, "_nss_ham_getpwnam_r() - [%u:%u] cmdline=\"%s\" name=\"%s\"\n",
           getuid(), getgid(), cmdline_p, name);

    uint64_t  before_nsec = verbose ? get_now_nsec() : 0ULL;

    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection            conn = DBus::Connection::SystemBus();
    name_service_proxy_c        intf(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    ::DBus::Struct <
       bool,        /* _1: success   */
       std::string, /* _2: pw_name   */
       std::string, /* _3: pw_passwd */
       uint32_t,    /* _4: pw_uid    */
       uint32_t,    /* _5: pw_gid    */
       std::string, /* _6: pw_gecos  */
       std::string, /* _7: pw_dir    */
       std::string  /* _8: pw_shell  */
    >  ham_data;

    try
    {
        ham_data = intf.getpwnam(name);
    }
    catch (DBus::Error & ex)
    {
        SYSLOG(LOG_CRIT, "_nss_ham_getpwnam_r() - cmdline=\"%s\" Exception %s\n", cmdline_p, ex.what());
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    double  duration_msec = verbose ? (get_now_nsec() - before_nsec)/1000000.0 : 0.0;

    return pw_fill_result("_nss_ham_getpwnam_r",
                          duration_msec,
                          ham_data,
                          result,
                          buffer,
                          buflen,
                          errnop);
}

/**
 * @brief Retrieve passwd info from Host Account Management daemon (hamd).
 *
 * @param uid       User ID.
 * @param result    Where to write the result
 * @param buffer    Buffer used as a temporary pool where we can save
 *                  strings.
 * @param buflen    Size of memory pointed to by buffer
 * @param errnop    Where to return the errno
 *
 * @return NSS_STATUS_SUCCESS, NSS_STATUS_NOTFOUND, or NSS_STATUS_TRYAGAIN.
 */
enum nss_status _nss_ham_getpwuid_r(uid_t           uid,
                                    struct passwd * result,
                                    char          * buffer,
                                    size_t          buflen,
                                    int           * errnop)
{
    SYSLOG(LOG_DEBUG, "_nss_ham_getpwuid_r() - [%u:%u] cmdline=\"%s\" uid=%u\n",
           getuid(), getgid(), cmdline_p, uid);

    uint64_t  before_nsec = verbose ? get_now_nsec() : 0ULL;

    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection            conn = DBus::Connection::SystemBus();
    name_service_proxy_c        intf(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    ::DBus::Struct <
       bool,        /* _1: success   */
       std::string, /* _2: pw_name   */
       std::string, /* _3: pw_passwd */
       uint32_t,    /* _4: pw_uid    */
       uint32_t,    /* _5: pw_gid    */
       std::string, /* _6: pw_gecos  */
       std::string, /* _7: pw_dir    */
       std::string  /* _8: pw_shell  */
    >  ham_data;

    try
    {
        ham_data = intf.getpwuid(uid);
    }
    catch (DBus::Error & ex)
    {
        SYSLOG(LOG_CRIT, "_nss_ham_getpwuid_r() - cmdline=\"%s\" Exception %s\n", cmdline_p, ex.what());
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    double  duration_msec = verbose ? (get_now_nsec() - before_nsec)/1000000.0 : 0.0;

    return pw_fill_result("_nss_ham_getpwuid_r",
                          duration_msec,
                          ham_data,
                          result,
                          buffer,
                          buflen,
                          errnop);
}

/**
 * @brief Retrieve group info from Host Account Management daemon (hamd).
 *
 * @param name      Group name.
 * @param result    Where to write the result
 * @param buffer    Buffer used as a temporary pool where we can save
 *                  strings.
 * @param buflen    Size of memory pointed to by buffer
 * @param errnop    Where to return the errno
 *
 * @return NSS_STATUS_SUCCESS, NSS_STATUS_NOTFOUND, or NSS_STATUS_TRYAGAIN.
 */
enum nss_status _nss_ham_getgrnam_r(const char    * name,
                                    struct group  * result,
                                    char          * buffer,
                                    size_t          buflen,
                                    int           * errnop)
{
    SYSLOG(LOG_DEBUG, "_nss_ham_getgrnam_r() - [%u:%u] cmdline=\"%s\" name=\"%s\"\n",
           getuid(), getgid(), cmdline_p, name);

    uint64_t  before_nsec = verbose ? get_now_nsec() : 0ULL;

    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection            conn = DBus::Connection::SystemBus();
    name_service_proxy_c        intf(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    ::DBus::Struct <
       bool,                      /* _1: success   */
       std::string,               /* _2: gr_name   */
       std::string,               /* _3: gr_passwd */
       uint32_t,                  /* _4: gr_gid    */
       std::vector< std::string > /* _5: gr_mem    */
    >  ham_data;

    try
    {
        ham_data = intf.getgrnam(name);
    }
    catch (DBus::Error & ex)
    {
        SYSLOG(LOG_CRIT, "_nss_ham_getgrnam_r() - cmdline=\"%s\" Exception %s\n", cmdline_p, ex.what());
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    double  duration_msec = verbose ? (get_now_nsec() - before_nsec)/1000000.0 : 0.0;

    return gr_fill_result("_nss_ham_getgrnam_r",
                          duration_msec,
                          ham_data,
                          result,
                          buffer,
                          buflen,
                          errnop);
}

/**
 * @brief Retrieve group info from Host Account Management daemon (hamd).
 *
 * @param gid       Group ID.
 * @param result    Where to write the result
 * @param buffer    Buffer used as a temporary pool where we can save
 *                  strings.
 * @param buflen    Size of memory pointed to by buffer
 * @param errnop    Where to return the errno
 *
 * @return NSS_STATUS_SUCCESS, NSS_STATUS_NOTFOUND, or NSS_STATUS_TRYAGAIN.
 */
enum nss_status _nss_ham_getgrgid_r(gid_t           gid,
                                    struct group  * result,
                                    char          * buffer,
                                    size_t          buflen,
                                    int           * errnop)
{
    SYSLOG(LOG_DEBUG, "_nss_ham_getgrgid_r() - [%u:%u] cmdline=\"%s\" gid=%u\n",
           getuid(), getgid(), cmdline_p, gid);

    uint64_t  before_nsec = verbose ? get_now_nsec() : 0ULL;

    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection            conn = DBus::Connection::SystemBus();
    name_service_proxy_c        intf(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    ::DBus::Struct <
       bool,                      /* _1: success   */
       std::string,               /* _2: gr_name   */
       std::string,               /* _3: gr_passwd */
       uint32_t,                  /* _4: gr_gid    */
       std::vector< std::string > /* _5: gr_mem    */
    >  ham_data;

    try
    {
        ham_data = intf.getgrgid(gid);
    }
    catch (DBus::Error & ex)
    {
        SYSLOG(LOG_CRIT, "_nss_ham_getgrgid_r() - cmdline=\"%s\" Exception %s\n", cmdline_p, ex.what());
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    double  duration_msec = verbose ? (get_now_nsec() - before_nsec)/1000000.0 : 0.0;

    return gr_fill_result("_nss_ham_getgrgid_r",
                          duration_msec,
                          ham_data,
                          result,
                          buffer,
                          buflen,
                          errnop);
}

/**
 * @brief Invoke Host Account Management Daemon (hamd) over DBus to
 *        retrieve shadow password information for user @name. Upon receipt
 *        of hamd data, fill structure pointed to by @result.
 *
 * @details The string fields pointed to by the members of the spwd
 *          structure are stored in @buffer of size buflen. In case @buffer
 *          has insufficient memory to hold the strings of struct spwd,
 *          the ENOMEM error will be reported.
 *
 * @param name    User name
 * @param result  Pointer to destination where data gets copied
 * @param buffer  Pointer to memory where strings can be stored
 * @param buflen  Size of buffer
 * @param errnop  Pointer to where errno code can be written
 *
 * @return - If no entry was found, return NSS_STATUS_NOTFOUND and set
 *           errno=ENOENT.
 *         - If @buffer has insufficient memory to hold the strings of
 *           struct passwd then return NSS_STATUS_TRYAGAIN and set
 *           errno=ENOMEM.
 *         - Otherwise return NSS_STATUS_SUCCESS and set errno to 0.
 */
enum nss_status _nss_ham_getspnam_r(const char    * name,
                                    struct spwd   * result,
                                    char          * buffer,
                                    size_t          buflen,
                                    int           * errnop)
{
    SYSLOG(LOG_DEBUG, "_nss_ham_getspnam_r() - [%u:%u] cmdline=\"%s\" name=\"%s\"\n",
           getuid(), getgid(), cmdline_p, name);

    uint64_t  before_nsec = verbose ? get_now_nsec() : 0ULL;

    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection            conn = DBus::Connection::SystemBus();
    name_service_proxy_c        intf(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    ::DBus::Struct <
        bool,        /* _1:  success   */
        std::string, /* _2:  sp_namp   */
        std::string, /* _3:  sp_pwdp   */
        int32_t,     /* _4:  sp_lstchg */
        int32_t,     /* _5:  sp_min    */
        int32_t,     /* _6:  sp_max    */
        int32_t,     /* _7:  sp_warn   */
        int32_t,     /* _8:  sp_inact  */
        int32_t,     /* _9:  sp_expire */
        uint32_t     /* _10: sp_flag   */
    > ham_data;

    try
    {
       ham_data = intf.getspnam(name);
    }
    catch (DBus::Error & ex)
    {
        SYSLOG(LOG_CRIT, "_nss_ham_getspnam_r() - cmdline=\"%s\" Exception %s\n", cmdline_p, ex.what());
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    double  duration_msec = verbose ? (get_now_nsec() - before_nsec)/1000000.0 : 0.0;
    bool    success       = ham_data._1;

    if (!success)
    {
        SYSLOG(LOG_DEBUG, "_nss_ham_getspnam_r() - cmdline=\"%s\" exec time=%.3f ms. Failed!\n", cmdline_p, duration_msec);
        *errnop = ENOENT;
        return NSS_STATUS_NOTFOUND;
    }

    std::string   & sp_namp_r = ham_data._2;
    std::string   & sp_pwdp_r = ham_data._3;
    long            sp_lstchg = ham_data._4;
    long            sp_min    = ham_data._5;
    long            sp_max    = ham_data._6;
    long            sp_warn   = ham_data._7;
    long            sp_inact  = ham_data._8;
    long            sp_expire = ham_data._9;
    unsigned long   sp_flag   = ham_data._10;

    SYSLOG(LOG_DEBUG, "_nss_ham_getspnam_r() - cmdline=\"%s\" exec time=%.3f ms. Success! sp_namp=\"%s\", sp_pwdp=\"%s\", sp_lstchg=%ld, sp_min=%ld, sp_max=%ld, sp_warn=%ld, sp_inact=%ld, sp_expire=%ld, sp_flag=%lu\n",
           cmdline_p, duration_msec, sp_namp_r.c_str(), sp_pwdp_r.c_str(), sp_lstchg, sp_min, sp_max, sp_warn, sp_inact, sp_expire, sp_flag);

    size_t sp_namp_l = sp_namp_r.length() + 1; /* +1 to include NUL terminating char */
    size_t sp_pwdp_l = sp_pwdp_r.length() + 1;
    if (buflen < (sp_namp_l + sp_pwdp_l))
    {
        SYSLOG(LOG_DEBUG, "_nss_ham_getspnam_r() - cmdline=\"%s\" not enough memory for struct spwd data\n", cmdline_p);

        if (errnop) *errnop = ENOMEM;
        return NSS_STATUS_TRYAGAIN;
    }

    result->sp_namp = buffer;
    result->sp_pwdp = cpy2buf(result->sp_namp, sp_namp_r.c_str(), sp_namp_l);
    cpy2buf(result->sp_pwdp, sp_pwdp_r.c_str(), sp_pwdp_l);
    result->sp_lstchg = sp_lstchg;
    result->sp_min    = sp_min;
    result->sp_max    = sp_max;
    result->sp_warn   = sp_warn;
    result->sp_inact  = sp_inact;
    result->sp_expire = sp_expire;
    result->sp_flag   = sp_flag;

    if (errnop) *errnop = 0;

    return NSS_STATUS_SUCCESS;
}

#ifdef __cplusplus
}
#endif
