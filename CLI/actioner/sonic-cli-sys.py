#:!/usr/bin/python3
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################

import sys
from cli_client import ApiClient
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


def util_capitalize(value):
    for key, val in list(value.items()):
        temp = key.replace('-', '_').split('_')
        alt_key = ''
        for i in temp:
            alt_key = alt_key + i.capitalize() + ' '
        value[alt_key] = value.pop(key)
    return value


def system_state_key_change(value):
    value.pop('motd-banner', None)
    value.pop('login-banner', None)
    return util_capitalize(value)


def memory_key_change(value):
    value['Total'] = value.pop('physical', 'Unknown')
    value['Used'] = value.pop('reserved', 'Unknown')
    return value


class Handlers:
    @staticmethod
    def get_openconfig_system_system_state(template, *args):
        resp = ApiClient().get("/restconf/data/openconfig-system:system/state")
        if not resp.ok():
            print(resp.error_message())
            return 1
        if resp.content:
            data = resp.content.get("openconfig-system:state", {})
            show_cli_output(template, system_state_key_change(data))
        return 0

    @staticmethod
    def get_openconfig_system_system_memory(template, *args):
        resp = ApiClient().get("/restconf/data/openconfig-system:system/memory/state")
        if not resp.ok():
            print(resp.error_message())
            return 1
        if resp.content:
            data = resp.content.get("openconfig-system:state", {})
            show_cli_output(template, memory_key_change(data))
        return 0

    @staticmethod
    def get_openconfig_system_system_cpus(template, *args):
        resp = ApiClient().get("/restconf/data/openconfig-system:system/cpus/cpu")
        if not resp.ok():
            print(resp.error_message())
            return 1
        if resp.content:
            data = resp.content.get("openconfig-system:cpu", [])
            show_cli_output(template, data)
        return 0

    @staticmethod
    def get_openconfig_system_system_processes(template, pid, *args):
        resp = ApiClient().get("/restconf/data/openconfig-system:system/processes/process")
        if not resp.ok():
            print(resp.error_message())
            return 1
        data = resp.content.get("openconfig-system:process", []) if resp.content else []
        # Show all process info if pid is not given
        if not pid.isnumeric():
            show_cli_output(template, data)
            return 0
        # Lookup specific process by pid and display its state.
        # Not passing pid in the GET url since server is not handling 'not found' case properly
        for proc in data:
            if proc["pid"] == pid:
                proc_state = proc.get("state", {})
                show_cli_output(template, util_capitalize(proc_state))
                break
        return 0


def run(func, args):
    return getattr(Handlers, func)(*args)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])
