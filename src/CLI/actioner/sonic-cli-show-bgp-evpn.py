#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom, Inc.
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

vniDict = {}

def apply_type_filter(response, rt_type):
    new_list = []
    if 'openconfig-bgp-evpn-ext:routes' in response:
        if 'route' in response['openconfig-bgp-evpn-ext:routes']:
            for i in range(len(response['openconfig-bgp-evpn-ext:routes']['route'])):
                route = response['openconfig-bgp-evpn-ext:routes']['route'][i]
                t = route['prefix'].split(':')[0].rstrip(']').lstrip('[')
                if rt_type == t:
                    new_list.append(route)
    response['openconfig-bgp-evpn-ext:routes']['route'] = new_list
    return response

def apply_macip_filter(response, mac, ip):
    new_list = []
    if 'openconfig-bgp-evpn-ext:routes' in response:
        if 'route' in response['openconfig-bgp-evpn-ext:routes']:
            for i in range(len(response['openconfig-bgp-evpn-ext:routes']['route'])):
                route = response['openconfig-bgp-evpn-ext:routes']['route'][i]
                t = route['prefix'].split(':')[0].rstrip(']').lstrip('[')
                if '2' == t and mac in route['prefix'] and ip in route['prefix']:
                    new_list.append(route)
    response['openconfig-bgp-evpn-ext:routes']['route'] = new_list
    return response

def apply_rd_filter(response, rd):
    new_list = []
    if 'openconfig-bgp-evpn-ext:routes' in response:
        if 'route' in response['openconfig-bgp-evpn-ext:routes']:
            for i in range(len(response['openconfig-bgp-evpn-ext:routes']['route'])):
                route = response['openconfig-bgp-evpn-ext:routes']['route'][i]
                if rd == route['route-distinguisher']:
                    new_list.append(route)
    response['openconfig-bgp-evpn-ext:routes']['route'] = new_list
    return response

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None
    
    if func == 'get_bgp_evpn_vni':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/state',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        return api.get(keypath)
    elif func == 'get_bgp_evpn_routes':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/rib/afi-safis/afi-safi={af_name}/openconfig-bgp-evpn-ext:l2vpn-evpn'
            +'/loc-rib/routes',
                vrf=args[0], af_name=args[1])
        return api.get(keypath)
    elif func == 'get_bgp_evpn_routes_filter':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/rib/afi-safis/afi-safi={af_name}/openconfig-bgp-evpn-ext:l2vpn-evpn'
            +'/loc-rib/routes',
                vrf=args[0], af_name=args[1])
        response = api.get(keypath)
        filter_type = args[2].split('=')[1]
        if filter_type == "type":
            apply_type_filter(response.content, args[3])
        elif filter_type == "rd":
            apply_rd_filter(response.content, args[3])
        elif filter_type == "rd,type":
            apply_rd_filter(response.content, args[3])
            apply_type_filter(response.content, args[4])
        elif filter_type == "rd,macip":
            apply_rd_filter(response.content, args[3])
            apply_macip_filter(response.content, args[4], args[5])
        else:
            print("Unsupported filter ", filter_type)
        return response

    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args, renderer):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            api_response = response.content
            show_cli_output(renderer, api_response)
        else:
            print("Empty response")
    else:
        print(response.error_message())

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[3:], sys.argv[2])

