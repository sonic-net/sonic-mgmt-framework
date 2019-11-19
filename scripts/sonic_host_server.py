#!/usr/bin/env python3
"""Host Service to handle docker-to-host communication"""

import os
import os.path
import glob
import importlib
import sys

import dbus
import dbus.service
import dbus.mainloop.glib

from gi.repository import GObject

BUS_NAME = 'org.SONiC.HostService'
BUS_PATH = '/org/SONiC/HostService'

def register_modules():
    """Register all host modules"""
    mod_path = os.path.join(os.path.dirname(__file__), 'host_modules')
    sys.path.append(mod_path)
    for mod_file in glob.glob(os.path.join(mod_path, '*.py')):
        if os.path.isfile(mod_file) and not mod_file.endswith('__init__.py'):
            mod_name = os.path.basename(mod_file)[:-3]
            module = importlib.import_module(mod_name)

            register_cb = getattr(module, 'register', None)
            if not register_cb:
                raise Exception('Missing register function for ' + mod_name)

            register_dbus(register_cb)

def register_dbus(register_cb):
    """Register DBus handlers for individual modules"""
    handler_class, mod_name = register_cb()
    handlers[mod_name] = handler_class(mod_name)

# Create a main loop reactor
GObject.threads_init()
dbus.mainloop.glib.threads_init()
dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)
loop = GObject.MainLoop()
handlers = {}

class SignalManager(object):
    ''' This is used to manage signals received (e.g. SIGINT).
        When stopping a process (systemctl stop [service]), systemd sends
        a SIGTERM signal.
    '''
    shutdown = False
    def __init__(self):
        ''' Install signal handlers.

            SIGTERM is invoked when systemd wants to stop the daemon.
            For example, "systemctl stop mydaemon.service"
            or,          "systemctl restart mydaemon.service"

        '''
        import signal
        signal.signal(signal.SIGTERM, self.sigterm_hdlr)

    def sigterm_hdlr(self, _signum, _frame):
        self.shutdown = True
        loop.quit()

sigmgr = SignalManager()
register_modules()

# Only run if we actually have some handlers
if handlers:
    import systemd.daemon
    systemd.daemon.notify("READY=1")

    while not sigmgr.shutdown:
        loop.run()
        if sigmgr.shutdown:
            break

    systemd.daemon.notify("STOPPING=1")
else:
    print("No handlers to register, quitting...")
