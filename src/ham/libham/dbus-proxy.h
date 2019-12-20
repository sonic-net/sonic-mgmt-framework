// Host Account Management
#ifndef DBUS_PROXY_H
#define DBUS_PROXY_H

#include <dbus-c++/dbus.h>  // DBus

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-but-set-variable"   /* SUPPRESS: warning: variable 'ri' set but not used [-Wunused-but-set-variable] */
#include "../shared/org.SONiC.HostAccountManagement.dbus-proxy.h"
#pragma GCC diagnostic pop

class accounts_proxy_c : public ham::accounts_proxy,
                         public DBus::IntrospectableProxy,
                         public DBus::ObjectProxy
{
public:
    accounts_proxy_c(DBus::Connection &connection, const char * dbus_bus_name_p, const char * dbus_obj_name_p) :
    DBus::ObjectProxy(connection, dbus_obj_name_p, dbus_bus_name_p)
    {
    }
};



#endif // DBUS_PROXY_H
