// Host Account Management
#include "ham.h"
#include "dbus-proxy.h"
#include "../shared/dbus-address.h" /* DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE */
#include "../shared/utils.h"        /* split() */

#include <syslog.h>

int ham_useradd(const char * login, const char * roles_p, const char * hashed_pw)
{
    DBus::BusDispatcher         dispatcher;
    DBus::default_dispatcher = &dispatcher;
    DBus::Connection conn    = DBus::Connection::SystemBus();

    accounts_proxy_c interface(conn, DBUS_BUS_NAME_BASE, DBUS_OBJ_PATH_BASE);

    std::vector< std::string > roles = split(roles_p, ',');

    ::DBus::Struct<bool, std::string> ret;

    try
    {
        ret = interface.useradd(login, roles, hashed_pw);
    }
    catch (DBus::Error & ex)
    {
        syslog(LOG_CRIT, "ham_useradd(login=\"%s\", roles_p=\"%s\" - Exception %s\n", login, roles_p, ex.what());
        ret._1 = false;
    }

    return ret._1;
}


