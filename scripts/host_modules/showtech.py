"""Show techsupport command handler"""

import host_service
import subprocess
import re

MOD_NAME = 'showtech'

class Showtech(host_service.HostModule):
    """DBus endpoint that executes the "show techsupport" command
    """
    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='s', out_signature='is')
    def info(self, date):
        print("Host side: Running show techsupport")
        cmd = ['/usr/bin/generate_dump']
        if date:
            cmd.append("-s")
            cmd.append(date)

        try:
            rc = 0
            output = subprocess.check_output(cmd)
        except subprocess.CalledProcessError as err:
            rc = err.returncode
            output = ""
            print("Host side: Failed", rc)
            return rc, output
        output_string = output.decode("utf-8")
        output_file_match = re.search('\/var\/.*dump.*\.gz', output_string)
        if output_file_match is not None:
            output_filename = output_file_match.group()
        else:
            output_filename = ""
        return rc, output_filename

def register():
    """Return the class name"""
    return Showtech, MOD_NAME
