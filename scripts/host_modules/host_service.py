"""Base class for host modules"""

import dbus.service
import dbus

BUS_NAME_BASE = 'org.SONiC.HostService'
BUS_PATH = '/org/SONiC/HostService'

def bus_name(mod_name):
    """Return the bus name for the service"""
    return BUS_NAME_BASE + '.' + mod_name

method = dbus.service.method

class HostModule(dbus.service.Object):
    """Base class for all host modules"""
    def __init__(self, mod_name):
        self.bus = dbus.SystemBus()
        self.bus_name = dbus.service.BusName(BUS_NAME_BASE, self.bus)
        super(HostModule, self).__init__(self.bus_name, BUS_PATH)
