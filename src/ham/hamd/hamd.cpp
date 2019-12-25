// Host Account Management
#include "hamd.h"               // hamd_c
#include "../shared/utils.h"    // startswith(), streq()
#include "siphash24.h"          // siphash24()
#include "subprocess.h"         // run()
                                //
#include <glib.h>               // g_file_test()
#include <glib/gstdio.h>        // g_chdir()
#include <stdio.h>
#include <sys/types.h>          // getpwnam(), getpid()
#include <pwd.h>                // fgetpwent()
#include <string>               // std::string
#include <sstream>              // std::ostringstream
#include <algorithm>            // std::find
#include <syslog.h>             // syslog()
#include <pwd.h>                // getpwnam(), getpwuid()
#include <grp.h>                // getgrnam(), getgrgid()
#include <shadow.h>             // getspnam()
#include <unistd.h>             // getpid()

#include <features.h>           // __GNUC_PREREQ()
#if __GNUC_PREREQ(8,0) // If GCC >= 8.0
#   include <filesystem>
    typedef std::filesystem::path                 path_t;
#else
#   include <experimental/filesystem>
    typedef std::experimental::filesystem::path   path_t;
#endif // __GNUC_PREREQ(8,0)

int change_credentials(uid_t uid, gid_t gid)
{
    int rv;
    rv = setegid(gid);
    if (rv == -1)
    {
        syslog(LOG_WARNING, "change_credentials() - Error! setegid() failed");
    }
    else
    {
        rv = seteuid(uid);
        if (rv == -1)
        {
            syslog(LOG_WARNING, "change_credentials() - Error! seteuid() failed");
        }
    }

    return rv;
}

/**
 * @brief DBus adaptor class constructor
 *
 * @param config_r Structure containing configuration parameters
 * @param conn_r
 */
hamd_c::hamd_c(hamd_config_c & config_r, DBus::Connection & conn_r) :
    DBus::ObjectAdaptor(conn_r, DBUS_OBJ_PATH_BASE),
    config_rm(config_r),
    poll_timer_m((double)config_rm.poll_period_sec_m, hamd_c::on_poll_timeout, this)
{
    apply_config();
}

/**
 * @brief This is called when the poll_timer_m expires.
 *
 * @param user_data_p Pointer to user data. In this case this point to the
 *                    hamd_c object.
 * @return bool
 */
bool hamd_c::on_poll_timeout(gpointer user_data_p)
{
    hamd_c * p = static_cast<hamd_c *>(user_data_p);
    LOG_CONDITIONAL(p->is_tron(), LOG_INFO, "hamd_c::on_poll_timeout()");
    p->rm_unconfirmed_users();
    return true; // Return true to repeat timer
}

/**
 * @brief reload configuration and apply to running daemon.
 */
void hamd_c::reload()
{
    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::reload()");
    config_rm.reload();
    apply_config();
}

/**
 * @brief Apply the configuration to the running daemon
 */
void hamd_c::apply_config()
{
    if (config_rm.poll_period_sec_m > 0)
        poll_timer_m.start((double)config_rm.poll_period_sec_m);
    else
        poll_timer_m.stop();
}

/**
 * @brief This is called just before the destructor is called and is used
 *        to clean up all resources in use by the class instance.
 */
void hamd_c::cleanup()
{
    poll_timer_m.stop();
}

#if (0)
/**
 * @brief Scan "/etc/passwd" looking for user. If found, return a pointer
 *        to a "struct passwd" containing all the data related to user.
 *
 * @param fn E.g. /etc/passwd
 * @param user The user we're looking for
 *
 * @return If user found, return a pointer to a struct passwd.
 */
static struct passwd * fgetpwname(const char * login, const char * fname_p="/etc/passwd");
static struct passwd * fgetpwname(const char * login, const char * fname_p)
{
    struct passwd * pwd = NULL;
    FILE          * f   = fopen(fname_p, "re");
    if (f)
    {
        struct passwd * ent;
        while (NULL != (ent = fgetpwent(f)))
        {
            if (streq(ent->pw_name, login))
            {
                pwd = ent;
                break;
            }
        }
        fclose(f);
    }

    return pwd;
}
#endif

std::string hamd_c::certgen(const std::string  & login) const
{
    std::string cmd = config_rm.certgen_m + ' ' + login;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::certgen() - Generate user \"%s\" certificates [%s]", login.c_str(), cmd.c_str());

    int rc;
    std::string std_out;
    std::string std_err;
    std::tie(rc, std_out, std_err) = run(cmd);

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::certgen() - Generate user \"%s\" certificates rc=%d, stdout=%s, stderr=%s",
                    login.c_str(), rc, std_out.c_str(), std_err.c_str());

    if (rc != 0)
        return "Failed to generate certificates for " + login + ". " + std_err;

    return "";
}

static std::string roles_as_string(const std::vector< std::string > & roles)
{
    std::vector< std::string > new_roles = roles;
    if (std::find(new_roles.cbegin(), new_roles.cend(), "admin") != new_roles.cend())
    {
        if (std::find(new_roles.cbegin(), new_roles.cend(), "sudo") == new_roles.cend())
            new_roles.push_back("sudo");

        if (std::find(new_roles.cbegin(), new_roles.cend(), "docker") == new_roles.cend())
            new_roles.push_back("docker");
    }

    if (std::find(new_roles.cbegin(), new_roles.cend(), "operator") != new_roles.cend())
    {
        if (std::find(new_roles.cbegin(), new_roles.cend(), "docker") == new_roles.cend())
            new_roles.push_back("docker");
    }

    return join(new_roles.cbegin(), new_roles.cend(), ",");
}


/**
 * @brief Create a new user
 */
::DBus::Struct< bool, std::string > hamd_c::useradd(const std::string                & login,
                                                    const std::vector< std::string > & roles,
                                                    const std::string                & hashed_pw)
{
    std::string errmsg  = "";
    bool        success = false;

    ::DBus::Struct< bool,       /* success */
                    std::string /* errmsg  */ > ret;

    ret._1 = true; // Let's be optimistic
    ret._2 = "";   // ..and set returned value to success.

    std::string cmd = "/usr/sbin/useradd"
                      " --create-home"
                      " --password '" + hashed_pw + "'";

    const std::string & shell_r = config_rm.shell();
    if (!shell_r.empty())
        cmd += " --shell " + shell_r;

    std::string roles_str = roles_as_string(roles);
    if (!roles_str.empty())
        cmd += " --groups " + roles_str;

    if (!roles.empty())
        cmd += " --gid " + roles[0];

    cmd += ' ' + login;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::useradd() - Create user \"%s\" [%s]", login.c_str(), cmd.c_str());

    int rc;
    std::string std_out;
    std::string std_err;
    std::tie(rc, std_out, std_err) = run(cmd);

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::useradd() - Create user \"%s\" rc=%d, stdout=%s, stderr=%s",
                    login.c_str(), rc, std_out.c_str(), std_err.c_str());

    if (rc == 0)
    {
        errmsg  = certgen(login);
        success = errmsg.empty(); // The errmsg should be empty on success

        if (!success)
        {
            // Since we failed to generate the certificates,
            // we now need to delete the user.
            cmd = "/usr/sbin/userdel --force --remove " + login;

            LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::useradd() - executing command \"%s\"", cmd.c_str());

            std::tie(rc, std_out, std_err) = run(cmd);

            LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::useradd() - command returned rc=%d, stdout=%s, stderr=%s",
                            rc, std_out.c_str(), std_err.c_str());
        }
    }
    else
    {
        success = false;
        errmsg  = std_err;
    }

    ret._1 = success;
    ret._2 = errmsg;

    return ret;
}

/**
 * @brief Delete a user account
 */
::DBus::Struct< bool, std::string > hamd_c::userdel(const std::string& login)
{
    ::DBus::Struct< bool,       /* success */
                    std::string /* errmsg  */ > ret;

    struct passwd * pwd = ::getpwnam(login.c_str());
    if (pwd == nullptr)
    {
        // User doesn't exist...so success!
        ret._1 = true;
        ret._2 = "";
    }
    else
    {
        std::string  cmd = "/usr/sbin/userdel --force --remove " + login;

        LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::userdel() - executing command \"%s\"", cmd.c_str());

        int rc;
        std::string std_out;
        std::string std_err;
        std::tie(rc, std_out, std_err) = run(cmd);

        LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::userdel() - command returned rc=%d, stdout=%s, stderr=%s",
                        rc, std_out.c_str(), std_err.c_str());

        ret._1 = rc == 0;
        ret._2 = rc == 0 ? "" : std_err;
    }

    return ret;
}

/**
 * @brief Change user password
 */
::DBus::Struct< bool, std::string > hamd_c::passwd(const std::string& login, const std::string& hashed_pw)
{
    std::string  cmd = "/usr/sbin/usermod --password " + hashed_pw + ' ' + login;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::passwd() - executing command \"%s\"", cmd.c_str());

    int rc;
    std::string std_out;
    std::string std_err;
    std::tie(rc, std_out, std_err) = run(cmd);

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::passwd() - command returned rc=%d, stdout=%s, stderr=%s",
                    rc, std_out.c_str(), std_err.c_str());

    ::DBus::Struct< bool,       /* success */
                    std::string /* errmsg  */ > ret;

    ret._1 = rc == 0;
    ret._2 = rc == 0 ? "" : std_err;

    return ret;
}

/**
 * @brief Set user's roles (supplementary groups)
 */
::DBus::Struct< bool, std::string > hamd_c::set_roles(const std::string& login, const std::vector< std::string >& roles)
{
    std::string roles_str = roles_as_string(roles);
    std::string cmd;

    if (!roles_str.empty())
        cmd = "/usr/sbin/usermod --groups " + roles_str + ' ' + login;
    else
        cmd = "/usr/sbin/usermod --groups \"\" " + login;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::set_roles() - executing command \"%s\"", cmd.c_str());

    int rc;
    std::string std_out;
    std::string std_err;
    std::tie(rc, std_out, std_err) = run(cmd);

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::set_roles() - command returned rc=%d, stdout=%s, stderr=%s",
                    rc, std_out.c_str(), std_err.c_str());

    ::DBus::Struct< bool,       /* success */
                    std::string /* errmsg  */ > ret;

    ret._1 = rc == 0;
    ret._2 = rc == 0 ? "" : std_err;

    return ret;
}

/**
 * @brief Create a group
 */
::DBus::Struct< bool, std::string > hamd_c::groupadd(const std::string& group)
{
    ::DBus::Struct< bool, std::string > ret;
    ret._1 = false;
    ret._2 = "Not implemented";
    return ret;
}

/**
 * @brief Delete a group
 */
::DBus::Struct< bool, std::string > hamd_c::groupdel(const std::string& group)
{
    ::DBus::Struct< bool, std::string > ret;
    ret._1 = false;
    ret._2 = "Not implemented";
    return ret;
}

::DBus::Struct< bool, std::string, std::string, uint32_t, uint32_t, std::string, std::string, std::string > hamd_c::getpwnam(const std::string& name)
{
    ::DBus::Struct< bool,         /* success   */
                    std::string,  /* pw_name   */
                    std::string,  /* pw_passwd */
                    uint32_t,     /* pw_uid    */
                    uint32_t,     /* pw_gid    */
                    std::string,  /* pw_gecos  */
                    std::string,  /* pw_dir    */
                    std::string > /* pw_shell  */ ret;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::getpwnam(%s)", name.c_str());

    struct passwd * p = ::getpwnam(name.c_str());

    ret._1 = p != NULL;
    if (ret._1) // success?
    {
        ret._2 = p->pw_name;
        ret._3 = p->pw_passwd;
        ret._4 = p->pw_uid;
        ret._5 = p->pw_gid;
        ret._6 = p->pw_gecos;
        ret._7 = p->pw_dir;
        ret._8 = p->pw_shell;
    }

    return ret;
}

::DBus::Struct< bool, std::string, std::string, uint32_t, uint32_t, std::string, std::string, std::string > hamd_c::getpwuid(const uint32_t& uid)
{
    ::DBus::Struct< bool,         /* success   */
                    std::string,  /* pw_name   */
                    std::string,  /* pw_passwd */
                    uint32_t,     /* pw_uid    */
                    uint32_t,     /* pw_gid    */
                    std::string,  /* pw_gecos  */
                    std::string,  /* pw_dir    */
                    std::string > /* pw_shell  */ ret;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::getpwuid(%u)", uid);

    struct passwd * p = ::getpwuid(uid);

    ret._1 = p != NULL;
    if (ret._1) // success?
    {
        ret._2 = p->pw_name;
        ret._3 = p->pw_passwd;
        ret._4 = p->pw_uid;
        ret._5 = p->pw_gid;
        ret._6 = p->pw_gecos;
        ret._7 = p->pw_dir;
        ret._8 = p->pw_shell;
    }

    return ret;
}

::DBus::Struct< bool, std::string, std::string, uint32_t, std::vector< std::string > > hamd_c::getgrnam(const std::string& name)
{
    ::DBus::Struct< bool,                        /* success   */
                    std::string,                 /* gr_name   */
                    std::string,                 /* gr_passwd */
                    uint32_t,                    /* gr_gid    */
                    std::vector< std::string > > /* gr_mem    */ ret;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::getgrnam(%s)", name.c_str());

    struct group * p = ::getgrnam(name.c_str());

    ret._1 = p != NULL;
    if (ret._1) // success?
    {
        ret._2 = p->gr_name;
        ret._3 = p->gr_passwd;
        ret._4 = p->gr_gid;

        for (unsigned i = 0; p->gr_mem[i] != NULL; i++)
            ret._5.push_back(p->gr_mem[i]);
    }

    return ret;
}

::DBus::Struct< bool, std::string, std::string, uint32_t, std::vector< std::string > > hamd_c::getgrgid(const uint32_t& gid)
{
    ::DBus::Struct< bool,                        /* success   */
                    std::string,                 /* gr_name   */
                    std::string,                 /* gr_passwd */
                    uint32_t,                    /* gr_gid    */
                    std::vector< std::string > > /* gr_mem    */ ret;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::getgrgid(%u)", gid);

    struct group * p = ::getgrgid(gid);

    ret._1 = p != NULL;
    if (ret._1) // success?
    {
        ret._2 = p->gr_name;
        ret._3 = p->gr_passwd;
        ret._4 = p->gr_gid;

        for (unsigned i = 0; p->gr_mem[i] != NULL; i++)
            ret._5.push_back(p->gr_mem[i]);
    }

    return ret;
}

::DBus::Struct< bool, std::string, std::string, int32_t, int32_t, int32_t, int32_t, int32_t, int32_t, uint32_t > hamd_c::getspnam(const std::string& name)
{
    ::DBus::Struct< bool,        /* success   */
                    std::string, /* sp_namp   */
                    std::string, /* sp_pwdp   */
                    int32_t,     /* sp_lstchg */
                    int32_t,     /* sp_min    */
                    int32_t,     /* sp_max    */
                    int32_t,     /* sp_warn   */
                    int32_t,     /* sp_inact  */
                    int32_t,     /* sp_expire */
                    uint32_t >   /* sp_flag   */ ret;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::getspnam(%s)", name.c_str());

    struct spwd * p = ::getspnam(name.c_str());

    ret._1 = p != NULL;
    if (ret._1) // success?
    {
        ret._2  = p->sp_namp;
        ret._3  = p->sp_pwdp;
        ret._4  = p->sp_lstchg;
        ret._5  = p->sp_min;
        ret._6  = p->sp_max;
        ret._7  = p->sp_warn;
        ret._8  = p->sp_inact;
        ret._9  = p->sp_expire;
        ret._10 = p->sp_flag;
    }

    return ret;
}

/**
 * @brief Remove unconfirmed users from /etc/passwd. Unconfirmed users have
 *        the string "Unconfirmed sac user [PID]" in their GECOS string and
 *        the PID does not exist anymore.
 */
void hamd_c::rm_unconfirmed_users() const
{
    FILE  * f = fopen("/etc/passwd", "re");
    if (f)
    {
        struct passwd * ent;
        std::string     base_cmd("/usr/sbin/userdel --remove ");
        std::string     full_cmd;
        g_chdir("/proc");
        while (NULL != (ent = fgetpwent(f)))
        {
            const char * pid_p;
            if ((ent->pw_uid >= (uid_t)config_rm.sac_uid_min_m) && (ent->pw_uid <= (uid_t)config_rm.sac_uid_max_m) &&
                (NULL != (pid_p = startswith(ent->pw_gecos, "Unconfirmed system-assigned credentials "))))
            {
                if (!g_file_test(pid_p, G_FILE_TEST_EXISTS))
                {
                    // Directory does not exist, which means process does not
                    // exist either. Let's remove this user which was never
                    // confirmed by PAM authentification.
                    full_cmd = base_cmd + ent->pw_name;

                    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::rm_unconfirmed_users() - executing command \"%s\"", full_cmd.c_str());

                    int rc;
                    std::string std_out;
                    std::string std_err;
                    std::tie(rc, std_out, std_err) = run(full_cmd);

                    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "hamd_c::rm_unconfirmed_users() - command returned rc=%d, stdout=%s, stderr=%s",
                                    rc, std_out.c_str(), std_err.c_str());

                    if (rc != 0)
                    {
                        syslog(LOG_ERR, "User \"%s\": Failed to removed unconfirmed user UID=%d. %s",
                               ent->pw_name, ent->pw_uid, std_err.c_str());
                    }
                }
            }
        }
        fclose(f);
    }
}

/**
 * @brief This is a DBus interface used by remote programs to add an
 *        unconfirmed user.
 *
 * @param username  Username to be added
 * @param pid       PID of the caller.
 *
 * @return bool     true if user was added successfully,
 *                  false otherwise.
 */
bool hamd_c::add_unconfirmed_user(const std::string& username, const uint32_t& pid)
{
    // First, let's check if there are any
    // unconfirmed users that could be removed.
    rm_unconfirmed_users();

    // Next, add <username> as an unconfirmed user.
    static const uint8_t hash_key[] =
    {
        0x37, 0x53, 0x7e, 0x31, 0xcf, 0xce, 0x48, 0xf5,
        0x8a, 0xbb, 0x39, 0x57, 0x8d, 0xd9, 0xec, 0x59
    };

    unsigned     n_tries;
    uid_t        candidate;
    std::string  name(username);
    std::string  full_cmd;
    std::string  base_cmd = "/usr/sbin/useradd"
                            " --create-home"
                            " --user-group"
                            " --comment \"Unconfirmed system-assigned credentials " + std::to_string(pid) + '"';

    const std::string & shell_r = config_rm.shell();
    if (!shell_r.empty())
        base_cmd += " --shell " + shell_r;

    for (n_tries = 0; n_tries < 100; n_tries++) /* Give up retrying eventually */
    {
        // Find a unique UID in the range sac_uid_min_m..sac_uid_max_m.
        // We use a hash function to always get the same ID for a given user
        // name. Hash collisions (i.e. two user names with the same hash) will
        // be handled by trying with a slightly different username.
        candidate = config_rm.uid_fit_into_range(siphash24(name.c_str(), name.length(), hash_key));

        LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": attempt %d using name \"%s\", candidate UID=%lu",
                        username.c_str(), n_tries, name.c_str(), (unsigned long)candidate);

        // Note: The range 60000-64999 is reserved on Debian platforms
        //       and should be avoided and the value 65535 is traditionally
        //       reserved as an "error" code.
        if (!((candidate >= 60000) && (candidate <= 64999)) &&
             (candidate != 65535) &&
            !::getpwuid(candidate)) /* make sure not already allocated */
        {
            full_cmd = base_cmd + " --uid " + std::to_string(candidate) + ' ' + username;

            LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": executing \"%s\"", username.c_str(), full_cmd.c_str());

            int rc;
            std::string std_out;
            std::string std_err;
            std::tie(rc, std_out, std_err) = run(full_cmd);

            LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": command returned rc=%d, stdout=%s, stderr=%s",
                            username.c_str(), rc, std_out.c_str(), std_err.c_str());

            return rc == 0;
        }
        else
        {
            // Try with a slightly different name
            name = username + std::to_string(n_tries);
            LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": candidate UID=%lu already in use. Retry with name = \"%s\"",
                            username.c_str(), (unsigned long)candidate, name.c_str());
        }
    }

    syslog(LOG_ERR, "User \"%s\": unable to create unconfirmed user after %d attempts",
           username.c_str(), n_tries);

    return false;
}

/**
 * @brief This is a DBus interface used by remote programs to confirm a
 *        user.
 *
 * @param username  Username to be confirmed
 * @param groupname User's Primary group
 * @param groups    User's Supplementory groups (comma-separated list)
 * @param label     Label to be added in the comment (e.g. "RADIUS",
 *                  "TACACS+", "AAA", etc...)
 *
 * @return bool     true if user was confirmed successfully,
 *                  false otherwise.
 */
bool hamd_c::confirm_user(const std::string& username, const std::string& groupname, const std::string& groups, const std::string& label)
{
    std::string  cmd("/usr/sbin/usermod --comment \"Automagic user");

    if (!label.empty())
        cmd += ' ' + label;

    cmd += '"';

    if (!groups.empty())
        cmd += " --append --groups " + groups;

    cmd += " --gid " + groupname + ' ' + username;

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": executing \"%s\"", username.c_str(), cmd.c_str());

    int rc;
    std::string std_out;
    std::string std_err;
    std::tie(rc, std_out, std_err) = run(cmd);

    LOG_CONDITIONAL(is_tron(), LOG_DEBUG, "User \"%s\": command returned rc=%d, stdout=%s, stderr=%s",
                    username.c_str(), rc, std_out.c_str(), std_err.c_str());
    if (rc != 0)
        return false;

    std::string errmsg = certgen(username);
    return errmsg.empty();
}

/**
 * @brief This is a DBus interface used to turn tracing on. This allows
 *        the daemon to run with verbosity turned on.
 *
 * @return std::string
 */
std::string hamd_c::tron()
{
    config_rm.tron_m = true;
    return "Tracing is now ON";
}

/**
 * @brief This is a DBus interface used to turn tracing off. This allows
 *        the daemon to run with verbosity turned off.
 *
 * @return std::string
 */
std::string hamd_c::troff()
{
    config_rm.tron_m = false;
    return "Tracing is now OFF";
}

/**
 * @brief This is a DBus interface used to retrieve daemon running info
 *
 * @return std::string
 */
std::string hamd_c::show()
{
    std::ostringstream  oss;
    oss << "Process data:\n"
        << "  PID                       = " << getpid() << '\n'
        << "  poll_timer_m              = " << poll_timer_m << '\n'
        << '\n'
        << config_rm << '\n';

    return oss.str();
}
