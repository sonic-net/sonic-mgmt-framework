// Host Account Management
#include <glib.h>               // g_option_context_new(), g_file_test(), etc...
#include <stdlib.h>             // strtoll(), EXIT_SUCCESS
#include <syslog.h>             // syslog()
#include <limits.h>             // LINE_MAX, LLONG_MIN, LLONG_MAX
#include <errno.h>              // errno, EINVAL, ERANGE

#include "hamd.h"               // hamd_config_c
#include "../shared/utils.h"    // true_false()


static long long numberize(const char  * str_p, long long minval, long long maxval, const char ** errstr_pp = nullptr);
static std::string expand_certgen_cmd(const std::string & certgen_cmd_r, const std::string  & user_r, const std::string  & certdir_r);

//******************************************************************************
hamd_config_c::hamd_config_c(int argc, char **argv)
{
    GOptionContext  * ctx_p;
    std::string       verbose_help   = "Print extra debug        [" + std::string(true_false(tron_default_m)) + ']';
    std::string       conf_file_help = "Configuration file       [" + std::string(conf_file_default_pm) + ']';

    static const GOptionEntry options[] =
    {
        { "verbose",   'v',  G_OPTION_FLAG_NONE, G_OPTION_ARG_NONE,   &tron_m,       verbose_help.c_str(),   nullptr },
        { "conf-file", 'f',  G_OPTION_FLAG_NONE, G_OPTION_ARG_STRING, &conf_file_pm, conf_file_help.c_str(), nullptr },
        { nullptr,     '\0', G_OPTION_FLAG_NONE, G_OPTION_ARG_NONE,   nullptr,       nullptr,                nullptr }
    };

    const std::string description =
        "Configuration file parameters:\n"
        "  debug=[yes/no]      Enable additional debug info to the syslog                       [" + std::string(true_false(tron_default_m, "yes", "no")) + "]\n" +
        "  poll_period=[sec]   Daemon's polling period. Used for periodic house keeping tasks   [" + std::to_string(poll_period_sec_default_m) + "s]\n" +
        "  uid_min=[uint32]    System-assigned credentials minimum UID. Should be >= 1000       [" + std::to_string(sac_uid_min_default_m) + "]\n" +
        "  uid_max=[uint32]    System-assigned credentials maximum UID. Should be > uid_min     [" + std::to_string(sac_uid_max_default_m) + "]\n" +
        "  certgen=[command]   Certificate generation command. Supp. vars: $CERTDIR, $USERNAME  [" + certgen_cmd_default_m + "]\n" +
        "  shell=[path]        Shell to be assigned to new users                                [" + shell_default_m + "]";

    ctx_p = g_option_context_new(nullptr);
    g_option_context_set_summary(ctx_p, "Host Account Management Daemon (hamd)");
    g_option_context_set_description(ctx_p, description.c_str());
    g_option_context_add_main_entries (ctx_p, &options[0], nullptr);
    g_option_context_parse (ctx_p, &argc, &argv, nullptr);
    g_option_context_free (ctx_p);

    reload();
}

//******************************************************************************
void hamd_config_c::reload()
{
    FILE * file = fopen(conf_file_pm, "re");
    if (file)
    {
        gint poll_period_sec       = poll_period_sec_default_m;
        gint sac_uid_min           = sac_uid_min_default_m;
        gint sac_uid_max           = sac_uid_max_default_m;
        bool tron                  = tron_default_m;
        std::string certgen_cmd    = certgen_cmd_default_m;
        std::string shell          = shell_default_m;

        #define WHITESPACE " \t\n\r"
        char    line[LINE_MAX];
        char  * p;
        char  * s;
        while (nullptr != (p = fgets(line, sizeof line, file)))
        {
            p += strspn(p, WHITESPACE);            // Remove leading newline and spaces
            if (*p == '#' || *p == '\0') continue; // Skip comments and empty lines
            p[strcspn(p, "\n\r")] = '\0';          // Remove trailing newline chars

            if (nullptr != (s = startswith(p, "debug")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                tron = strneq(s, "yes", 3);
            }
            else if (nullptr != (s = startswith(p, "poll_period")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                const char * errstr_p = nullptr;
                poll_period_sec = (gint)numberize(s, 0, G_MAXINT, &errstr_p);
                if (errstr_p != nullptr)
                {
                    syslog(LOG_ERR, "Error reading %s: poll_period %s (ignored)", conf_file_pm, errstr_p);
                }
            }
            else if (nullptr != (s = startswith(p, "uid_min")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                const char * errstr_p = nullptr;
                sac_uid_min = (gint)numberize(s, 1000, G_MAXUINT, &errstr_p);
                if (errstr_p != nullptr)
                {
                    syslog(LOG_ERR, "Error reading %s: uid_min %s (ignored)", conf_file_pm, errstr_p);
                }
            }
            else if (nullptr != (s = startswith(p, "uid_max")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                const char * errstr_p = nullptr;
                sac_uid_max = (gint)numberize(s, 1000, G_MAXUINT, &errstr_p);
                if (errstr_p != nullptr)
                {
                    syslog(LOG_ERR, "Error reading %s: uid_max %s (ignored)", conf_file_pm, errstr_p);
                }
            }
            else if (nullptr != (s = startswith(p, "certgen")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                certgen_cmd = s;
            }
            else if (nullptr != (s = startswith(p, "shell")))
            {
                s += strspn(s, " \t=");            // Skip leading spaces and equal sign (=)
                shell = s;
            }
        }

        fclose(file);

        tron_m            = tron;
        poll_period_sec_m = poll_period_sec;

        if (sac_uid_min > sac_uid_max)
        {
            syslog(LOG_ERR, "Error reading %s: uid_max is less than uid_min", conf_file_pm);
        }
        else if ((1 + (sac_uid_max - sac_uid_min)) < 200)
        {
            syslog(LOG_ERR, "Error reading %s: uid_min..uid_max range too small (should be >= 200).", conf_file_pm);
        }
        else
        {
            sac_uid_min_m   = sac_uid_min;
            sac_uid_max_m   = sac_uid_max;
            sac_uid_range_m = 1 + (sac_uid_max_m - sac_uid_min_m);
        }

        if (certgen_cmd_m != certgen_cmd)
        {
#if (0)
            // Make sure that the program exists and can be executed
            std::string cmd = expand_certgen_cmd(certgen_cmd, "____bozo____", "/tmp");

            int  rc          = system(cmd.c_str());
            bool term_normal = WIFEXITED(rc);
            if (term_normal)
            {
#endif
                certgen_cmd_m = certgen_cmd;
#if (0)
            }
            else
            {
                syslog(LOG_ERR, "Error reading %s: certgen=%s. Invalid command.", conf_file_pm, certgen_cmd.c_str());
            }
#endif
        }

        if (shell_m != shell)
        {
            // Make sure that the shell exists
            if (g_file_test(shell.c_str(), G_FILE_TEST_EXISTS))
            {
                shell_m = shell;
            }
            else
            {
                syslog(LOG_ERR, "Error reading %s: shell=%s. File not found.", conf_file_pm, shell.c_str());
            }
        }
    }

    LOG_CONDITIONAL(tron_m, LOG_DEBUG,
                    "hamd_config_c::reload() - conf_file_pm=\"%s\"\n"
                    "  tron_m            = %s\n"
                    "  poll_period_sec_m = %ds\n"
                    "  sac_uid_min_m     = %d\n"
                    "  sac_uid_max_m     = %d\n"
                    "  sac_uid_range_m   = %d\n",
                    conf_file_pm, true_false(tron_m), poll_period_sec_m,
                    sac_uid_min_m, sac_uid_max_m, sac_uid_range_m);
}

//******************************************************************************
std::string hamd_config_c::certgen_cmd(const std::string & user_r,
                                       const std::string & certdir_r) const
{
    return expand_certgen_cmd(certgen_cmd_m, user_r, certdir_r);
}

//******************************************************************************
std::string hamd_config_c::to_string() const
{
    std::ostringstream  oss;

    oss << "Running config:\n"
        << "  conf_file_pm              = " << conf_file_pm << '\n'
        << "  certgen_cmd_m             = " << certgen_cmd_m << '\n'
        << "  poll_period_sec_m         = " << std::to_string(poll_period_sec_m)  << "s\n"
        << "  sac_uid_min_m             = " << std::to_string(sac_uid_min_m) << '\n'
        << "  sac_uid_max_m             = " << std::to_string(sac_uid_max_m) << '\n'
        << "  sac_uid_range_m           = " << std::to_string(sac_uid_range_m)  << '\n'
        << "  shell_m                   = " << shell_m << '\n'
        << "  tron_m                    = " << true_false(tron_m) << '\n'
        << '\n'
        << "Default config:\n"
        << "  conf_file_default_pm      = " << conf_file_default_pm << '\n'
        << "  certgen_cmd_default_m     = " << certgen_cmd_default_m << '\n'
        << "  poll_period_sec_default_m = " << std::to_string(poll_period_sec_default_m)  << "s\n"
        << "  sac_uid_min_default_m     = " << std::to_string(sac_uid_min_default_m) << '\n'
        << "  sac_uid_max_default_m     = " << std::to_string(sac_uid_max_default_m) << '\n'
        << "  shell_default_m           = " << shell_default_m << '\n'
        << "  tron_default_m            = " << (tron_default_m ? "true" : "false");

    return oss.str();
}

//******************************************************************************
std::ostream & operator<<(std::ostream  & stream_r, const hamd_config_c  & obj_r)
{
    stream_r << obj_r.to_string();
    return stream_r;
}

//******************************************************************************
static inline char * _startswith(const char *s, const char *prefix_p, size_t prefix_l)
{
    if (strneq(s, prefix_p, prefix_l)) return (char *)s + prefix_l - 1;
    return nullptr;
}

//******************************************************************************
/**
 * @brief Replace variable fields of the certificate generation command
 *        (#certgen_cmd_r) with the values provided (#user_r, #certdir_r)
 *
 *        There are two variables recognized:
 *          $CERTDIR  - User's home directory
 *          $USERNAME - User's name
 *
 * @param certgen_cmd_r: Certificate generation command that may contain
 *                       $USERNAME and/or $CERTDIR to be replaced by user_r
 *                       and/or certdir_r respectively.
 * @param user_r:        User's name
 * @param certdir_r:     Location where to put certificates
 *
 * @return This method will return #certgen_cmd_r with $CERTDIR replaced by
 *         certdir_r and $USERNAME by user_r.
 */
static std::string expand_certgen_cmd(const std::string  & certgen_cmd_r,
                                      const std::string  & user_r,
                                      const std::string  & certdir_r)
{
    struct
    {
        const char          * pattern_p;
        size_t                pattern_l;
        const std::string   & replacement_r;
    } patterns[] =
    {
        { "$USERNAME",  sizeof("$USERNAME")-1, user_r    }, // Replace $USERNAME by user_r
        { "$CERTDIR",   sizeof("$CERTDIR")-1,  certdir_r }, // Replace $CERTDIR by certdir_r
        { nullptr ,     0,                     ""        }, // sentinel
    };

    std::string  output = "";

    const char * certgen_cmd_p = certgen_cmd_r.c_str();
    for (; certgen_cmd_p[0] != '\0'; certgen_cmd_p++)
    {
        if (certgen_cmd_p[0] == '$')
        {
            if (certgen_cmd_p[1] == '$')
            {
                output += '$';
                certgen_cmd_p++;
            }
            else
            {
                for (unsigned i = 0; patterns[i].pattern_p != nullptr; i++)
                {
                    const char * s = _startswith(certgen_cmd_p, patterns[i].pattern_p, patterns[i].pattern_l);
                    if (s != nullptr)
                    {
                        output += patterns[i].replacement_r;
                        certgen_cmd_p = s;
                        break;
                    }
                }
            }

            continue;
        }

        output += certgen_cmd_p[0];
    }

    return output;
}

/**
 * @brief Convert a "string" to an integer value. Handles overflow and/or
 *        underflow.
 *
 * @param str_p   The string to convert
 * @param minval  Minimum acceptable value
 * @param maxval  Maximum acceptable value
 * @param err_p   A place where to return an error string indicating why
 *                the function failed.
 *
 * @return str_p converted to a long long. On failure 0 is returned.
 */
static long long numberize(const char  * str_p,
                           long long     minval,
                           long long     maxval,
                           const char ** errstr_pp)
{
    #define OK       0
    #define INVALID  1
    #define TOOSMALL 2
    #define TOOLARGE 3

    struct
    {
        const char * str;
        int          err;
    } table[] =
    {
        { nullptr,        errno  }, // preserve current errno
        { "invalid",   EINVAL },
        { "too small", ERANGE },
        { "too large", ERANGE }
    };

    long long    number = 0;
    unsigned     result = OK;
    if (minval > maxval)
    {
        result = INVALID;
    }
    else
    {
        char  * ep;
        errno = 0;
        number = strtoll(str_p, &ep, 10);
        if (str_p == ep || *ep != '\0')
            result = INVALID;
        else if ((number == LLONG_MIN && errno == ERANGE) || number < minval)
            result = TOOSMALL;
        else if ((number == LLONG_MAX && errno == ERANGE) || number > maxval)
            result = TOOLARGE;
    }

    if (errstr_pp != nullptr) *errstr_pp = table[result].str;
    errno = table[result].err;

    return result != OK ? 0 : number;
}



