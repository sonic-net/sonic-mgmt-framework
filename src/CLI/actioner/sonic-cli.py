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
import collections
import re
import ast
import openconfig_acl_client
from rpipe_utils import pipestr
from openconfig_acl_client.rest import ApiException
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
    # Get the rules of all ACL table entries.
    if func.__name__ == 'get_openconfig_acl_acl_acl_sets':
       keypath = []

    # Get Interface binding to ACL table info
    elif func.__name__ == 'get_openconfig_acl_acl_interfaces':
       keypath = []

    # Get all the rules specific to an ACL table.
    elif func.__name__ == 'get_openconfig_acl_acl_acl_sets_acl_set_acl_entries':
       keypath = [ args[0], args[1] ]

    # Configure ACL table
    elif func.__name__ == 'patch_openconfig_acl_acl_acl_sets_acl_set' :
        keypath = [ args[0], args[1] ]
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
    # Configure ACL rule specific to an ACL table
    elif func.__name__ == 'patch_list_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry' :
       	keypath = [ args[0], args[1] ]
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

    # Add the ACL table binding to an Interface(Ingress / Egress).
    elif func.__name__ == 'patch_list_openconfig_acl_acl_interfaces_interface':
        keypath = []
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
    # Remove the ACL table binding to an Ingress interface.
    elif func.__name__ == 'delete_openconfig_acl_acl_interfaces_interface_ingress_acl_sets_ingress_acl_set':
        keypath = [args[0], args[1], args[2]]

    # Remove the ACL table binding to an Egress interface.
    elif func.__name__ == 'delete_openconfig_acl_acl_interfaces_interface_egress_acl_sets_egress_acl_set':
        keypath = [args[0], args[1], args[2]]

    # Remove all the rules and delete the ACL table.
    elif func.__name__ == 'delete_openconfig_acl_acl_acl_sets_acl_set':
        keypath = [args[0], args[1]]
    elif func.__name__ == 'delete_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry':
        keypath = [args[0], args[1], args[2]]
    else:
       body = {}
    if body is not None:
       body = json.dumps(body,ensure_ascii=False, indent=4, separators=(',', ': '))
       return keypath, ast.literal_eval(body)
    else:
       return keypath,body

def run(func, args):

    c = openconfig_acl_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_acl_client.OpenconfigAclApi(api_client=openconfig_acl_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            response = api_response.to_dict()
            if 'openconfig_aclacl_entry' in response.keys():
                value = response['openconfig_aclacl_entry']
                if value is None:
                    print("Success")
                else:
                    print ("Failed")
            elif 'openconfig_aclacl_set' in response.keys():
                value = response['openconfig_aclacl_set']
                if value is None:
                    print("Success")
                else:
                    print("Failed")
            elif 'openconfig_aclacl_entries' in response.keys():
                value = response['openconfig_aclacl_entries']
                if value is None:
                    return
                show_cli_output(args[2], value)
            elif 'openconfig_aclacl_sets' in response.keys():
                value = response['openconfig_aclacl_sets']
                if value is None:
                    return
                show_cli_output(args[0], value)
            elif 'openconfig_aclinterfaces' in response.keys():
                value = response['openconfig_aclinterfaces']
                if value is None:
                    return
                show_cli_output(args[0], value)
            else:
                print("Failed")

    except ApiException as e:
        #print("Exception when calling OpenconfigAclApi->%s : %s\n" %(func.__name__, e))
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


if __name__ == '__main__':

    pipestr().write(sys.argv)
    #pdb.set_trace()
    func = eval(sys.argv[1], globals(), openconfig_acl_client.OpenconfigAclApi.__dict__)
    run(func, sys.argv[2:])
