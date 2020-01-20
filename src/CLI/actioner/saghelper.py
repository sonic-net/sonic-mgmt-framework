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
import os
#import json
#import ast
import subprocess

def get_netdev_all_sag_ifa(afi):
    output = subprocess.check_output("ip addr show type macvlan", shell=True)
    result = {}
    for row in output.split('\n'):
        row = row.lstrip()
        if len(row.split(" ")) < 2:
            continue
        if row.split(" ")[1].startswith("sag"):
            iface = row.split(" ")[1].split("@")[0]
        if row.split(" ")[0] == "inet6":
            if len(row.split(" ")) < 4:
                continue
            address = row.split(" ")[1]
            if iface.startswith('sag'):
                if iface not in result:
                    result[iface] = []
                    result[iface].append(address)
                else:
                    result[iface].append(address)
        else:
            if len(row.split(" ")) != 5:
                continue
            address = row.split(" ")[1]
            iface = row.split(" ")[4]
            if iface.startswith('sag'):
                if iface not in result:
                    result[iface] = []
                    result[iface].append(address)
                else:
                    result[iface].append(address)
    return result

# get_if_master
#
# Given an interface name, return its master netdev from kernel
#
def get_if_master(iface):
    master_file = "/sys/class/net/{0}/master"
    if os.path.exists(master_file.format(iface)) == False:
        return ""

    try:
        master_name = os.readlink(master_file.format(iface))
    except IOError as e:
        #print "Error: unable to read from file: %s" % str(e)
        return ""

    master_name = master_name.strip("../")
    return master_name

#
# get_if_oper_state
#
# Given an interface name, return its oper state reported by the kernel.
#
def get_if_oper_state(iface):
    oper_file = "/sys/class/net/{0}/carrier"

    if os.path.exists(oper_file.format(iface)) == False:
        return "down"

    try:
        state_file = open(oper_file.format(iface), "r")
    except IOError as e:
        print "Error: unable to open file: %s" % str(e)
        return "error"

    try:
        oper_state = state_file.readline().rstrip()
    except IOError as e:
        return "down"

    if oper_state == "1":
        return "up"
    else:
        return "down"

def get_sag_netdev(intf):
    vlan_id = intf.lstrip('Vlan')
    iface = 'sag'+vlan_id+'.256'
    return iface

def get_if_master_and_oper(in_data):
    data = in_data
    output = {}
    sag_ifa4 = get_netdev_all_sag_ifa('IPv4')
    for item in data:
        if "ifname" in item:
            if item["ifname"] not in output:
                output[item["ifname"]] = {}
            output[item["ifname"]]["master"] = get_if_master(get_sag_netdev(item["ifname"]))
            oper_state = get_if_oper_state(get_sag_netdev(item["ifname"]))
            for ip in item['gwip']:
                output[item["ifname"]][ip] = {}
                output[item["ifname"]][ip]["oper_state"] = oper_state
                if get_sag_netdev(item["ifname"]) in  sag_ifa4 and ip in sag_ifa4[get_sag_netdev(item["ifname"])]:
                    output[item["ifname"]][ip]["admin_state"] = "up"
                else:
                    output[item["ifname"]][ip]["admin_state"] = "down"

    return output

def run(intf_dict):
    output = get_if_master_and_oper(intf_dict)
    return output
