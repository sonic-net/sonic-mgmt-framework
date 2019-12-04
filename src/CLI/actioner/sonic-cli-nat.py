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

import urllib3
urllib3.disable_warnings()


def invoke_api(func, args=[]):

    nat_type_map = {"snat" : "SNAT", "dnat": "DNAT"}
    nat_protocol_map = {"icmp": "1", "tcp": "6", "udp": "17"}

    api = cc.ApiClient()

    # Enable/Disable NAT Feature
    if func == 'patch_openconfig_nat_nat_instances_instance_config_enable':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/enable', id=args[0])
        if args[1] == "True":
           body = { "openconfig-nat:enable": True }
        else:
           body = { "openconfig-nat:enable": False }
        return api.patch(path,body)

    # Config NAT Timeout
    elif func == 'patch_openconfig_nat_nat_instances_instance_config_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/timeout', id=args[0])
        body = { "openconfig-nat:timeout":  int(args[1]) }
        return api.patch(path, body)

    # Config NAT TCP Timeout
    elif func == 'patch_openconfig_nat_nat_instances_instance_config_tcp_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/tcp-timeout', id=args[0])
        body = { "openconfig-nat:tcp-timeout":  int(args[1]) }
        return api.patch(path, body)

    # Config NAT UDP Timeout
    elif func == 'patch_openconfig_nat_nat_instances_instance_config_udp_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/udp-timeout', id=args[0])
        body = { "openconfig-nat:udp-timeout":  int(args[1]) }
        return api.patch(path, body)

    # Config NAT Static basic translation entry
    elif func == 'patch_openconfig_nat_nat_instances_instance_nat_mapping_table_nat_mapping_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-mapping-table/nat-mapping-entry={externaladdress}/config', id=args[0], externaladdress=args[1])
        body = { "openconfig-nat:config" : { "internal-address": args[2]} }
        l = len(args)
        if l >= 4:
            body["openconfig-nat:config"].update( {"type": nat_type_map[args[3]] } )
        if l == 5:
            body["openconfig-nat:config"].update( {"twice-nat-id": int(args[4])} )
        return api.patch(path, body)

    # Remove NAT Static basic translation entry
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_mapping_table_nat_mapping_entry_config_internal_address':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-mapping-table/nat-mapping-entry={externaladdress}/config/internal-address', id=args[0], externaladdress=args[1])
        return api.delete(path)

    # Remove all NAT Static basic translation entries
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_mapping_table':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-mapping-table', id=args[0])
        return api.delete(path)

    # Config NAPT Static translation entry
    elif func == 'patch_openconfig_nat_nat_instances_instance_napt_mapping_table_napt_mapping_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/napt-mapping-table/napt-mapping-entry={externaladdress},{protocol},{externalport}/config', id=args[0],externaladdress=args[1],protocol=nat_protocol_map[args[2]],externalport=args[3])
        body = { "openconfig-nat:config" : {"internal-address": args[4], "internal-port": int(args[5])} }
        l = len(args)
        if l >= 7:
            body["openconfig-nat:config"].update( {"type": nat_type_map[args[6]] } )
        if l == 8:
            body["openconfig-nat:config"].update( {"twice-nat-id": int(args[7])} )
        return api.patch(path, body)

    # Remove NAPT Static translation entry
    elif func == 'delete_openconfig_nat_nat_instances_instance_napt_mapping_table_napt_mapping_entry':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/napt-mapping-table/napt-mapping-entry={externaladdress},{protocol},{externalport}', id=args[0],externaladdress=args[1],protocol=nat_protocol_map[args[2]],externalport=args[3])
        return api.delete(path)

    # Config NAT Pool
    elif func == 'patch_openconfig_nat_nat_instances_instance_nat_pool_nat_pool_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-pool/nat-pool-entry={poolname}/config', id=args[0],poolname=args[1])
        ip = args[2].split("-")
        if len(ip) == 1:
            body = { "openconfig-nat:config": {"ip-address": args[2]} }
        else:
            body =  { "openconfig-nat:config": {"ip-address-range": args[2]} }

        if args[3] != "":
            body["openconfig-nat:config"].update( {"nat-port": args[3] } )
        return api.patch(path, body)

    # Remove all NAPT Static basic translation entries
    elif func == 'delete_openconfig_nat_nat_instances_instance_napt_mapping_table':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/napt-mapping-table', id=args[0])
        return api.delete(path)

    # Remove NAT Pool
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_pool_nat_pool_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-pool/nat-pool-entry={poolname}/config',id=args[0],poolname=args[1])
        return api.delete(path)

    # Remove all NAT Pools
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_pool':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-pool', id=args[0])
        return api.delete(path)

    # Config NAT Binding
    elif func == 'patch_openconfig_nat_nat_instances_instance_nat_acl_pool_binding_nat_acl_pool_binding_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-acl-pool-binding/nat-acl-pool-binding-entry={name}/config', id=args[0],name=args[1])
        body = { "openconfig-nat:config": {"access-list": args[2], "nat-pool": args[3]} }
        l = len(args)
        if l >= 5:
            body["openconfig-nat:config"].update( {"nat-type": nat_type_map[args[4]] } )
        if l == 6:
            body["openconfig-nat:config"].update( {"twice-nat-id": args[5]} )
        return api.patch(path, body)

    # Remove NAT Binding
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_acl_pool_binding_nat_acl_pool_binding_entry_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-acl-pool-binding/nat-acl-pool-binding-entry={name}/config', id=args[0],name=args[1])
        return api.delete(path)

    # Remove all NAT Bindings
    elif func == 'delete_openconfig_nat_nat_instances_instance_nat_acl_pool_binding':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-acl-pool-binding', id=args[0])
        return api.delete(path)

    # Config NAT Zone
    elif func == 'patch_openconfig_interfaces_ext_interfaces_interface_nat_zone_config_nat_zone':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-interfaces-ext:nat-zone/config/nat-zone', name=args[1])
        body = { "openconfig-interfaces-ext:nat-zone": int(args[2]) }
        return api.patch(path, body)

    # Remove NAT Zone
    elif func == 'delete_openconfig_interfaces_ext_interfaces_interface_nat_zone_config_nat_zone':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-interfaces-ext:nat-zone/config/nat-zone', name=args[1])
        return api.delete(path)


    # Get NAT Global Config
    elif func == 'get_openconfig_nat_nat_instances_instance_config':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config', id=args[0])
        return api.get(path)

    # Get NAT Bindings
    elif func == 'get_openconfig_nat_nat_instances_instance_nat_acl_pool_binding':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-acl-pool-binding', id=args[0])
        return api.get(path)

    # Get NAT Pools
    elif func == 'get_openconfig_nat_nat_instances_instance_nat_pool':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/nat-pool', id=args[0])
        return api.get(path)

    else:
        return api.cli_not_implemented(func)

def run(func, args):   

    try:
       args.insert(0,"0")  # NAT instance 0
       response = invoke_api(func, args)    
       if response.ok():
           if response.content is not None:
               # Get Command Output
               api_response = response.content
               show_cli_output(args[1], api_response)
       else:
           print response.error_message()

        
    except Exception as e:
        print("Failure: %s\n" %(e))

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

