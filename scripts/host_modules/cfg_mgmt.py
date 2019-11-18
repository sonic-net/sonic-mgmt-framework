""" Config management handler"""
import host_service
import subprocess

MOD_NAME= 'cfg_mgmt'

class CFG_MGMT(host_service.HostModule):
    """DBus endpoint that executes CFG_MGMT related commands """

    @staticmethod
    def _run_command(commands, options):
        """ Run config mgmt command """
        cmd = ['/usr/bin/config']
        if isinstance(commands, list):
            cmd.extend(commands)
        else:
            cmd.append(commands)
        
        for x in options: 
            cmd.append(str(x)) 
        output =""
        try:
            print("cmd", cmd)
            rc = 0
            output= subprocess.check_output(cmd)
            print('Output -> ',output)

        except subprocess.CalledProcessError as err:
            print("Exception when calling get_sonic_error -> %s\n" %(err))
            rc = err.returncode
            output = err.output
            
        return rc,output

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def save(self, options):
        return CFG_MGMT._run_command(["save","-y"], options)

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def reload(self, options):
        return CFG_MGMT._run_command(["reload", "-y"], options)

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def load(self, options):
        return CFG_MGMT._run_command(["load", "-y"], options)
        
def register():
    """Return class name"""
    return CFG_MGMT, MOD_NAME
