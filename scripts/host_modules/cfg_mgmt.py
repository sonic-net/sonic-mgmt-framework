""" Config management handler"""
import host_service
import subprocess
import os

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
        
    @staticmethod
    def _get_version():
        '''Return the SONiC version string, or NONE if command to retrieve it fails'''
        try:
            proc = subprocess.Popen("sonic-cfggen -y /etc/sonic/sonic_version.yml -v build_version", shell=True, stdout=subprocess.PIPE)
            out,err = proc.communicate()
            build_version_info = out.strip()
            return build_version_info.decode("utf-8")
        except:
            return None

    @staticmethod
    def _create_host_file(fname, content=None):
        '''Create a file under /host'''
        version = CFG_MGMT._get_version()
        if version:
            filename = '/host/image-' + version + "/" + fname
            try:
                f = open(filename, "w+")
                if content:
                    f.write(content)
                f.close()
                return 0, ""
            except IOError as e:
                return 1, ("Unable to create file [%s] - %s" % (filename, e))
        else:
            return 1, "Unable to get SONiC version: operation not performed"

    @staticmethod
    def _delete_host_file(fname):
        '''Delete a file under /host'''
        version = CFG_MGMT._get_version()
        if version:
            filename = '/host/image-' + version + "/" + fname
            if not os.path.exists(filename):
                return 1, "No configuration erase operation to cancel."
            try:
                os.remove(filename)
                return 0, ""
            except IOError as e:
                return 1, ("Unable to delete file [%s] - %s" % (filename, e))
        else:
            return 1, "Unable to get SONiC version: operation not performed"


    @staticmethod
    def _run_command_erase(option):
        """ Run config mgmt command """
        rc = 1
        if option == "":
            rc,err = CFG_MGMT._create_host_file("/pending_erase")
        elif option == "boot":
            rc,err = CFG_MGMT._create_host_file("/pending_erase", "boot")
        elif option == "install":
            rc,err = CFG_MGMT._create_host_file("/pending_erase", "install")
        elif option == "no":
            rc,err = CFG_MGMT._delete_host_file("/pending_erase")
        return rc, err

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='s', out_signature='is')
    def write_erase(self, option):
        return CFG_MGMT._run_command_erase(option)

def register():
    """Return class name"""
    return CFG_MGMT, MOD_NAME
