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


def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    # SAG delete ipv4 anycast gateway address
    if func == 'del_llist_openconfig_interfaces_ext_interfaces_interface_subinterfaces_subinterface_ipv4_sag_ipv4_config_static_anycast_gateway' :
        sag_key = collections.defaultdict(dict)
        sag_key = {
            "name": args[0],
            "index": "0",
            "static-anycast-gateway": args[1]
                  }
				  
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/openconfig-interfaces-ext:sag-ipv4/config/static-anycast-gateway={static-anycast-gateway}',
                          **sag_key)
        return aa.delete(keypath)         

    # SAG configure ipv4 anycast gateway address
    if func == 'patch_openconfig_interfaces_ext_interfaces_interface_subinterfaces_subinterface_ipv4_sag_ipv4_config_static_anycast_gateway' :
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/openconfig-interfaces-ext:sag-ipv4/config/static-anycast-gateway',
                          name=args[0], index="0")
        body = collections.defaultdict(dict)
        body = {
                     "openconfig-interfaces-ext:static-anycast-gateway": [args[1]]
               }

        return aa.patch(keypath, body)
        
    # SAG delete ipv6 anycast gateway address
    if func == 'del_llist_openconfig_interfaces_ext_interfaces_interface_subinterfaces_subinterface_ipv6_sag_ipv6_config_static_anycast_gateway' :
        sag_key = collections.defaultdict(dict)
        sag_key = {
            "name": args[0],
            "index": "0",
            "static-anycast-gateway": args[1]
                  }
				  
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/openconfig-interfaces-ext:sag-ipv6/config/static-anycast-gateway={static-anycast-gateway}',
                          **sag_key)
        return aa.delete(keypath)         

    # SAG configure ipv6 anycast gateway address
    if func == 'patch_openconfig_interfaces_ext_interfaces_interface_subinterfaces_subinterface_ipv6_sag_ipv6_config_static_anycast_gateway' :
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/openconfig-interfaces-ext:sag-ipv6/config/static-anycast-gateway',
                          name=args[0], index="0")
        body = collections.defaultdict(dict)
        body = {
                    "openconfig-interfaces-ext:static-anycast-gateway": [args[1]]
               }

        return aa.patch(keypath, body)      

    # SAG delete global mac
    if func == 'delete_openconfig_network_instance_ext_network_instances_network_instance_global_sag_config_anycast_mac' :
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/openconfig-network-instance-ext:global-sag/config/anycast-mac',
                          name="default")
        return aa.delete(keypath)         


    # SAG configure global mac
    if func == 'patch_openconfig_network_instance_ext_network_instances_network_instance_global_sag_config_anycast_mac' :
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/openconfig-network-instance-ext:global-sag/config/anycast-mac',
                          name="default")
        body = collections.defaultdict(dict)
        body = {
                    "openconfig-network-instance-ext:anycast-mac": args[0]
               }

        return aa.patch(keypath, body)      

    # SAG IPv4 enable/disable
    if func == 'patch_openconfig_network_instance_ext_network_instances_network_instance_global_sag_config_ipv4_enable' :
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/openconfig-network-instance-ext:global-sag/config/ipv4-enable',
                          name="default")
        if args[0] == "True":
             body = { "openconfig-network-instance-ext:ipv4-enable": True }
        elif args[0] == "False":
             body = { "openconfig-network-instance-ext:ipv4-enable": False }                          
        return aa.patch(keypath, body)       

    # SAG IPv6 enable/disable
    if func == 'patch_openconfig_network_instance_ext_network_instances_network_instance_global_sag_config_ipv6_enable' :
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/openconfig-network-instance-ext:global-sag/config/ipv6-enable',
                          name="default")
        if args[0] == "True":
             body = { "openconfig-network-instance-ext:ipv6-enable": True }
        elif args[0] == "False":
             body = { "openconfig-network-instance-ext:ipv6-enable": False }                          
        return aa.patch(keypath, body) 

def run(func, args):
    try:
        api_response = invoke(func, args)

        if api_response.ok():
            response = api_response.content
            if response is None:
                pass
            elif 'openconfig-interfaces:config' in response.keys():
                value = response['openconfig-interfaces:config']
                if value is None:
                    return
                show_cli_output(args[2], value)
            elif 'openconfig-network-instance:config' in response.keys():
                value = response['openconfig-interfaces:config']
                if value is None:
                    return
                show_cli_output(args[2], value)             
        else:
            #error response
            print(api_response.error_message())

    except:
            # system/network error
            raise


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])


