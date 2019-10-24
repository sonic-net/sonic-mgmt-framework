#!/usr/bin/python
import sys
import time
import json
import ast
import sonic_vxlan_client
from rpipe_utils import pipestr
from sonic_vxlan_client.rest import ApiException
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
    if func.__name__ == 'patch_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list':
       keypath = []
       body = {
         "sonic-vxlan:VXLAN_TUNNEL_LIST": [
           {
             "name": args[0][5:],
             "src_ip": args[1] 
           }
         ]
       }
    elif func.__name__ == 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel':
       #keypath = [ args[0][5:] ]
       keypath = []
    elif func.__name__ == 'patch_list_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list':
       keypath = []
       body = {
         "sonic-vxlan:EVPN_NVO_LIST": [
           {
             "name": args[0][4:],
             "source_vtep": args[1] 
           }
         ]
       }
    elif func.__name__ == 'delete_sonic_vxlan_sonic_vxlan_evpn_nvo':
       #keypath = [ args[0][4:] ]
       keypath = []
    elif func.__name__ == 'patch_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list':
       keypath = []
       body = {
         "sonic-vxlan:VXLAN_TUNNEL_MAP_LIST": [
           {
             "name": args[0][5:],
             "mapname": 'map_'+ args[1] + '_' + args[2],
             "vlan": 'Vlan' + args[2],
             "vni": int(args[1]) 
           }
         ]
       }
    elif func.__name__ == 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list':
       keypath = [ args[0][5:] , 'map_'+ args[1] + '_' + args[2]]
    else:
       body = {}

    return keypath,body

def getId(item):
    prfx = "Ethernet"
    state_dict = item['state']
    ifName = state_dict['name']

    if ifName.startswith(prfx):
        ifId = int(ifName[len(prfx):])
        return ifId
    return ifName

def run(func, args):

    c = sonic_vxlan_client.Configuration()
    c.verify_ssl = False
    aa = sonic_vxlan_client.SonicVxlanApi(api_client=sonic_vxlan_client.ApiClient(configuration=c))

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
#           if 'openconfig-interfaces:interfaces' in api_response:
#               value = api_response['openconfig-interfaces:interfaces']
#               if 'interface' in value:
#                   tup = value['interface']
#                   value['interface'] = sorted(tup, key=getId)

            if api_response is None:
                print("Failed")
            else:
                return

    except ApiException as e:
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
        else:
            print "Failed"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), sonic_vxlan_client.SonicVxlanApi.__dict__)

    run(func, sys.argv[2:])
