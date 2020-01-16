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
import cli_client as cc
import collections
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output

def invoke_api(func, args):
    api = cc.ApiClient()
    body = None

    if func == 'get_openconfig_lldp_lldp_interfaces':
       path = cc.Path('/restconf/data/openconfig-lldp:lldp/interfaces')
       return api.get(path)
    elif func == 'get_openconfig_lldp_lldp_interfaces_interface':
       path = cc.Path('/restconf/data/openconfig-lldp:lldp/interfaces/interface={name}', name=args[1])
       return api.get(path)
    else:
       body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content

            if api_response:
                response = api_response
                if 'openconfig-lldp:interfaces' in response.keys():
                    if not response['openconfig-lldp:interfaces']:
                        return
                    neigh_list = response['openconfig-lldp:interfaces']['interface']
                    if neigh_list is None:
                        return
                    show_cli_output(args[0],neigh_list)
                elif 'openconfig-lldp:interface' in response.keys():
                    neigh = response['openconfig-lldp:interface']#[0]['neighbors']['neighbor']
                    if neigh is None:
                        return
                    if args[1] is not None:
                        if 'state' in neigh[0]['neighbors']['neighbor'][0].keys():
                            show_cli_output(args[0],neigh)
                        else:
                            print('No LLDP neighbor interface')
                    else:
                        show_cli_output(args[0],neigh)
                else:
                    print("Failed")
    else:
        print response.error_message()

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

