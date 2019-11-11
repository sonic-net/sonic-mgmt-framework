"""ZTP command handler"""

import host_service
import subprocess

MOD_NAME = 'ztp'

class ZTP(host_service.HostModule):
    """DBus endpoint that executes ZTP related commands
    """
    @staticmethod
    def _run_command(commands):
        """Run a ZTP command"""
        cmd = ['/usr/bin/ztp']
        if isinstance(commands, list):
            cmd.extend(commands)
        else:
            cmd.append(commands)

        try:
            rc = 0
            output = subprocess.check_output(cmd)
        except subprocess.CalledProcessError as err:
            rc = err.returncode
            output = ""

        return rc, output

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='')
    def enable(self):
        self._run_command("enable")

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='')
    def disable(self):
        self._run_command(["disable", "-y"])

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='is')
    def status(self):
        return self._run_command("status")

def register():
    """Return the class name"""
    return ZTP
