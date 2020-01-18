""" Fetch Environment handler"""
import host_service
import subprocess

MOD_NAME= 'fetch_environment'

class FETCH_ENVIRONMENT(host_service.HostModule):
    """DBus endpoint that executes CFG_MGMT related commands """

    @staticmethod
    def _run_command(options):
        """ Run show environment command """
        cmd = ['/usr/bin/show', 'environment']

        output = ""
        try:
            rc = 0
            output = subprocess.check_output(cmd)
            print('Output -> ', output)

        except subprocess.CalledProcessError as err:
            print ("Exception when calling get_sonic_error -> %s\n" %(err))
            rc = err.returncode
            output = err.output
        return rc, output

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def action(self, options):
        return FETCH_ENVIRONMENT._run_command(options)

def register ():
    """Return class name"""
    return FETCH_ENVIRONMENT, MOD_NAME
