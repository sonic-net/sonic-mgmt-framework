#!/usr/bin/python
import sys
import time
import json
import collections
import re
import ast
import swagger_client
from swagger_client.rest import ApiException
from scripts.render_cli import show_cli_output


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
    if func.__name__ == 'get_acl_acl_sets':
       keypath = []

    # Get Interface binding to ACL table info
    elif func.__name__ == 'get_acl_interfaces':
       keypath = []

    # Get all the rules specific to an ACL table.
    elif func.__name__ == 'get_acl_set_acl_entries':
       keypath = [ args[0], args[1] ]

    # Configure ACL table
    elif func.__name__ == 'patch_acl_acl_sets_acl_set' :
       keypath = [ args[0], args[1] ]
       body = { "openconfig-acl:config": {
                   "name": args[0],
                   "type": args[1],
                   "description": ""
                 }
              }

    # Configure ACL rule specific to an ACL table
    elif func.__name__ == 'post_list_base_acl_entries_acl_entry' :
       	keypath = [ args[0], args[1] ]
        forwarding_action = "ACCEPT" if args[3] == 'permit' else 'DROP'
        proto_number = {"icmp":"IP_ICMP","tcp":"IP_TCP","udp":"IP_UDP","6":"IP_TCP","17":"IP_UDP","1":"IP_ICMP",
                       "2":"IP_IGMP","103":"IP_PIM","46":"IP_RSVP","47":"IP_GRE","51":"IP_AUTH","115":"IP_L2TP"}
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
	flags_list=[]
        i=6
        while(i<len(args)):
            if args[i] == 'src-port-eq':
                i+=1
                body["acl-entry"][0]["transport"]["config"]["source-port"]=args[i]

            if re_ip.match(args[i]):
                body["acl-entry"][0]["ipv4"]["config"]["destination-address"]=args[i]
            
            if args[i] == 'dst-port-eq':
                i+=1
                body["acl-entry"][0]["transport"]["config"]["destination-port"]=args[i]

	    if args[i] == 'dscp':
        	i+=1
        	body["acl-entry"][0]["ipv4"]["config"]["dscp"]=int(args[i])

            if "tcp_" in args[i]: 
                body["acl-entry"][0]["transport"]["config"]["tcp-flags"]=flags_list.append(args[i]) 
            i+=1
	
    # Add the ACL table binding to an Interface(Ingress / Egress).
    elif func.__name__ == 'post_list_base_interfaces_interface':
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
    elif func.__name__ == 'delete_interface_ingress_acl_sets_ingress_acl_set':
        keypath = [args[0], args[1], args[2]]

    # Remove the ACL table binding to an Egress interface.
    elif func.__name__ == 'delete_interface_egress_acl_sets_egress_acl_set':
        keypath = [args[0], args[1], args[2]]

    # Remove all the rules and delete the ACL table.
    elif func.__name__ == 'delete_acl_acl_sets_acl_set':
        keypath = [args[0], args[1]]
    elif func.__name__ == 'delete_acl_set_acl_entries_acl_entry':
        keypath = [args[0], args[1], args[2]]
    else:
       body = {} 
    if body is not None: 
       body = json.dumps(body,ensure_ascii=False, indent=4, separators=(',', ': '))
       print body
       return keypath, ast.literal_eval(body)
    else:
       return keypath,body

def run(func, args):

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(swagger_client.OpenconfigAclApi(),func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(swagger_client.OpenconfigAclApi(),func.__name__)(*keypath)
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
        print("Exception when calling OpenconfigAclApi->%s : %s\n" %(func.__name__, e))

if __name__ == '__main__':

    #pdb.set_trace()
    func = eval(sys.argv[1], globals(), swagger_client.OpenconfigAclApi.__dict__)
    run(func, sys.argv[2:])
