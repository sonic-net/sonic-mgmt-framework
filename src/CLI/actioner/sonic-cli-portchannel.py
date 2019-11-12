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
import collections
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

pcDict = {}
memberDict = {}

def invoke_api(func, args=[]):
    api = cc.ApiClient()

    if func == 'get_sonic_portchannel_sonic_portchannel_lag_table':
        path = cc.Path('/restconf/data/sonic-portchannel:sonic-portchannel/LAG_TABLE')
        return api.get(path)

    if func == 'get_sonic_portchannel_sonic_portchannel_lag_member_table':
        path = cc.Path('/restconf/data/sonic-portchannel:sonic-portchannel/LAG_MEMBER_TABLE')
        return api.get(path)

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            laglst =[]
            if 'sonic-portchannel:LAG_TABLE' in api_response:
                value = api_response['sonic-portchannel:LAG_TABLE']
                if 'LAG_TABLE_LIST' in value:
                    laglst = value['LAG_TABLE_LIST']
            if api_response is None:
                print("Failed")
            else:
                if func == 'get_sonic_portchannel_sonic_portchannel_lag_table':
                    memlst=[]
                    # Get members for all PortChannels
                    memebrs_resp = invoke_api('get_sonic_portchannel_sonic_portchannel_lag_member_table')
                    if not memebrs_resp.ok():
                        print memebrs_resp.error_message()
                        return

                    api_response_members = memebrs_resp.content

                    if 'sonic-portchannel:LAG_MEMBER_TABLE' in api_response_members:
                        memlst = api_response_members['sonic-portchannel:LAG_MEMBER_TABLE']['LAG_MEMBER_TABLE_LIST']
                    for pc_dict in laglst:
                        pc_dict['members']=[]
                        pc_dict['type']="Eth"
                        for mem_dict in memlst:
                            if mem_dict['name'] == pc_dict['lagname']:
                                keypath = [mem_dict['ifname']]
                                pc_dict['members'].append(mem_dict['ifname'])
                    show_cli_output(args[0], laglst)
                else:
                     return
    else:
        print response.error_message()

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])
