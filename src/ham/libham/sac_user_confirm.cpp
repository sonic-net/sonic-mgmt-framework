#include <pwd.h>                    /* getpwnam() */
#include <grp.h>                    /* getgrnam() */
#include <systemd/sd-journal.h>     /* sd_journal_print() */
#include <unordered_set>            /* std::unordered_set */

#include "ham.h"
#include "dbus-proxy.h"
#include "../shared/dbus-address.h" /* DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE */
#include "../shared/utils.h"        /* streq(), strneq() */

bool sac_user_confirm(const char * login_p, const char * roles_p)
{
    bool  success = false;

    try
    {
        // DBus::default_dispatcher must be initialized before DBus::Connection.
        DBus::default_dispatcher = get_dispatcher();
        DBus::Connection    conn = DBus::Connection::SystemBus();
        sac_proxy_c         sac(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

        std::string errmsg = sac.user_confirm(login_p, split(roles_p, ','));
        success = errmsg.empty(); // empty error message means success
        if (!success)
            sd_journal_print(LOG_ERR, "SAC - user_confirm() - User \"%s\": Error! %s", login_p, errmsg);
    }
    catch (DBus::Error & ex)
    {
        sd_journal_print(LOG_ERR, "SAC - user_confirm() - User \"%s\": DBus Exception %s", login_p, ex.what());
    }

    return success;
}
