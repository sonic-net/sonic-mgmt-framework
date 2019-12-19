#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or
# its subsidiaries.
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
import time
import json
import ast
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

SYSTEM='/restconf/data/openconfig-system:system/'
AAA=SYSTEM+'aaa/'
SERVER_GROUPS=AAA+'server-groups/'
RADIUS_SERVER_GROUP=SERVER_GROUPS+'server-group=RADIUS/'

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None

    if func == 'patch_openconfig_radius_global_config_source_address':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:source-address')
        body = { "openconfig-system-ext:source-address": args[0] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_radius_global_config_timeout':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:timeout')
        body = { "openconfig-system-ext:timeout": int(args[0]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_radius_global_config_retransmit':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:retransmit-attempts')
        body = { "openconfig-system-ext:retransmit-attempts": int(args[0])}
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_radius_global_config_key':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:secret-key')
        body = { "openconfig-system-ext:secret-key": args[0] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_radius_global_config_auth_type':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:auth-type')
        body = { "openconfig-system-ext:auth-type": args[0] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_radius_global_config_host':

        auth_port=(args[1])[10:]
        timeout=(args[2])[8:]
        retransmit=(args[3])[11:]
        key=(args[4])[4:]
        auth_type=(args[5])[10:]
        priority=(args[6])[9:]
        vrf=(args[7])[4:]

        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'servers/server={address}', address=args[0])
        body = {   "openconfig-system:server": [ {

                       "address": args[0],

                       "openconfig-system:config": {
                           "name": args[0],
                       },

                       "openconfig-system:radius": {
                           "openconfig-system:config": {
                           }
                       }
                  } ]
               }

        if len(auth_port) != 0:
            body["openconfig-system:server"][0]["openconfig-system:radius"]\
                ["openconfig-system:config"]["auth-port"] = int(auth_port)
        if len(retransmit) != 0:
            body["openconfig-system:server"][0]["openconfig-system:radius"]\
                ["openconfig-system:config"]["retransmit-attempts"] \
                = int(retransmit)
        if len(key) != 0:
            body["openconfig-system:server"][0]["openconfig-system:radius"]\
                ["openconfig-system:config"]["secret-key"] = key

        if len(timeout) != 0:
            body["openconfig-system:server"][0]["openconfig-system:config"]\
                ["timeout"] = int(timeout)
        if len(auth_type) != 0:
            body["openconfig-system:server"][0]["openconfig-system:config"]\
                ["openconfig-system-ext:auth-type"] = auth_type
        if len(priority) != 0:
            body["openconfig-system:server"][0]["openconfig-system:config"]\
                ["openconfig-system-ext:priority"] = int(priority)
        if len(vrf) != 0:
            body["openconfig-system:server"][0]["openconfig-system:config"]\
                ["openconfig-system-ext:vrf"] = vrf

        return api.patch(keypath, body)
    elif func == 'delete_openconfig_radius_global_config_source_address':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:source-address')
        return api.delete(keypath)
    elif func == 'delete_openconfig_radius_global_config_retransmit':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:retransmit-attempts')
        return api.delete(keypath)
    elif func == 'delete_openconfig_radius_global_config_key':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:secret-key')
        return api.delete(keypath)
    elif func == 'delete_openconfig_radius_global_config_auth_type':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:auth-type')
        return api.delete(keypath)
    elif func == 'delete_openconfig_radius_global_config_timeout':
        keypath = cc.Path(RADIUS_SERVER_GROUP +
            'config/openconfig-system-ext:timeout')
        return api.delete(keypath)
    elif func == 'delete_openconfig_radius_global_config_host':
        path = RADIUS_SERVER_GROUP + 'servers/server={address}'
        if (len(args) >= 2) and (len(args[1]) != 0):
            uri_suffix = {
                "auth-port": "/radius/config/auth-port",
                "retransmit": "/radius/config/retransmit-attempts",
                "key": "/radius/config/secret-key",
                "timeout": "/config/timeout",
                "auth-type": "/config/openconfig-system-ext:auth-type",
                "priority": "/config/openconfig-system-ext:priority",
                "vrf": "/config/openconfig-system-ext:vrf",
            }

            path = path + uri_suffix.get(args[1], "Invalid Attribute")

        keypath = cc.Path(path, address=args[0])
        return api.delete(keypath)
    else:
        body = {}

    return api.cli_not_implemented(func)

def get_sonic_radius_global():
    api_response = {} 
    api = cc.ApiClient()
    
    path = cc.Path(RADIUS_SERVER_GROUP+'config')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response = response.content

    show_cli_output("show_radius_global.j2", api_response)

def get_sonic_radius_servers(args):
    api_response = {}
    api = cc.ApiClient()

    path = cc.Path(RADIUS_SERVER_GROUP+'servers')
    response = api.get(path)


    if not response.ok():
        print("%Error: Get Failure")
        return

    if (not ('openconfig-system:servers' in response.content)) \
        or (not ('server' in response.content['openconfig-system:servers'])):
        return

    api_response['header'] = 'True'
    show_cli_output("show_radius_server.j2", api_response)

    for server in response.content['openconfig-system:servers']['server']:
        api_response.clear()
        api_response['header'] = 'False'
        if 'address' in server:
            api_response['address'] = server['address']

        if 'config' in server \
                and 'timeout' in server['config']:
            api_response['timeout'] = server['config']['timeout']

        if 'radius' in server \
                and 'config' in server['radius'] \
                and 'auth-port' in server['radius']['config']:
            api_response['port'] = server['radius']['config']['auth-port']

        if 'radius' in server \
                and 'config' in server['radius'] \
                and 'secret-key' in server['radius']['config']:
            api_response['key'] = server['radius']['config']['secret-key']

        if 'radius' in server \
                and 'config' in server['radius'] \
                and 'retransmit-attempts' in server['radius']['config']:
            api_response['retransmit'] = \
                server['radius']['config']['retransmit-attempts']

        if 'config' in server \
                and 'openconfig-system-ext:auth-type' in server['config']:
            api_response['authtype'] = \
                server['config']['openconfig-system-ext:auth-type']

        if 'config' in server \
                and 'openconfig-system-ext:priority' in server['config']:
            api_response['priority'] = \
                server['config']['openconfig-system-ext:priority']

        if 'config' in server \
                and 'openconfig-system-ext:vrf' in server['config']:
            api_response['vrf'] = \
                server['config']['openconfig-system-ext:vrf']

        show_cli_output("show_radius_server.j2", api_response)


def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if api_response is None:
                print("%Error: Transaction Failure")
    else:
        print(response.error_message())

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    if func == 'get_sonic_radius':
        get_sonic_radius_global()
        get_sonic_radius_servers(sys.argv[2:])
    else:
        run(func, sys.argv[2:])

