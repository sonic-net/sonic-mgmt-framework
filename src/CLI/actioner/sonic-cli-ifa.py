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

    # Set/Get the rules of all IFA table entries.
    if func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FEATURE_TABLE')
       return api.get(path)
    elif func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE')
       return api.get(path)
    elif func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE/TAM_INT_IFA_FLOW_TABLE_LIST={name}', name=args[0])
       return api.get(path)
    elif func == 'patch_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table_tam_int_ifa_feature_table_list_enable':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FEATURE_TABLE/TAM_INT_IFA_FEATURE_TABLE_LIST={name}/enable', name=args[0])
       if args[1] == 'True':
           body = { "sonic-ifa:enable": True }
       else:
           body = { "sonic-ifa:enable": False }
       return api.patch(path, body)
    elif func == 'patch_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE/TAM_INT_IFA_FLOW_TABLE_LIST={name}', name=args[0])
       bodydict = {"name": args[0], "acl-rule-name": args[1], "acl-table-name": args[2]}
       for i in range(len(args)):
           if args[i] == "sv":
               if args[i+1] != "cv":
                   bodydict["sampling-rate"] = int(args[i+1])
           elif args[i] == "cv":
               if i+1 < len(args):
                   bodydict["collector-name"] = args[i+1]
           else:
               pass
       body = { "sonic-ifa:TAM_INT_IFA_FLOW_TABLE_LIST": [ bodydict ] }
       return api.patch(path, body)
    elif func == 'delete_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE/TAM_INT_IFA_FLOW_TABLE_LIST={name}', name=args[0])
       return api.delete(path)
    elif func == 'delete_list_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE/TAM_INT_IFA_FLOW_TABLE_LIST',)
       return api.delete(path)
    else:
       body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if 'sonic-ifa:sonic-ifa' in api_response:
                value = api_response['sonic-ifa:sonic-ifa']
                if 'TAM_INT_IFA_FEATURE_TABLE' in value:
                    tup = value['TAM_INT_IFA_FEATURE_TABLE']
                elif 'TAM_INT_IFA_FLOW_TABLE' in value:
                    tup = value['TAM_INT_IFA_FLOW_TABLE']
                else:
                    api_response = None

            if api_response is None:
                print("Failed")
            elif func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
                show_cli_output(args[1], api_response)
            else:
                return
    else:
        print response.error_message()

def get_tam_ifa_status(args):
    api_response = {}
    api = cc.ApiClient()

    path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_DEVICE_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['device'] = response.content['sonic-tam:TAM_DEVICE_TABLE']['TAM_DEVICE_TABLE_LIST']

    path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_COLLECTOR_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['collector'] = response.content['sonic-tam:TAM_COLLECTOR_TABLE']['TAM_COLLECTOR_TABLE_LIST']

    path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FEATURE_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            print(response.content)
            api_response['feature'] = response.content['sonic-ifa:TAM_INT_IFA_FEATURE_TABLE']['TAM_INT_IFA_FEATURE_TABLE_LIST']

    path = cc.Path('/restconf/data/sonic-ifa:sonic-ifa/TAM_INT_IFA_FLOW_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['flow'] = response.content['sonic-ifa:TAM_INT_IFA_FLOW_TABLE']['TAM_INT_IFA_FLOW_TABLE_LIST']

    show_cli_output("show_tam_ifa_status.j2", api_response)

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    if func == 'get_tam_ifa_status':
        get_tam_ifa_status(sys.argv[2:])

    run(func, sys.argv[2:])

