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

    # Get the rules of all ACL table entries.
    if func == 'get_openconfig_acl_acl_acl_sets':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets')
        return aa.get(keypath)

    # Get Interface binding to ACL table info
    if func == 'get_openconfig_acl_acl_interfaces':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/interfaces')
        return aa.get(keypath)

    # Get all the rules specific to an ACL table.
    if func == 'get_openconfig_acl_acl_acl_sets_acl_set_acl_entries':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}/acl-entries',
                name=args[0], type=args[1] )
        return aa.get(keypath)

    # Configure ACL table
    if func == 'patch_openconfig_acl_acl_acl_sets_acl_set' :
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}',
                name=args[0], type=args[1] )
        body=collections.defaultdict(dict)
        body["acl-set"]=[{
                        "name": args[0],
                        "type": args[1],
                       "config": {
                                "name": args[0],
                                "type": args[1],
                                "description": ""
                       }
                    }]

        return aa.patch(keypath, body)

    # Configure ACL rule specific to an ACL table
    if func == 'patch_list_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry' :
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}/acl-entries/acl-entry',
                name=args[0], type=args[1] )
        forwarding_action = "ACCEPT" if args[3] == 'permit' else 'DROP'
        proto_number = {"icmp":"IP_ICMP","tcp":"IP_TCP","udp":"IP_UDP","6":"IP_TCP","17":"IP_UDP","1":"IP_ICMP",
                       "2":"IP_IGMP","103":"IP_PIM","46":"IP_RSVP","47":"IP_GRE","51":"IP_AUTH","115":"IP_L2TP"}
        if args[4] not in proto_number.keys():
            print("%Error: Invalid protocol number")
            exit(1)
        else:
            protocol = proto_number.get(args[4])
        body=collections.defaultdict(dict)
        body["acl-entry"]=[{
                       "sequence-id": int(args[2]),
                       "config": {
                                "sequence-id": int(args[2])
                       },
                       "ipv4":{
                           "config":{
                               "protocol": protocol
                           }
                       },
                       "transport": {
                           "config": {
                           }
                       },
                       "actions": {
                           "config": {
                               "forwarding-action": forwarding_action
                           }
                       }
                   }]
        re_ip = re.compile("^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}")
        if re_ip.match(args[5]):
            body["acl-entry"][0]["ipv4"]["config"]["source-address"]=args[5]
        elif args[5]=="any":
            body["acl-entry"][0]["ipv4"]["config"]["source-address"]="0.0.0.0/0"
        flags_list=[]
        i=6
        while(i<len(args)):
            if args[i] == 'src-eq':
                i+=1
                body["acl-entry"][0]["transport"]["config"]["source-port"]=int(args[i])

            if re_ip.match(args[i]):
                body["acl-entry"][0]["ipv4"]["config"]["destination-address"]=args[i]

            if args[i]=="any":
                body["acl-entry"][0]["ipv4"]["config"]["destination-address"]="0.0.0.0/0"

            if args[i] == 'dst-eq':
                i+=1
                body["acl-entry"][0]["transport"]["config"]["destination-port"]=int(args[i])

            if args[i] == 'dscp':
                i+=1
                body["acl-entry"][0]["ipv4"]["config"]["dscp"]=int(args[i])

            if args[i] in ['fin','syn','ack','urg','rst','psh']:
                args[i]=("tcp_"+args[i]).upper()
                flags_list.append(args[i])
                body["acl-entry"][0]["transport"]["config"]["tcp-flags"]=flags_list
            i+=1

        return aa.patch(keypath, body)

    # Add the ACL table binding to an Interface(Ingress / Egress).
    if func == 'patch_list_openconfig_acl_acl_interfaces_interface':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/interfaces/interface')
        if args[3] == "ingress":
            body = { "openconfig-acl:interface": [ {
                        "id": args[2],
                        "config": {
                            "id": args[2]
                        },
                        "interface-ref": {
                            "config": {
                                "interface": args[2]
                            }
                        },
                        "ingress-acl-sets": {
                            "ingress-acl-set": [
                            {
                                "set-name": args[0],
                                "type": args[1],
                                "config": {
                                    "set-name": args[0],
                                    "type": args[1]
                                }
                            } ] }
                    } ] }
        else:
            body = { "interface": [ {
                        "id": args[2],
                        "config": {
                            "id": args[2]
                        },
                        "interface-ref": {
                            "config": {
                                "interface": args[2]
                            }
                        },
                        "egress-acl-sets": {
                            "egress-acl-set": [
                            {
                                "set-name": args[0],
                                "type": args[1],
                                "config": {
                                    "set-name": args[0],
                                    "type": args[1]
                                }
                            } ] }
                    } ] }

        return aa.patch(keypath, body)

    # Remove the ACL table binding to an Ingress interface.
    if func == 'delete_openconfig_acl_acl_interfaces_interface_ingress_acl_sets_ingress_acl_set':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/interfaces/interface={id}/ingress-acl-sets/ingress-acl-set={set_name},{type}',
                id=args[0], set_name=args[1], type=args[2] )
        return aa.delete(keypath)

    # Remove the ACL table binding to an Egress interface.
    if func == 'delete_openconfig_acl_acl_interfaces_interface_egress_acl_sets_egress_acl_set':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/interfaces/interface={id}/egress-acl-sets/egress-acl-set={set_name},{type}',
                id=args[0], set_name=args[1], type=args[2] )
        return aa.delete(keypath)

    # Remove all the rules and delete the ACL table.
    if func == 'delete_openconfig_acl_acl_acl_sets_acl_set':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}',
                name=args[0], type=args[1] )
        return aa.delete(keypath)

    # Remove a rule from ACL
    if func == 'delete_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry':
        keypath = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}/acl-entries/acl-entry={sequence_id}',
                name=args[0], type=args[1], sequence_id=args[2] )
        return aa.delete(keypath)

    else:
        print("%Error: not implemented")
        exit(1)

def run(func, args):
    try:
        api_response = invoke(func, args)

        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            elif 'openconfig-acl:acl-entry' in response.keys():
                value = response['openconfig-acl:acl-entry']
                if value is None:
                    print("Success")
                else:
                    print ("Failed")
            elif 'openconfig-acl:acl-set' in response.keys():
                value = response['openconfig-acl:acl-set']
                if value is None:
                    print("Success")
                else:
                    print("Failed")
            elif 'openconfig-acl:acl-entries' in response.keys():
                value = response['openconfig-acl:acl-entries']
                if value is None:
                    return
                show_cli_output(args[2], value)
            elif 'openconfig-acl:acl-sets' in response.keys():
                value = response['openconfig-acl:acl-sets']
                if value is None:
                    return
                show_cli_output(args[0], value)
            elif 'openconfig-acl:interfaces' in response.keys():
                value = response['openconfig-acl:interfaces']
                if value is None:
                    return
                show_cli_output(args[0], value)

        else:
            #error response
            print api_response.error_message()

    except:
            # system/network error
            print "%Error: Transaction Failure"


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

