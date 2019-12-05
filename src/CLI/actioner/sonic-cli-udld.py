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

    # show udld global
    if func == 'get_sonic_udld_sonic_udld_udld_udld_list':
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD/UDLD_LIST={id}', id='GLOBAL')
        resp = aa.get(keypath)
        if not resp.ok() and resp.status_code == 404:
            resp.set_error_message('UDLD not configured')
        return resp

    # show udld neighbors
    if func == 'get_list_sonic_udld_sonic_udld_udld_port_neigh_table_udld_port_neigh_table_list_zz':
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT_NEIGH_TABLE/UDLD_PORT_NEIGH_TABLE_LIST')
        return aa.get(keypath)

    # show udld interface <ifname>
    if func == 'get_sonic_udld_sonic_udld_udld_port_neigh_table_udld_port_neigh_table_list':
        return generateShowUdldInterfaceResponse(aa, args)

    # show udld statistics
    if func == 'get_list_sonic_udld_sonic_udld_udld_port_table_udld_port_table_list_zz':
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT_TABLE/UDLD_PORT_TABLE_LIST')
        return aa.get(keypath)

    # show udld statistics interface <ifname>
    if func == 'get_sonic_udld_sonic_udld_udld_port_table_udld_port_table_list_zz':
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT_TABLE/UDLD_PORT_TABLE_LIST={ifname}', ifname=args[1])
        return aa.get(keypath)

    # Enable UDLD global
    if func == 'post_list_sonic_udld_sonic_udld_udld_udld_list' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD')
        body=collections.defaultdict(dict)
        body["UDLD_LIST"] = [{
            "id": "GLOBAL",
            "msg_time": 1,
            "multiplier":3,
            "admin_enable":True,
            "aggressive":False
            }]
        return aa.post(keypath, body)

    # Disable UDLD global
    if func == 'delete_sonic_udld_sonic_udld_udld' :
        # Delete all udld config tables including port and global level.
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld')
        return aa.delete(keypath)

    # Configure UDLD aggressive
    if func == 'patch_sonic_udld_sonic_udld_udld_udld_list_aggressive' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD/UDLD_LIST={id}/aggressive', id='GLOBAL')
        body = { "sonic-udld:aggressive": str2bool(args[0]) }
        return aa.patch(keypath, body)

    # Configure UDLD message-time
    if func == 'patch_sonic_udld_sonic_udld_udld_udld_list_msg_time' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD/UDLD_LIST={id}/msg_time', id='GLOBAL')
        if args[0] == '0':
            body = { "sonic-udld:msg_time": 1 }
        else:
            body = { "sonic-udld:msg_time": int(args[0]) }
        return aa.patch(keypath, body)

    # Configure UDLD multiplier
    if func == 'patch_sonic_udld_sonic_udld_udld_udld_list_multiplier' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD/UDLD_LIST={id}/multiplier', id='GLOBAL')
        if args[0] == '0':
            body = { "sonic-udld:multiplier": 3 }
        else:
            body = { "sonic-udld:multiplier": int(args[0]) }
        return aa.patch(keypath, body)

    # Enable UDLD at Interface
    if func == 'post_list_sonic_udld_sonic_udld_udld_port_udld_port_list' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT')
        body=collections.defaultdict(dict)
        body["UDLD_PORT_LIST"] = [{
            "ifname": args[0],
            "admin_enable":True,
            "aggressive":False
            }]
        return aa.post(keypath, body)

    # Disable UDLD at Interface
    if func == 'delete_sonic_udld_sonic_udld_udld_port_udld_port_list' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT/UDLD_PORT_LIST={ifname}', ifname=args[0])
        return aa.delete(keypath)

    # Configure UDLD aggressive at Interface
    if func == 'patch_sonic_udld_sonic_udld_udld_port_udld_port_list_aggressive' :
        keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT/UDLD_PORT_LIST={ifname}/aggressive', ifname=args[1])
        body = { "sonic-udld:aggressive": str2bool(args[0]) }
        return aa.patch(keypath, body)

    # enable/disable debug udld at global level
    if func == 'udldGlobalDebugHandler' :
        return udldGlobalDebugHandler(args)

    # enable/disable debug udld at interface level
    if func == 'udldInterfaceDebugHandler' :
        return udldInterfaceDebugHandler(args)

    # clear udld statistics 
    if func == 'udldInterfaceCountersClearHandler' :
        return udldInterfaceCountersClearHandler(args)

def generateShowUdldInterfaceResponse(clientApi, args):
    resp_status = 0
    # Retrieve global UDLD message time
    keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD/UDLD_LIST={id}/msg_time', id='GLOBAL')
    resp = clientApi.get(keypath)
    gbl_msg_time = 0
    if resp.ok() and 'sonic-udld:msg_time' in resp.content.keys():
        gbl_msg_time = resp.content['sonic-udld:msg_time']

    # Retrieve port level UDLD configs
    keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT/UDLD_PORT_LIST={ifname}', ifname=args[1])
    resp = clientApi.get(keypath)
    port_conf_dict = {}
    if resp.ok() and 'sonic-udld:UDLD_PORT_LIST' in resp.content.keys():
        port_conf_dict = resp.content['sonic-udld:UDLD_PORT_LIST'][0]
        resp_status = resp.response.status_code
    else:
        if resp.status_code == 404:
            resp.set_error_message('UDLD not configured on ' + args[1])
        return resp

    # Retrieve UDLD Global operatioal info
    keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_GLOBAL_TABLE/UDLD_GLOBAL_TABLE_LIST={id}', id='GLOBAL')
    resp = clientApi.get(keypath)
    gbl_oper_dict = {}
    if resp.ok() and 'sonic-udld:UDLD_GLOBAL_TABLE_LIST' in resp.content.keys():
        gbl_oper_dict = resp.content['sonic-udld:UDLD_GLOBAL_TABLE_LIST'][0]

    # Retrieve UDLD port local status
    keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT_TABLE/UDLD_PORT_TABLE_LIST={ifname}/status', ifname=args[1])
    resp = clientApi.get(keypath)
    port_status = 0
    if resp.ok() and 'sonic-udld:status' in resp.content.keys():
        port_status = resp.content['sonic-udld:status']

    # Retrieve neighbors info for a given interface
    keypath = cc.Path('/restconf/data/sonic-udld:sonic-udld/UDLD_PORT_NEIGH_TABLE/UDLD_PORT_NEIGH_TABLE_LIST={ifname},{index}', ifname=args[1], index='*')
    resp = clientApi.get(keypath)
    neigh_dict = {}
    if resp.ok() and 'sonic-udld:UDLD_PORT_NEIGH_TABLE_LIST' in resp.content.keys():
        neigh_dict = resp.content['sonic-udld:UDLD_PORT_NEIGH_TABLE_LIST'][0]
    else:
        if resp.status_code != 404:
            return resp

    final_dict = {}
    final_dict['interface'] = args[1]
    final_dict['msg_time'] = gbl_msg_time
    final_dict['status'] = port_status
    final_dict['port_config'] = port_conf_dict
    final_dict['global_oper'] = gbl_oper_dict
    final_dict['neighbor'] = neigh_dict

    body=collections.defaultdict(dict)
    body['intf_show'] = final_dict

    import requests
    presp = requests.Response()
    presp.status_code = resp_status
    response = cc.Response(presp)
    response.content = body

    return response


def udldGlobalDebugHandler(args):
    if args[0] == '1':
        print("Enabled Debug at global level")
    else:
        print("Disable Debug at global level")


def udldInterfaceDebugHandler(args):
    if args[0] == '1':
        print("Enabled Debug at interface level for " + args[1])
    else:
        print("Disable Debug at interface level for " + args[1])


def udldInterfaceCountersClearHandler(args):
    if len(args) == 0:
        print("Clearing counters for all interfaces")
    else:
        print("Clearing counters for interface: " + args[0])


def str2bool(s):
    return s.lower() in ("yes", "true", "t", "1")


def run(func, args):
        api_response = invoke(func, args)
        # Temporary for Mock CLI. Needs to be removed
        if api_response is None:
            value = [{'id': 'GLOBAL'}]
            show_cli_output(args[0], value)
            return
        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            elif 'sonic-udld:UDLD_LIST' in response.keys():
                value = response['sonic-udld:UDLD_LIST']
                if value is None:
                    return
                else:
                    show_cli_output(args[0], value)
            elif 'sonic-udld:UDLD_PORT_TABLE_LIST' in response.keys():
                value = response['sonic-udld:UDLD_PORT_TABLE_LIST']
                if value is None:
                    return
                else:
                    show_cli_output(args[0], value)
            elif 'intf_show' in response.keys():
                value = response['intf_show']
                if value is None:
                    return
                else:
                    show_cli_output(args[0], value)
            elif 'sonic-udld:UDLD_PORT_NEIGH_TABLE_LIST' in response.keys():
                value = response['sonic-udld:UDLD_PORT_NEIGH_TABLE_LIST']
                if value is None:
                    return
                else:
                    show_cli_output(args[0], value)
        else:
            print(api_response.error_message())


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

