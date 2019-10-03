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
import openconfig_interfaces_client
from rpipe_utils import pipestr
from openconfig_interfaces_client.rest import ApiException
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()


plugins = dict()

def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func


def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    body = None
    keypath = []
    # Get the rules of all ACL table entries.
    if func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_description':
       keypath = [ args[0] ]
       body = { "openconfig-interfaces:description": args[1] }
    elif func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_enabled':
       keypath = [ args[0] ]
       if args[1] == "True":
           body = { "openconfig-interfaces:enabled": True }
       else:
           body = { "openconfig-interfaces:enabled": False }
    elif func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_mtu':
       keypath = [ args[0] ]
       body = { "openconfig-interfaces:mtu":  int(args[1]) }
    elif func.__name__ == 'patch_openconfig_if_ethernet_interfaces_interface_ethernet_config_auto_negotiate':
        keypath = [ args[0] ]
        if args[1] == "true":
            body = { "openconfig-if-ethernet:auto-negotiate": True }
        else :
            body = { "openconfig-if-ethernet:auto-negotiate": False }
    elif func.__name__ == 'patch_openconfig_if_ethernet_interfaces_interface_ethernet_config_port_speed':
        keypath = [ args[0] ]
        speed_map = {"10MBPS":"SPEED_10MB", "100MBPS":"SPEED_100MB", "1GIGE":"SPEED_1GB", "auto":"SPEED_1GB" }
        if args[1] not in speed_map.keys():
            print("%Error: Invalid port speed config")
            exit(1)
        else:
            speed = speed_map.get(args[1])

        body = { "openconfig-if-ethernet:port-speed": speed }
    elif func.__name__ == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config':
       sp = args[1].split('/')
       keypath = [ args[0], 0, sp[0] ]
       body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
    elif func.__name__ == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config':
       sp = args[1].split('/')
       keypath = [ args[0], 0, sp[0] ]
       body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
    elif func.__name__ == 'delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config_prefix_length':
       keypath = [args[0], 0, args[1]]
    elif func.__name__ == 'delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config_prefix_length':
       keypath = [args[0], 0, args[1]]
    elif func.__name__ == 'get_openconfig_interfaces_interfaces_interface':
	keypath = [args[0]]
    elif func.__name__ == 'get_openconfig_interfaces_interfaces':
        keypath = []
    else:
       body = {}

    return keypath,body

def getId(item):
    prfx = "Ethernet"
    state_dict = item['state']
    ifName = state_dict['name']

    if ifName.startswith(prfx):
        ifId = int(ifName[len(prfx):])
        return ifId
    return ifName

def run(func, args):

    c = openconfig_interfaces_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_interfaces_client.OpenconfigInterfacesApi(api_client=openconfig_interfaces_client.ApiClient(configuration=c))

# Code for PortChannel cli skeleton

    #create a port-channel        
    if "PortChannel" in args[0] and func.__name__ == 'patch_openconfig_interfaces_interfaces_interface':
        return

    #show port-channels summary
    if "PortChannel" in args[0] and func.__name__ == 'get_openconfig_interfaces_interfaces':

        dummy_resp= {
            "PORTCHANNEL": [
                {
                    "members": [
                        "Ethernet56",
                        "Ethernet60"
                    ],
                    "min-links": 2,
                    "mtu": 9100,
                    "admin_status": "up",
                    "oper_status": "down",
                    "name": "PortChannel1",
                    "id": "1"
                },
                {
                    "members": [],
                    "min-links": 1,
                    "mtu": 9100,
                    "admin_status": "up",
                    "oper_status": "down",
                    "name": "PortChannel12",
                    "id": "12"
                },
                {
                    "members": [],
                    "min-links": 1,
                    "mtu": 9100, 
                    "admin_status": "up",
                    "oper_status": "down",
                    "name": "PortChannel3",
                    "id": "3"
                }
            ]
        }
        show_cli_output(args[1], dummy_resp)
        return

    #add members to port-channel
    if func.__name__ == 'patch_openconfig_if_aggregate_interfaces_interface_ethernet_config_aggregate_id':
        return

    #remove members from port-channel
    if func.__name__ == 'delete_openconfig_if_aggregate_interfaces_interface_ethernet_config_aggregate_id':
        return

    #config mtu for port-channel
    if "po" in args[0] and func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_mtu':
        return

    #delete port-channel
    if "PortChannel" in args[0] and func.__name__ == 'delete_openconfig_interfaces_interfaces_interface':
        return

    #config min-links in port-channel
    if func.__name__ == 'patch_openconfig_if_aggregate_interfaces_interface_aggregation_config_min_links':
        return

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        # Temporary code for #show vlan command with dummy data
        if func.__name__ == "get_openconfig_vlan_interfaces_interface_ethernet_switched_vlan_state":
            api_response = {'Vlan100': {'Ethernet20': 'tagged', 'Ethernet40': 'untagged'}}
            show_cli_output(args[0], api_response)
            return
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            if 'openconfig-interfaces:interfaces' in api_response:
                value = api_response['openconfig-interfaces:interfaces']
                if 'interface' in value:
                    tup = value['interface']
                    value['interface'] = sorted(tup, key=getId)

            if api_response is None:
                print("Failed")
            else:
                if func.__name__ == 'get_openconfig_interfaces_interfaces_interface':
                     show_cli_output(args[1], api_response)
                elif func.__name__ == 'get_openconfig_interfaces_interfaces':
                     show_cli_output(args[0], api_response)
                else:
                     return
    except ApiException as e:
        #print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func.__name__, e))
        if e.body != "":
            body = json.loads(e.body)
            if "ietf-restconf:errors" in body:
                 err = body["ietf-restconf:errors"]
                 if "error" in err:
                     errList = err["error"]

                     errDict = {}
                     for dict in errList:
                         for k, v in dict.iteritems():
                              errDict[k] = v

                     if "error-message" in errDict:
                         print "%Error: " + errDict["error-message"]
                         return
                     print "%Error: Transaction Failure"
                     return
            print "%Error: Transaction Failure"
        else:
            print "Failed"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_interfaces_client.OpenconfigInterfacesApi.__dict__)

    run(func, sys.argv[2:])
