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

    # Set/Get the rules of all IFA table entries.
    if func == 'get_sonic_tam_sonic_tam_tam_device_table':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_DEVICE_TABLE')
       return api.get(path)
    elif func == 'get_sonic_tam_sonic_tam_tam_collector_table':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_COLLECTOR_TABLE')
       return api.get(path)
    elif func == 'get_sonic_tam_sonic_tam_tam_collector_table_tam_collector_table_list':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_COLLECTOR_TABLE/TAM_COLLECTOR_TABLE_LIST={name}', name=args[0])
       return api.get(path)
    elif func == 'patch_sonic_tam_sonic_tam_tam_device_table_tam_device_table_list_deviceid':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_DEVICE_TABLE/TAM_DEVICE_TABLE_LIST={name}/deviceid', name=args[0])
       body = { "sonic-tam:deviceid": int(args[1]) }
       return api.patch(path, body)
    elif func == 'delete_sonic_tam_sonic_tam_tam_device_table_tam_device_table_list_deviceid':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_DEVICE_TABLE/TAM_DEVICE_TABLE_LIST={name}/deviceid', name=args[0])
       return api.delete(path, body)
    elif func == 'patch_list_sonic_tam_sonic_tam_tam_collector_table_tam_collector_table_list':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_COLLECTOR_TABLE/TAM_COLLECTOR_TABLE_LIST')
       body = {
           "sonic-tam:TAM_COLLECTOR_TABLE_LIST": [
              {
                  "name": args[0], "ipaddress-type": args[1], "ipaddress": args[2], "port": int(args[3])
              }
           ]
       }
       return api.patch(path, body)
    elif func == 'delete_sonic_tam_sonic_tam_tam_collector_table_tam_collector_table_list':
       path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_COLLECTOR_TABLE/TAM_COLLECTOR_TABLE_LIST={name}', name=args[0])
       return api.delete(path, body)
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
                print("api_response is None")
            elif func == 'get_sonic_tam_sonic_tam_tam_device_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_tam_sonic_tam_tam_collector_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_tam_sonic_tam_tam_collector_table_tam_collector_table_list':
                show_cli_output(args[1], api_response)
            else:
                return
    else:
        print "invoke_api failed"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

