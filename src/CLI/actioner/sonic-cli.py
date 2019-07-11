#!/usr/bin/python
import sys
import time
import json
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
       forwarding_action = "ACCEPT" if args[4] == 'permit' else 'DROP'
       if args[3] == 'ip' : 
          protocol = "IP_PROTOCOL"
       elif args[3] == 'icmp':
	  protocol = "IP_ICMP"
       elif args[4] == "6":
	  protocol = "IP_TCP";
       elif args[4] == "17":
	  protocol = "IP_UDP";
       elif args[4] == "1":
	  protocol = "IP_ICMP";
       elif args[4] == "2":
	  protocol = "IP_IGMP";
       elif args[4] == "103":
	  protocol = "IP_PIM";
       elif args[4] == "46":
	  protocol = "IP_RSVP";
       elif args[4] == "47":
	  protocol = "IP_GRE";
       elif args[4] == "51":
	  protocol = "IP_AUTH";
       elif args[4] == "115":
	  protocol = "IP_L2TP";
       elif args[4] == 'tcp' :
          protocol = "IP_TCP"
       else :
          protocol = "IP_UDP"
       if (len(args) <= 7):
            body =  { "openconfig-acl:acl-entry": [ {
                        "sequence-id": int(args[2]),
                        "config": {
                            "sequence-id": int(args[2]),
                        },
                        "ipv4": {
                            "config": {
                                "source-address": args[5],
                                "destination-address": args[6],
                                "protocol": protocol 
                            }
                        },
                        "actions": {
                            "config": {
                                "forwarding-action": forwarding_action 
                            }
                        }
                        } ] }
       else:
            body =  { "acl-entry": [ {
                        "sequence-id": int(args[2]),
                        "config": {
                            "sequence-id": int(args[2]),
                        },
                        "ipv4": {
                            "config": {
                                "source-address": args[5],
                                "destination-address": args[7],
                                "protocol": protocol
                            }
                        },
                        "transport": {
                            "config": {
                                "source-port": int(args[6]),
                                "destination-port": int(args[8])
                            }
                        },
                        "actions": {
                            "config": {
                                "forwarding-action": forwarding_action
                            }
                        }
                        } ] }

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
       body = json.dumps(body,ensure_ascii=False)
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
