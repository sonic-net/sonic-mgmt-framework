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
import time
import json
import ast
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

IDENTIFIER='VRF'
NAME1='vrf'

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None

    if func == 'patch_openconfig_network_instance_network_instances_network_instance':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}', name=args[0])
        body = { "openconfig-network-instance:network-instance": [ { "name": args[0],
                                                                     "config" : { "name": args[0],
                                                                                  "type": args[1],
                                                                                  "enabled": True if args[2] == "True" else False } } ] }
        return api.patch(keypath, body)

    elif func == 'delete_openconfig_network_instance_network_instances_network_instance':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}', name=args[0])
        return api.delete(keypath)

    elif func == 'get_openconfig_network_instance_network_instances_network_instances':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance')
        return api.get(keypath)

    elif func == 'get_openconfig_network_instance_network_instances_network_instance':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/config', name=args[0])
        return api.get(keypath)

    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    try:
        api_response = invoke_api(func, args)

        if api_response.ok():
            response = api_response.content
            if response is None:
                return
            else:
                show_cli_output(args[1], response)

        else:
            # error response
            print api_response.error_message()

    except:
            # system/network error
            print "%Error: Transaction Failure"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

