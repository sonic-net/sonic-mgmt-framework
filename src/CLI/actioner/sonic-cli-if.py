#!/usr/bin/python
import sys
import time
import json
import ast
import openconfig_interfaces_client
from openconfig_interfaces_client.rest import ApiException
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
    if func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_description':
       keypath = [ args[0] ]
       body = { "openconfig-interfaces:description": args[1] }
    elif func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_enabled':
       keypath = [ args[0] ]
       if args[1] == "True":
           body = { "openconfig-interfaces:enabled": True }
       else:
           body = { "openconfig-interfaces:enabled": False }
    elif func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_mtu':
       keypath = [ args[0] ]
       body = { "openconfig-interfaces:mtu":  int(args[1]) }
    elif func.__name__ == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config':
       sp = args[1].split('/')
       keypath = [ args[0], 0, sp[0] ]
       body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
    elif func.__name__ == 'patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config':
       sp = args[1].split('/')
       keypath = [ args[0], 0, sp[0] ]
       body = { "openconfig-if-ip:config":  {"ip" : sp[0], "prefix-length" : int(sp[1])} }
    elif func.__name__ == 'get_openconfig_interfaces_interfaces_interface':
	keypath = [args[0]]
    elif func.__name__ == 'get_openconfig_interfaces_interfaces':
        keypath = []
    else:
       body = {} 

    return keypath,body

def run(func, args):

    c = openconfig_interfaces_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_interfaces_client.OpenconfigInterfacesApi(api_client=openconfig_interfaces_client.ApiClient(configuration=c))

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
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            if api_response is None:
                print("Failed")
            else:
                if func.__name__ == 'get_openconfig_interfaces_interfaces_interface':
                     show_cli_output(args[1], api_response)
                elif func.__name__ == 'get_openconfig_interfaces_interfaces':
                     show_cli_output(args[0], api_response)
                else:
                     return
    except ApiException as e:
        #print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func.__name__, e))
        if e.body != "":
            print "% Error:", e.body
        else:
            print "Failed"

if __name__ == '__main__':

    func = eval(sys.argv[1], globals(), openconfig_interfaces_client.OpenconfigInterfacesApi.__dict__)

    run(func, sys.argv[2:])
