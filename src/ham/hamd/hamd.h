// Host Account Management
#ifndef HAMD_H
#define HAMD_H

#include <dbus-c++/dbus.h>          // DBus::Connection
#include <glib.h>                   // gint, gpointer
#include "timer.h"                  // gtimer_c
#include "../shared/dbus-address.h" // DBUS_BUS_NAME_BASE

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-but-set-variable"   /* SUPPRESS: warning: variable 'ri' set but not used [-Wunused-but-set-variable] */
#include "../shared/org.SONiC.HostAccountManagement.dbus-adaptor.h"
#pragma GCC diagnostic pop

class hamd_config_c
{
public:
    hamd_config_c(int argc, char **argv);

    /**
     * @brief Read configuration and update hamd_config_c object
     */
    void reload();

    uid_t uid_fit_into_range(uint64_t hash) const
    {
        return (uid_t)((hash % sac_uid_range_m) + sac_uid_min_m);
    }

    /**
     * @brief return the shell program to assign to new users.
     */
    const std::string & shell() const {return shell_m;}

    /**
     * @brief return the command to execute to create a user's certificates.
     */
    std::string certgen_cmd(const std::string & user_r, const std::string & certdir_r) const;

    std::string to_string() const;
    bool        is_tron()   const { return tron_m; }

private:
    // PLEASE UPDATE etc/sonic/hamd/config
    // WHEN MAKING CHANGES TO DEFAULTS
    static const  gint  poll_period_sec_default_m = 30;
    static const  gint  sac_uid_min_default_m     = 5000;  // System-Assigned IDs will be in the
    static const  gint  sac_uid_max_default_m     = 59999; // range [sac_uid_min_m..sac_uid_max_m]
    static const  bool  tron_default_m            = false;
    const gchar       * conf_file_default_pm      = "/etc/sonic/hamd/config";
    std::string         certgen_cmd_default_m     = "/usr/bin/openssl req -newkey rsa:2048 -nodes -keyout $CERTDIR/key.pem -subj \"/O=SONiC/OU=CLI/CN=$USERNAME\" | /usr/bin/openssl x509 -req -days 365000 -out $CERTDIR/certificate.pem -CA /root/cli-ca/cert.pem -CAkey /root/cli-ca/key.pem -CAcreateserial -sha256";
    //std::string         shell_default_m           = "/usr/bin/sonic-cli";
    std::string         shell_default_m           = ""; // empty string -> let linux assign default shell

public:
    bool                tron_m                    = tron_default_m;
    gint                poll_period_sec_m         = poll_period_sec_default_m;

    gint                sac_uid_min_m             = sac_uid_min_default_m;  // System-Assigned IDs will be in the
    gint                sac_uid_max_m             = sac_uid_max_default_m;  // range [sac_uid_min_m..sac_uid_max_m]

private:
    const gchar       * conf_file_pm              = conf_file_default_pm;
    std::string         shell_m                   = shell_default_m;
    gint                sac_uid_range_m           = 1 + (sac_uid_max_m - sac_uid_min_m);
    std::string         certgen_cmd_m             = certgen_cmd_default_m;
};

std::ostream & operator<<(std::ostream  & stream_r, const hamd_config_c  & obj_r);



class hamd_c : public DBus::ObjectAdaptor,
               public DBus::IntrospectableAdaptor,
               public ham::accounts_adaptor,
               public ham::name_service_adaptor,
               public ham::debug_adaptor
{
public:
    hamd_c(hamd_config_c & config_r, DBus::Connection  & conn_r);
    virtual ~hamd_c() {}

    // DBus "accounts" interface
    virtual ::DBus::Struct< bool, std::string > useradd(const std::string& login, const std::vector< std::string >& roles, const std::string& hashed_pw);
    virtual ::DBus::Struct< bool, std::string > userdel(const std::string& login);
    virtual ::DBus::Struct< bool, std::string > passwd(const std::string& login, const std::string& hashed_pw);
    virtual ::DBus::Struct< bool, std::string > set_roles(const std::string& login, const std::vector< std::string >& roles);
    virtual ::DBus::Struct< bool, std::string > groupadd(const std::string& group);
    virtual ::DBus::Struct< bool, std::string > groupdel(const std::string& group);

    // DBus "nss" interface
    virtual ::DBus::Struct< bool, std::string, std::string, uint32_t, uint32_t, std::string, std::string, std::string > getpwnam(const std::string& name);
    virtual ::DBus::Struct< bool, std::string, std::string, uint32_t, uint32_t, std::string, std::string, std::string > getpwuid(const uint32_t& uid);
    virtual ::DBus::Struct< bool, std::string, std::string, uint32_t, std::vector< std::string > > getgrnam(const std::string& name);
    virtual ::DBus::Struct< bool, std::string, std::string, uint32_t, std::vector< std::string > > getgrgid(const uint32_t& gid);
    virtual ::DBus::Struct< bool, std::string, std::string, int32_t, int32_t, int32_t, int32_t, int32_t, int32_t, uint32_t > getspnam(const std::string& name);

    // DBus "sac" interface
    virtual bool add_unconfirmed_user(const std::string & username, const uint32_t & pid);
    virtual bool confirm_user(const std::string & username, const std::string & groupname, const std::string & groups, const std::string & label);

    // DBus "debug" interface
    virtual std::string  tron();
    virtual std::string  troff();
    virtual std::string  show();

    bool                 is_tron() const { return config_rm.tron_m; }
    virtual void         cleanup();
    void                 reload();
    void                 apply_config();

private:
    hamd_config_c      & config_rm;
    gtimer_c             poll_timer_m;
    static bool          on_poll_timeout(gpointer user_data_p); // This callback functions must follow GSourceFunc signature.
    void                 rm_unconfirmed_users() const;
    std::string          certgen(const std::string  & login) const;
};

#endif /* HAMD_H */
