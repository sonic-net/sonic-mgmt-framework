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
from netaddr import *
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

lag_type_map = {"active" : "LACP", "on": "STATIC"}

def filter_address(d, isIPv4):
    if d is None:
        return
    if 'sonic-mgmt-interface:MGMT_INTF_TABLE_IPADDR_LIST' in d:
        ipData = d['sonic-mgmt-interface:MGMT_INTF_TABLE_IPADDR_LIST']
        newIpData = []
        for l in ipData:
            for k, v in l.items():
               if k == "ipPrefix":
                  ip = IPNetwork(v)
                  if isIPv4:
                      if ip.version == 4:
                          newIpData.append(l)
                  else:
                      if ip.version == 6:
                          newIpData.append(l)
        del ipData[:]
        ipData.extend(newIpData)

    if 'sonic-interface:INTF_TABLE_IPADDR_LIST' in d:
        ipData = d['sonic-interface:INTF_TABLE_IPADDR_LIST']
        newIpData = []
        for l in ipData:
            for k, v in l.items():
               if k == "ipPrefix":
                  ip = IPNetwork(v)
                  if isIPv4:
                      if ip.version == 4:
                          newIpData.append(l)
                  else:
                      if ip.version == 6:
                          newIpData.append(l)
        del ipData[:]
        ipData.extend(newIpData)

def invoke_api(func, args=[]):
    api = cc.ApiClient()

    # handle interfaces using the 'switch' mode
    if func == 'if_config':
        if args[0] == 'phy-if-name' or args[0] == 'vlan-if-name':
            body = { "openconfig-interfaces:config": { "name": args[1] }}
            path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/config', name=args[1])
            return api.patch(path, body)

    # Create interface
    if func == 'patch_openconfig_interfaces_interfaces_interface':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}', name=args[0])
        return api.patch(path)
        
    elif func == 'patch_openconfig_interfaces_interfaces_interface_config':
        body = { "openconfig-interfaces:config": { "name": args[0] }}
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/config', name=args[0])
        return api.patch(path, body)

    #Configure PortChannel
    elif func == 'portchannel_config':
        body ={
                 "openconfig-interfaces:interface": [{
                                                      "name": args[0],
                                                      "config": {"name": args[0]},
                                                      "openconfig-if-aggregate:aggregation" : {"config": {}}
                                                    }]
               }

        # Configure lag type (active/on)
        mode = args[1].split("=")[1]
        if mode != "" :
            body["openconfig-interfaces:interface"][0]["openconfig-if-aggregate:aggregation"]["config"].update( {"lag-type": lag_type_map[mode] } )

        # Configure Min links
        links = args[2].split("=")[1]
        if links != "":
            body["openconfig-interfaces:interface"][0]["openconfig-if-aggregate:aggregation"]["config"].update( {"min-links": int(links) } )

        # Configure Fallback
        fallback = args[3].split("=")[1]
        if fallback != "":
            body["openconfig-interfaces:interface"][0]["openconfig-if-aggregate:aggregation"]["config"].update( {"openconfig-interfaces-ext:fallback": True} )

        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}', name=args[0])
        return api.patch(path, body)

    
    # Delete interface
    elif func == 'delete_openconfig_interfaces_interfaces_interface':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}', name=args[0])
        return api.delete(path)
    
    #Configure description
    elif func == 'patch_openconfig_interfaces_interfaces_interface_config_description':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/config/description', name=args[0])
        body = { "openconfig-interfaces:description": args[1] }
        return api.patch(path,body)
        
    # Enable or diable interface
    elif func == 'patch_openconfig_interfaces_interfaces_interface_config_enabled':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/config/enabled', name=args[0])
        if args[1] == "True":
           body = { "openconfig-interfaces:enabled": True }
        else:
           body = { "openconfig-interfaces:enabled": False }
        return api.patch(path, body)
        
    # Configure MTU
    elif func == 'patch_openconfig_interfaces_interfaces_interface_config_mtu':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/config/mtu', name=args[0])
        body = { "openconfig-interfaces:mtu":  int(args[1]) }
        return api.patch(path, body)
        
    elif func == 'patch_openconfig_if_ethernet_interfaces_interface_ethernet_config_auto_negotiate':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/config/auto-negotiate', name=args[0])
        if args[1] == "true":
            body = { "openconfig-if-ethernet:auto-negotiate": True }
        else :
            body = { "openconfig-if-ethernet:auto-negotiate": False }
        return api.patch(path, body)
        
    elif func == 'patch_openconfig_if_ethernet_interfaces_interface_ethernet_config_port_speed':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/config/port-speed', name=args[0])
        speed_map = {"10MBPS":"SPEED_10MB", "100MBPS":"SPEED_100MB", "1GIGE":"SPEED_1GB", "auto":"SPEED_1GB" }
        if args[1] not in speed_map.keys():
            print("%Error: Invalid port speed config")
            exit(1)
        else:
            speed = speed_map.get(args[1])

        body = { "openconfig-if-ethernet:port-speed": speed }
        return api.patch(path, body)
    
    elif func == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config':
        sp = args[1].split('/')
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/addresses/address={ip}/config', name=args[0], index="0", ip=sp[0])
        if len(args) > 2 and len(args[2]) > 0:
            body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1]), "openconfig-interfaces-ext:gw-addr": args[2]} }
        else:
            body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
        return api.patch(path, body)    
    
    elif func == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config':
        sp = args[1].split('/')
    
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/addresses/address={ip}/config', name=args[0], index="0", ip=sp[0])
        if len(args) > 2 and len(args[2]) > 0:
            body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1]), "openconfig-interfaces-ext:gw-addr": args[2]} }
        else:
            body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
        return api.patch(path, body)
        
    elif func == 'patch_openconfig_vlan_interfaces_interface_ethernet_switched_vlan_config':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/config', name=args[0])
        if args[1] == "ACCESS":
           body = {"openconfig-vlan:config": {"interface-mode": "ACCESS","access-vlan": int(args[2])}}
        else:
           vlanlst = args[2].split(',')
           vlanlst = [sub.replace('-', '..') for sub in vlanlst]
           body = {"openconfig-vlan:config": {"interface-mode": "TRUNK","trunk-vlans": [int(i) if '..' not in i else i for i in vlanlst]}}

        return api.patch(path, body)
        
    elif func == 'patch_openconfig_vlan_interfaces_interface_aggregation_switched_vlan_config':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/config', name=args[0])
        if args[1] == "ACCESS":
           body = {"openconfig-vlan:config": {"interface-mode": "ACCESS","access-vlan": int(args[2])}}
        else:
           vlanlst = args[2].split(',')
           vlanlst = [sub.replace('-', '..') for sub in vlanlst]
           body = {"openconfig-vlan:config": {"interface-mode": "TRUNK","trunk-vlans": [int(i) if '..' not in i else i for i in vlanlst]}}
        return api.patch(path, body)
        
    elif func == 'delete_openconfig_vlan_interfaces_interface_ethernet_switched_vlan_config_access_vlan':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/config/access-vlan', name=args[0])
        return api.delete(path)
    elif func == 'delete_openconfig_vlan_interfaces_interface_aggregation_switched_vlan_config_access_vlan':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/config/access-vlan', name=args[0])
        return api.delete(path)

    elif func == 'del_llist_openconfig_vlan_interfaces_interface_ethernet_switched_vlan_config_trunk_vlans':
        vlanStr = args[2].replace('-', '..')
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/openconfig-vlan:switched-vlan/config/trunk-vlans={trunk}', name=args[0], trunk=vlanStr)
        return api.delete(path)

    elif func == 'del_llist_openconfig_vlan_interfaces_interface_aggregation_switched_vlan_config_trunk_vlans':
        vlanStr = args[2].replace('-', '..')
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-aggregate:aggregation/openconfig-vlan:switched-vlan/config/trunk-vlans={trunk}', name=args[0], trunk=vlanStr)
        return api.delete(path)

    elif func == 'delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config_prefix_length':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/addresses/address={ip}/config/prefix-length', name=args[0], index="0", ip=args[1])
        return api.delete(path)
        
    elif func == 'delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config_prefix_length':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/addresses/address={ip}/config/prefix-length', name=args[0], index="0", ip=args[1])
        return api.delete(path)
        
    elif func == 'get_openconfig_interfaces_interfaces_interface':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}', name=args[0])
        return api.get(path)

    elif func == 'get_sonic_port_sonic_port_port_table':
        path = cc.Path('/restconf/data/sonic-port:sonic-port/PORT_TABLE')
        return api.get(path)

    elif func == 'get_openconfig_interfaces_interfaces':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces')
        return api.get(path)
    elif func == 'ip_interfaces_get' or func == 'ip6_interfaces_get':
        d = {}

        path = cc.Path('/restconf/data/sonic-mgmt-interface:sonic-mgmt-interface/MGMT_INTF_TABLE/MGMT_INTF_TABLE_IPADDR_LIST')
        responseMgmtIntfTbl = api.get(path)
        if responseMgmtIntfTbl.ok():
            d.update(responseMgmtIntfTbl.content)
            if func == 'ip_interfaces_get':
               filter_address(d, True)
            else:
               filter_address(d, False)

        path = cc.Path('/restconf/data/sonic-interface:sonic-interface/INTF_TABLE/INTF_TABLE_IPADDR_LIST')
        responseIntfTbl = api.get(path)
        if responseIntfTbl.ok():
            d.update(responseIntfTbl.content)
            if func == 'ip_interfaces_get':
                filter_address(d, True)
            else:
                filter_address(d, False)

        path = cc.Path('/restconf/data/sonic-port:sonic-port/PORT_TABLE/PORT_TABLE_LIST')
        responsePortTbl = api.get(path)
        if responsePortTbl.ok():
            d.update(responsePortTbl.content)

        path = cc.Path('/restconf/data/sonic-portchannel:sonic-portchannel/LAG_TABLE/LAG_TABLE_LIST')
        responseLagTbl = api.get(path)
        if responseLagTbl.ok():
            d.update(responseLagTbl.content)

        path = cc.Path('/restconf/data/sonic-vlan:sonic-vlan/VLAN_TABLE/VLAN_TABLE_LIST')
        responseVlanTbl =  api.get(path)
        if responseVlanTbl.ok():
            d.update(responseVlanTbl.content)
        return d
        
    # Add members to port-channel
    elif func == 'patch_openconfig_if_aggregate_interfaces_interface_ethernet_config_aggregate_id':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/config/openconfig-if-aggregate:aggregate-id', name=args[0])
        body = { "openconfig-if-aggregate:aggregate-id": args[1] }
        return api.patch(path, body)
    
    # Remove members from port-channel
    elif func == 'delete_openconfig_if_aggregate_interfaces_interface_ethernet_config_aggregate_id':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-ethernet:ethernet/config/openconfig-if-aggregate:aggregate-id', name=args[0])
        return api.delete(path)

    # Config min-links for port-channel
    elif func == 'patch_openconfig_if_aggregate_interfaces_interface_aggregation_config_min_links':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-aggregate:aggregation/config/min-links', name=args[0])
        body = { "openconfig-if-aggregate:min-links": int(args[1]) }
        return api.patch(path, body)
    
    # Config fallback mode for port-channel    
    elif func == 'patch_dell_intf_augments_interfaces_interface_aggregation_config_fallback':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/openconfig-if-aggregate:aggregation/config/openconfig-interfaces-ext:fallback', name=args[0])
        if args[1] == "True":
            body = { "openconfig-interfaces-ext:fallback": True }
        else :
            body = { "openconfig-interfaces-ext:fallback": False }
        return api.patch(path, body)

    # Config IPv4 Unnumbered interface
    elif func == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_unnumbered_interface_ref_config_interface':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/unnumbered/interface-ref/config/interface', name=args[0], index="0")

        body = { "openconfig-if-ip:interface" : args[1] }
        return api.patch(path, body)    
    elif func == 'delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_unnumbered_interface_ref_config_interface':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/unnumbered/interface-ref/config/interface', name=args[0], index="0")
        return api.delete(path)    

    return api.cli_not_implemented(func)




def getId(item):
    prfx = "Ethernet"
    state_dict = item['state']
    ifName = state_dict['name']

    if ifName.startswith(prfx):
        ifId = int(ifName[len(prfx):])
        return ifId
    return ifName

def getSonicId(item):

    prfx = "Ethernet"
    state_dict = item
    ifName = state_dict['ifname']
    if ifName.startswith(prfx):
        ifId = int(ifName[len(prfx):])
        return ifId
    return ifName

def run(func, args):   

    try:
        response = invoke_api(func, args)    
        if func == 'ip_interfaces_get' or func == 'ip6_interfaces_get':
            show_cli_output(args[0], response)
            return
        if response.ok():
          if response.content is not None:
            # Get Command Output
            api_response = response.content
            if 'openconfig-interfaces:interfaces' in api_response:
                value = api_response['openconfig-interfaces:interfaces']
                if 'interface' in value:
                    tup = value['interface']
                    value['interface'] = sorted(tup, key=getId)
            elif 'sonic-port:PORT_TABLE' in api_response:
                value = api_response['sonic-port:PORT_TABLE']
                if 'PORT_TABLE_LIST' in value:
                    tup = value['PORT_TABLE_LIST']
                    value['PORT_TABLE_LIST'] =  sorted(tup, key=getSonicId)

            if api_response is None:
                print("Failed")
            else:
                if func == 'get_openconfig_interfaces_interfaces_interface':
                    show_cli_output(args[1], api_response)
                elif func == 'get_openconfig_interfaces_interfaces':
                    show_cli_output(args[0], api_response)
                elif func == 'get_sonic_port_sonic_port_port_table':
                    show_cli_output(args[0], api_response)
                else:
                    return
        else:
            print response.error_message()

        
    except Exception as e:
        print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func, e))

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

