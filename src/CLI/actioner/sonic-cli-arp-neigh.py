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
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

def get_keypath(func,args):
    keypath = None
    instance = None

    if func == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors':
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/neighbors', name=args[1], index="0")
    elif func == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors':
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/neighbors', name=args[1], index="0")
    elif func == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors_neighbor':
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv4/neighbors/neighbor={ip}', name=args[1], index="0", ip=args[3])
    elif func == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors_neighbor':
        keypath = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/subinterfaces/subinterface={index}/openconfig-if-ip:ipv6/neighbors/neighbor={ip}',name=args[1], index="0", ip=args[3])
    elif func == 'get_sonic_neigh_sonic_neigh_neigh_table':
        keypath = cc.Path('/restconf/data/sonic-neigh:sonic-neigh/NEIGH_TABLE')
    return keypath

def fdb_call(macAddr, vlanName):
    aa = cc.ApiClient()

    vlanId = vlanName[len("Vlan"):]
    macAddr = macAddr.strip()

    keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/fdb/mac-table/entries/entry={macaddress},{vlan}', name='default', macaddress=macAddr, vlan=vlanId)

    try:
        response = aa.get(keypath)
        response = response.content

        if 'openconfig-network-instance:entry' in response.keys():
                instance = response['openconfig-network-instance:entry'][0]['interface']['interface-ref']['state']['interface']

        if instance is not None:
                return instance
        return "-"

    except:
        return "-"

def run(func, args):
    aa = cc.ApiClient()

    # create a body block
    keypath = get_keypath(func, args)
    neigh_list = []

    try:
        response = aa.get(keypath)
        response  = response.content

        if 'openconfig-if-ip:neighbor' in response.keys():
                ext_intf_name = "-"
                neigh = response['openconfig-if-ip:neighbor']

                if neigh[0]['state'] is None:
                    return

                ipAddr = neigh[0]['state']['ip']
                if ipAddr is None:
                    return

                macAddr = neigh[0]['state']['link-layer-address']
                if macAddr is None:
                    return

                if args[1].startswith('Vlan'):
                        ext_intf_name = fdb_call(macAddr, args[1])

                neigh_table_entry = {'ipAddr':ipAddr,
                                    'macAddr':macAddr,
                                    'intfName':args[1],
                                    'extIntfName':ext_intf_name
                                  }
                neigh_list.append(neigh_table_entry)

        elif 'openconfig-if-ip:neighbors' in response.keys():
                if response['openconfig-if-ip:neighbors'] is None:
                    return

                neighs = response['openconfig-if-ip:neighbors']['neighbor']
                if neighs is None:
                    return

                for neigh in neighs:
                    ext_intf_name = "-"
                    ipAddr = neigh['state']['ip']
                    if ipAddr is None:
                        return

                    macAddr = neigh['state']['link-layer-address']
                    if macAddr is None:
                        return

                    if args[1].startswith('Vlan'):
                        ext_intf_name = fdb_call(macAddr, args[1])

                    neigh_table_entry = {'ipAddr':ipAddr,
                                    'macAddr':macAddr,
                                    'intfName':args[1],
                                    'extIntfName':ext_intf_name
                                  }
                neigh_list.append(neigh_table_entry)

        elif 'sonic-neigh:NEIGH_TABLE' in response.keys():
                if response['sonic-neigh:NEIGH_TABLE'] is None:
                    return

                neighs = response['sonic-neigh:NEIGH_TABLE']['NEIGH_TABLE_LIST']
                if neighs is None:
                    return

                for neigh in neighs:
                        ext_intf_name = "-"

                        family = neigh['family']
                        if family is None:
                            return
                        if family != args[1]:
                            continue

                        ifName = neigh['ifname']
                        if ifName is None:
                            return

                        ipAddr = neigh['ip']
                        if ipAddr is None:
                            return

                        macAddr = neigh['neigh']
                        if macAddr is None:
                            return

                        if ifName.startswith('Vlan'):
                            ext_intf_name = fdb_call(macAddr, ifName)

                        neigh_table_entry = {'ipAddr':ipAddr,
                                    'macAddr':macAddr,
                                    'intfName':ifName,
                                    'extIntfName':ext_intf_name
                                  }

                        if (len(args) == 3):
                                if (args[2] == "mac" and args[3] == macAddr):
                                    neigh_list.append(neigh_table_entry)
                                elif (args[1] == "IPv4" or args[1] == "IPv6") and args[2] == ipAddr:
                                    neigh_list.append(neigh_table_entry)
                        else:
                            neigh_list.append(neigh_table_entry)

        show_cli_output(args[0],neigh_list)
        return

    except:
        # system/network error
        print "Error: Transaction Failure"

if __name__ == '__main__':
    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])
