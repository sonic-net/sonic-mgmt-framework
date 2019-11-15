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
        body = { "openconfig-network-instance:network-instance": [ { "name": args[0],
                                                                     "config" : { "name": args[0],
                                                                                  "type": args[1],
                                                                                  "enabled": True if args[2] == "True" else False } } ] }
        return api.patch(keypath, body)
    elif func == 'get_openconfig_network_instance_network_instances_network_instance':
        if args[0] == 'mgmt':
	    keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/config', name=args[0])
        else:
	    keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance/config')
        return api.get(keypath)
    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if api_response is None:
                print("%Error: Transaction Failure")
    else:
        print response.error_message()

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

