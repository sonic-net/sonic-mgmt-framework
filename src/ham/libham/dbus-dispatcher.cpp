#include "dbus-proxy.h" // get_dispatcher() prototype


// The dispatcher is a "main loop" construct that handles
// DBus messages. This should be defined as a singleton.
DBus::BusDispatcher * get_dispatcher()
{
    static DBus::BusDispatcher  dispatcher;
    return &dispatcher;
}

