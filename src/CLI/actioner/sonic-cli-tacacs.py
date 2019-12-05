#!/usr/bin/python
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
import json
import collections
import re
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
import urllib3
urllib3.disable_warnings()

def invoke_api(func, args):
    body = None
    api = cc.ApiClient()

    # Set/Get tacacs configuration
    if func == 'patch_openconfig_system_augments_system_aaa_server_groups_server_group_config_source_address':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:source-address')
       body = { "openconfig-system-augments:source-address": args[0] }
       return api.patch(path, body)
    elif func == 'patch_openconfig_system_augments_system_aaa_server_groups_server_group_config_timeout':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:timeout')
       body = { "openconfig-system-augments:timeout": int(args[0]) }
       return api.patch(path, body)
    elif func == 'patch_openconfig_system_augments_system_aaa_server_groups_server_group_config_auth_type':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:auth-type')
       body = { "openconfig-system-augments:auth-type": args[0] }
       return api.patch(path, body)
    elif func == 'patch_openconfig_system_augments_system_aaa_server_groups_server_group_config_secret_key':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:secret-key')
       body = { "openconfig-system-augments:secret-key": args[0] }
       return api.patch(path, body)
    elif func == 'patch_tacacs_server':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/servers/server={address}/tacacs/config', address=args[0])
       body = {
         "openconfig-system:config": {
           "port": int(args[1]),
           "secret-key": args[2]
         }
       }
       api.patch(path, body)
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/servers/server={address}/config', address=args[0])
       body = {
           "openconfig-system:config": {
             "timeout": int(args[3]),
             "openconfig-system-augments:auth-type": args[4],
             "openconfig-system-augments:priority": int(args[5])
          }
       }
       return api.patch(path, body)
    elif func == 'delete_openconfig_system_system_aaa_server_groups_server_group_servers_server_tacacs_config':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/servers/server={address}/tacacs/config', address=args[0])
       return api.delete(path)
    elif func == 'delete_openconfig_system_augments_system_aaa_server_groups_server_group_config_source_address':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:source-address')
       return api.delete(path)
    elif func == 'delete_openconfig_system_augments_system_aaa_server_groups_server_group_config_timeout':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:timeout')
       return api.delete(path)
    elif func == 'delete_openconfig_system_augments_system_aaa_server_groups_server_group_config_auth_type':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:auth-type')
       return api.delete(path)
    elif func == 'delete_openconfig_system_augments_system_aaa_server_groups_server_group_config_secret-key':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config/openconfig-system-augments:secret-key')
       return api.delete(path)
    else:
       body = {}

    return api.cli_not_implemented(func)

def get_sonic_tacacs_server(args):
    api_response = {}
    api = cc.ApiClient()

    path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/servers/server={address}/tacacs/config', address=args[0])
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['port'] = response.content['openconfig-system:config']['port']
            api_response['key'] = response.content['openconfig-system:config']['secret-key']

    path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/servers/server={address}/config', address=args[0])
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['authtype'] = response.content['openconfig-system:config']['openconfig-system-augments:auth-type']
            api_response['priority'] = response.content['openconfig-system:config']['openconfig-system-augments:priority']
            api_response['timeout'] = response.content['openconfig-system:config']['timeout']

    api_response['address'] = args[0]
    show_cli_output("show_tacacs_server.j2", api_response)

def get_sonic_tacacs_global():
    api_response = {}
    api = cc.ApiClient()

    path = cc.Path('/restconf/data/openconfig-system:system/aaa/server-groups/server-group=TACACS/config')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response = response.content

    show_cli_output("show_tacacs_global.j2", api_response)

def run(func, args):
    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content

            if api_response is None:
                print("Failed")
            else:
                return
    else:
        print response.error_message()

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    if func == 'get_sonic_tacacs_global':
        get_sonic_tacacs_global()
    elif func == 'get_sonic_tacacs_server':
        get_sonic_tacacs_server(sys.argv[2:])
    else:
        run(func, sys.argv[2:])

