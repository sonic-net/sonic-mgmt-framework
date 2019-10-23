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
import openconfig_network_instance_client 
from rpipe_utils import pipestr
from openconfig_network_instance_client.rest import ApiException 
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

# Update with network instance API
def generate_body(func, args):
    body = None
    if func.__name__ == 'get_openconfig_network_instance_network_instances_network_instance_fdb_mac_table_entries':
        keypath = ['default']
    else:
       body = {}

    return keypath,body

def mac_fill_count(mac_entries):
    static = dynamic = 0
    for mac_entry in mac_entries:
        if mac_entry['state']['entry_type'] == 'STATIC':
            static += 1
        else:
            dynamic += 1

    mac_entry_table = {'vlan-mac': len(mac_entries),
                       'static-mac': static,
                       'dynamic-mac': dynamic,
                       'total-mac': (static + dynamic)
    }
    return mac_entry_table

def fill_mac_info(mac_entry):
    mac_entry_table = {'Vlan':mac_entry['vlan'], 
                        'mac-address':mac_entry['mac_address'],
                        'entry-type': mac_entry['state']['entry_type'],
                        'port': mac_entry['interface']
                                ['interface_ref']['state']['interface']
    }
    return mac_entry_table


def run(func, args):

    c = openconfig_network_instance_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_network_instance_client.OpenconfigNetworkInstanceApi(api_client=openconfig_network_instance_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
            api_response = getattr(aa, func.__name__)(*keypath, body=body)

        else:
            api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            response = api_response.to_dict()
        
        mac_entries = response['openconfig_network_instanceentries']['entry']
        mac_table_list = [] 
        if func.__name__ == 'get_openconfig_network_instance_network_instances_network_instance_fdb_mac_table_entries':
            if args[1] == 'show': #### -- show mac address table --- ###
                for mac_entry in mac_entries:
                    mac_table_list.append(fill_mac_info(mac_entry))

            elif args[1] == 'mac-addr': #### -- show mac address table [address <mac-address>]--- ###
                for mac_entry in mac_entries:
                    if args[2] == mac_entry['mac_address']:
                        mac_table_list.append(fill_mac_info(mac_entry))

            elif args[1] == 'vlan': #### -- show mac address table [vlan <vlan-id>]--- ###
                for mac_entry in mac_entries:
                    if (int(args[2]) == mac_entry['vlan']):
                        mac_table_list.append(fill_mac_info(mac_entry))
 
            elif args[1] == 'interface': #### -- show mac address table [interface {Ethernet <id> | Portchannel <id>}]--- ###
                for mac_entry in mac_entries:
                    if args[2] == mac_entry['interface']['interface_ref']['state']['interface']:
                        mac_table_list.append(fill_mac_info(mac_entry))

            #### -- show mac address table [static {address <mac-address> | vlan <vlan-id> | interface {Ethernet <id>| Portchannel <id>}}]--- ###
            elif args[1] == 'static':
                if args[2] == 'address':
                    for mac_entry in mac_entries:
                        if args[3] == mac_entry['mac_address'] and mac_entry['state']['entry_type'] == 'STATIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

                elif args[2] == 'vlan':
                    for mac_entry in mac_entries:
                        if (int(args[3]) == mac_entry['vlan']) and mac_entry['state']['entry_type'] == 'STATIC':
                            mac_table_list.append(fill_mac_info(mac_entry))
                
                elif args[2] == 'interface':
                    for mac_entry in mac_entries:
                        if args[3] == mac_entry['interface']['interface_ref']['state']['interface'] and mac_entry['state']['entry_type'] == 'STATIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

                else:
                    for mac_entry in mac_entries:
                        if mac_entry['state']['entry_type'] == 'STATIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

            #### -- show mac address table [dynamic {address <mac-address> | vlan <vlan-id> | interface {Ethernet <id>| Portchannel <id>}}]--- ###
            elif args[1] == 'dynamic':
                if args[2] == 'address':
                    for mac_entry in mac_entries:
                        if args[3] == mac_entry['mac_address'] and mac_entry['state']['entry_type'] == 'DYNAMIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

                elif args[2] == 'vlan':
                    for mac_entry in mac_entries:
                        if (int(args[3]) == mac_entry['vlan']) and mac_entry['state']['entry_type'] == 'DYNAMIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

                elif args[2] == 'interface':
                    for mac_entry in mac_entries:
                        if args[3] == mac_entry['interface']['interface_ref']['state']['interface'] and mac_entry['state']['entry_type'] == 'DYNAMIC':
                            mac_table_list.append(fill_mac_info(mac_entry))

                else:
                    for mac_entry in mac_entries:
                        if mac_entry['state']['entry_type'] == 'DYNAMIC':
                            mac_table_list.append(fill_mac_info(mac_entry))


            elif args[1] == 'count': #### -- show mac address table count --- ###
                mac_table_list.append(mac_fill_count(mac_entries))
            show_cli_output(args[0], mac_table_list)
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
    func = eval(sys.argv[1], globals(), openconfig_network_instance_client.OpenconfigNetworkInstanceApi.__dict__)

    run(func, sys.argv[2:])
