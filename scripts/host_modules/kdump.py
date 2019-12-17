"""kdump provided command handler"""

import host_service
import shlex
import subprocess
import re

MOD_NAME = 'kdump'

class Kdump(host_service.HostModule):
    """DBus endpoint that executes the kdump provided command
    """

    @staticmethod
    def _run_command(cmd):
        '''!
        Execute a given command

        @param cmd (str) Command to execute. Since we execute the command directly, and not within the
                         context of the shell, the full path needs to be provided ($PATH is not used).
                         Command parameters are simply separated by a space.
                         Should be either string or a list

        '''
        try:
            shcmd = shlex.split(cmd)
            proc = subprocess.Popen(shcmd, shell=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=1, close_fds=True)
            output_stdout, output_stderr = proc.communicate()
            list_stdout = []
            for l in output_stdout.splitlines():
                list_stdout.append(str(l.decode()))
            list_stderr = []
            for l in output_stderr.splitlines():
                list_stderr.append(str(l.decode()))
            return (proc.returncode, list_stdout, list_stderr)
        except (OSError, ValueError) as e:
            print("!Exception [%s] encountered while processing the command : %s" % (str(e), str(cmd)))
            return (1, None, None)

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='bis', out_signature='is')
    def command(self, enabled, num_dumps, memory):

        if memory != '':
            cmd = '/usr/bin/config kdump memory %s' % memory
        elif num_dumps != 0:
            cmd = '/usr/bin/config kdump num_dumps %d' % num_dumps
        else:
            if enabled:
                cmd = '/usr/bin/config kdump enable'
            else:
                cmd = '/usr/bin/config kdump disable'
        (rc, output, output_err) = self._run_command(cmd);

        result=''
        for s in output:
            if s != '':
                s = '\n' + s
            result = result + s

        return 0, result

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='s', out_signature='is')
    def state(self, param):

        # All results, formatted as a  string
        if param != None:
            cmd = '/usr/bin/show kdump %s' % param
        else:
            cmd = '/usr/bin/show kdump'
        (rc, output, output_err) = self._run_command(cmd);
        result=''
        for s in output:
            if s != '':
                s = '\n' + s
            result = result + s

        return 0, result

def register():
    """Return the class name"""
    return Kdump, MOD_NAME
