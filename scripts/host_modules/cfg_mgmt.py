""" Config management handler"""
# pylint: disable=invalid-name
from __future__ import print_function

import json
import os
import shutil
import subprocess
import tarfile
import tempfile

import host_service

MOD_NAME = 'cfg_mgmt'

CFG_FILE = '/etc/sonic/cfg_mgmt.json'
DEFAULT_FILE = '/etc/sonic/sonic-config.tar'

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

        cmd.extend(str(x) for x in options)
        output = ""
        try:
            print("cmd", cmd)
            rc = 0
            output = subprocess.check_output(cmd)
            print('Output -> ', output)

        except subprocess.CalledProcessError as err:
            print("Exception when calling get_sonic_error -> %s\n" %(err))
            rc = err.returncode
            output = err.output

        return rc, output

    @staticmethod
    def _docker_copy(container, _file, tempdir, copy_to_container):
        # Copy files/folders to/from containers
        if copy_to_container:
            src = os.path.join(tempdir, container, _file.lstrip('/'))
            dst = container + ':' + _file
        else:
            src = container + ':' + _file
            dst = os.path.join(tempdir, container, _file.lstrip('/'))
            try:
                os.makedirs(os.path.dirname(dst), exist_ok=True)
            except OSError:
                pass

        cmd = ['/usr/bin/docker', 'cp', '-aL', src, dst]
        try:
            rc = 0
            subprocess.check_call(cmd)
        except subprocess.CalledProcessError as err:
            rc = err.returncode

        return rc

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def save(self, options):
        """Save configuration"""
        tempdir = tempfile.mkdtemp()
        try:
            config = self._load_config()
            for module, modcfg in config.items():
                destdir = os.path.join(tempdir, module)
                os.mkdir(destdir)

                container = modcfg.get('container', None)
                if container is None:
                    # Copy from host
                    for _file in modcfg.get('files', []):
                        # Copy individual files
                        src = _file
                        dst = os.path.join(destdir, os.path.dirname(_file.lstrip('/')))
                        try:
                            os.makedirs(dst, exist_ok=True)
                        except OSError:
                            # dir already exists
                            pass
                        shutil.copy(src, dst)

                    for folder in modcfg.get('folders', []):
                        # Copy directory trees
                        src = folder
                        dst = os.path.join(destdir, folder.lstrip('/'))
                        shutil.copytree(src, dst, symlinks=True)

                else:
                    # Copy from container
                    for _file in modcfg.get('files', []):
                        self._docker_copy(container, _file, tempdir, False)
                    for folder in modcfg.get('folders', []):
                        self._docker_copy(container, folder, tempdir, False)

            tfile = os.path.join(tempdir, 'config_db.json')
            rc, output = CFG_MGMT._run_command(["save", "-y"], [tfile])
            if rc != 0:
                return rc, output

            # Copy config_db.json to /etc/sonic
            shutil.copy(tfile, '/etc/sonic')

            if options:
                filename = options[0]
            else:
                filename = DEFAULT_FILE

            with tarfile.open(filename, 'w') as tf:
                tf.add(tempdir, arcname='.')

        finally:
            shutil.rmtree(tempdir, ignore_errors=True)

        return 0, filename

    def _load_from_tar(self, load_type, options):
        if options:
            filename = options[0]
        else:
            filename = DEFAULT_FILE

        tempdir = tempfile.mkdtemp()
        try:
            with tarfile.open(filename, 'r') as tf:
                tf.extractall(tempdir)

            config = self._load_config()
            for module, modcfg in config.items():
                container = modcfg.get('container', None)
                srcdir = os.path.join(tempdir, module)
                if container is None:
                    # Copy from host
                    for _file in modcfg.get('files', []):
                        # Copy individual files
                        src = os.path.join(srcdir, _file.lstrip('/'))
                        dst = os.path.dirname(_file)
                        try:
                            os.makedirs(dst, exist_ok=True)
                        except OSError:
                            # dir already exists
                            pass
                        shutil.copy(src, dst)

                    for folder in modcfg.get('folders', []):
                        # Copy directory trees
                        src = os.path.join(srcdir, folder.lstrip('/'))
                        dst = folder
                        try:
                            # We have to shell out to /bin/cp, since copytree
                            # will abort if the destination already exists
                            cmd = ['/bin/cp', '-a', src, dst]
                            subprocess.check_call(cmd)
                        except subprocess.CalledProcessError as err:
                            pass
                else:
                    # Copy to container
                    for _file in modcfg.get('files', []):
                        self._docker_copy(container, _file, tempdir, True)
                    for folder in modcfg.get('folders', []):
                        self._docker_copy(container, folder, tempdir, True)

            config_db_json = os.path.join(tempdir, 'config_db.json')
            rc, output = CFG_MGMT._run_command([load_type, "-y"], [config_db_json])
        except (tarfile.ReadError, FileNotFoundError):
            # Either file is not found or file is not a tarfile
            # If options is not given, the default file will be
            # /etc/sonic/sonic_config.tar. This file will not exist in
            # older releases, so on upgrade to newer releases, we will need
            # to fallback to the old `config load` command which will pull
            # from /etc/sonic/config_db.json. If options is given, then
            # try to read it as a tarfile, if that fails, then treat is as
            # a config_db.json dump.
            rc, output = CFG_MGMT._run_command([load_type, "-y"], options)
        finally:
            shutil.rmtree(tempdir, ignore_errors=True)

        return rc, output

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def reload(self, options):
        """Reload configuration and restart services"""
        return self._load_from_tar('reload', options)

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='as', out_signature='is')
    def load(self, options):
        """Load configuration"""
        return self._load_from_tar('load', options)

    @staticmethod
    def _get_version():
        '''Return the SONiC version string, or NONE if command to retrieve it fails'''
        try:
            proc = subprocess.Popen("sonic-cfggen -y /etc/sonic/sonic_version.yml -v build_version",
                                    shell=True, stdout=subprocess.PIPE)
            out, err = proc.communicate()
            build_version_info = out.strip()
            return build_version_info.decode("utf-8")
        except subprocess.CalledProcessError:
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
            rc, err = CFG_MGMT._create_host_file("/pending_erase")
        elif option == "boot":
            rc, err = CFG_MGMT._create_host_file("/pending_erase", "boot")
        elif option == "install":
            rc, err = CFG_MGMT._create_host_file("/pending_erase", "install")
        elif option == "no":
            rc, err = CFG_MGMT._delete_host_file("/pending_erase")
        return rc, err

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='s', out_signature='is')
    def write_erase(self, option):
        return self._run_command_erase(option)

    @staticmethod
    def _load_config():
        """Load configuration for save/load/reload"""
        config = {}
        if os.path.exists(CFG_FILE):
            with open(CFG_FILE, 'r') as cfg:
                config = json.load(cfg)

        return config

def register():
    """Return class name"""
    return CFG_MGMT, MOD_NAME
