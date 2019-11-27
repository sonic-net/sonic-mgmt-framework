#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom.  The term "Broadcom" refers to Broadcom Inc. and/or
# its subsidiaries.
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
import pdb
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output

vxlan_global_info = []

def invoke(func, args):
    body = None
    aa = cc.ApiClient()


    #[un]configure VTEP 
    if (func == 'patch_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list' or
        func == 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list'):
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL/VXLAN_TUNNEL_LIST={name}', name=args[0][6:])

        if (func.startswith("patch") is True):
            body = {
              "sonic-vxlan:VXLAN_TUNNEL_LIST": [
                {
                  "name": args[0][6:],
                  "src_ip": args[1] 
                }
              ]
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #[un]configure EVPN NVO
    if (func == 'patch_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list' or
        func == 'delete_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list'):
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/EVPN_NVO/EVPN_NVO_LIST={name}', name=args[0][4:])

        if (func.startswith("patch") is True):
            body = {
              "sonic-vxlan:EVPN_NVO_LIST": [
                {
                  "name": args[0][4:],
                  "source_vtep": args[1] 
                }
              ]
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #[un]configure Tunnel Map
    if (func == 'patch_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list' or
        func == 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list'):
        fail = 0
        for count in range(int(args[3])):
          vidstr = str(int(args[2]) + count)
          vnid = int(args[1]) + count
          vnistr = str(vnid)
          mapname = 'map_'+ vnistr + '_' + vidstr
          keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL_MAP/VXLAN_TUNNEL_MAP_LIST={name},{mapname1}', name=args[0][6:], mapname1=mapname)

          if (func.startswith("patch") is True):
              body = {
                "sonic-vxlan:VXLAN_TUNNEL_MAP_LIST": [
                  {
                    "name": args[0][6:],
                    "mapname": mapname,
                    "vlan": 'Vlan' + vidstr,
                    "vni": vnid 
                  }
                ]
              }
              api_response =  aa.patch(keypath, body)
          else:
              api_response = aa.delete(keypath)

          if api_response.ok():
              response = api_response.content
              if response is None:
                  result = "Success"
              elif 'sonic-vxlan:sonic-vxlan' in response.keys():
                  value = response['sonic-vxlan:sonic-vxlan']
                  if value is None:
                      result = "Success"
                  else:
                      result = "Failed"
              
          else:
              #error response
              result =  "Failed"
              fail = 1
              #print(api_response.error_message())
              if (func.startswith("patch") is True):
                print("Error:Map creation for VID:{} failed. Verify if the VLAN is created".format(vidstr)) 
              else:
                print ("Error:Map deletion for VID:{} failed with error = {}".format(vidstr,api_response.error_message()[7:]))

        if fail == 0:
          if (func.startswith("patch") is True):
            print("Map creation for {} vids succeeded.".format(count+1))
          else:
            print("Map deletion for {} vids succeeded.".format(count+1))

        return api_response
          
    if func == "get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list":
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL/VXLAN_TUNNEL_LIST')
        return aa.get(keypath)

    if func == "get_list_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list":
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/EVPN_NVO/EVPN_NVO_LIST')
        return aa.get(keypath)

    if func == "get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list":
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL_MAP/VXLAN_TUNNEL_MAP_LIST')
        return aa.get(keypath)

    if func == "get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_table_vxlan_tunnel_table_list":
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/VXLAN_TUNNEL_TABLE/VXLAN_TUNNEL_TABLE_LIST')
        return aa.get(keypath)

    if func == "get_list_sonic_vxlan_sonic_vxlan_evpn_remote_vni_table_evpn_remote_vni_table_list":
        keypath = cc.Path('/restconf/data/sonic-vxlan:sonic-vxlan/EVPN_REMOTE_VNI_TABLE/EVPN_REMOTE_VNI_TABLE_LIST')
        return aa.get(keypath)


    else:
        print("Error: not implemented")
        exit(1)

    return api_response

#show vxlan interface 
def vxlan_show_vxlan_interface(args):

    print("VXLAN INFO:")
    api_response = invoke("get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list", args)
    if api_response.ok():
        response = api_response.content
	if response is None:
	    print("no vxlan configuration")
	elif response is not None:
           tunnel_list = response['sonic-vxlan:VXLAN_TUNNEL_LIST']
           print("VTEP Name : " + tunnel_list[0]['name'])
           print("VTEP Source IP : " + tunnel_list[0]['src_ip'])
	       #show_cli_output(args[0], vxlan_info)
	#print(api_response.error_message())

    api_response = invoke("get_list_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list", args)
    if api_response.ok():
        response = api_response.content

	if response is None:
	    print("no evpn configuration")
	elif response is not None:
           nvo_list = response['sonic-vxlan:EVPN_NVO_LIST']
           print("EVPN NVO Name : " + nvo_list[0]['name'])
           print("EVPN VTEP : " + nvo_list[0]['source_vtep'])
	       #show_cli_output(args[0], vxlan_info)

    return

#show vxlan map 
def vxlan_show_vxlan_vlanvnimap(args):

    print("VLAN-VNI Mapping")
    print("{0:^8}  {1:^8}".format('VLAN','VNI'))
    api_response = invoke("get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list", args)
    if api_response.ok():
        response = api_response.content
	if response is None:
	    print("no vxlan configuration")
	elif response is not None:
           tunnel_list = response['sonic-vxlan:VXLAN_TUNNEL_MAP_LIST']
           for iter in tunnel_list:
             print("{0:^8}  {1:^8}".format(iter['vlan'],iter['vni']))
	       #show_cli_output(args[0], vxlan_info)
	#print(api_response.error_message())

    return

#show vxlan tunnel 
def vxlan_show_vxlan_tunnel(args):

    print("{:*^70s}".format("List of Tunnels"))
    print("{0:^20} {1:^15} {2:^15} {3:^8} {4:^12}".format('Name','SIP','DIP','source','operstatus'))
    api_response = invoke("get_list_sonic_vxlan_sonic_vxlan_vxlan_tunnel_table_vxlan_tunnel_table_list", args)
    if api_response.ok():
        response = api_response.content
	if response is None:
	    print("no vxlan configuration")
	elif response is not None:
           tunnel_list = response['sonic-vxlan:VXLAN_TUNNEL_TABLE_LIST']
           for iter in tunnel_list:
             print("{0:^20} {1:^15} {2:^15} {3:^8} {4:^12}".format(iter['name'],iter['src_ip'],iter['dst_ip'],iter['tnl_src'],iter['operstatus']))
	       #show_cli_output(args[0], vxlan_info)
	#print(api_response.error_message())

    return

#show vxlan evpn remote vni
def vxlan_show_vxlan_evpn_remote_vni(args):
    arg_length = len(args);
    print("{0:^20} {1:^15} {2:^10}".format('Vlan', 'Tunnel', 'VNI'))
    api_response = invoke("get_list_sonic_vxlan_sonic_vxlan_evpn_remote_vni_table_evpn_remote_vni_table_list", args)
    if api_response.ok():
        response = api_response.content
	if response is None:
	    print("no vxlan evpn remote vni entires")
	elif response is not None:
           tunnel_vni_list = response['sonic-vxlan:EVPN_REMOTE_VNI_TABLE_LIST']
           for iter in tunnel_vni_list:
               if (arg_length == 1) or (arg_length == 2 and args[1] == iter['ip_addr']):
                   print("{0:^20} {1:^15} {2:^10}".format(iter['vlan'], iter['ip_addr'], iter['vni']))
    return

def run(func, args):

    #show commands
    try:
        #show vxlan brief command
        if func == 'show vxlan interface':
            vxlan_show_vxlan_interface(args)
            return
        if func == 'show vxlan vlanvnimap':
            vxlan_show_vxlan_vlanvnimap(args)
            return
        if func == 'show vxlan tunnel':
            vxlan_show_vxlan_tunnel(args)
            return
        if func == 'show vxlan evpn remote vni':
            vxlan_show_vxlan_evpn_remote_vni(args)
            return

    except Exception as e:
            print(sys.exc_value)
            return


    #config commands
    try:
        api_response = invoke(func, args)

        if (func != 'patch_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list' and
            func != 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel_map_vxlan_tunnel_map_list'):
          if api_response.ok():
              response = api_response.content
              if response is None:
                  print("Success")
              elif 'sonic-vxlan:sonic-vxlan' in response.keys():
                  value = response['sonic-vxlan:sonic-vxlan']
                  if value is None:
                      print("Success")
                  else:
                      print ("Failed")
              
          else:
              #error response
              #print(api_response.error_message())
              if func == 'patch_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list':
                 print("Error : Only a single VTEP is supported.")
              if func == 'delete_sonic_vxlan_sonic_vxlan_vxlan_tunnel_vxlan_tunnel_list':
                 print("Error : Remove all VLAN-VNI mappings and also the EVPN NVO object .")
              if func == 'patch_sonic_vxlan_sonic_vxlan_evpn_nvo_evpn_nvo_list':
                 print("Error : Verify if the source vtep is configured and that this is the only EVPN object")

    except:
            # system/network error
            print("Error: Transaction Failure")


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

