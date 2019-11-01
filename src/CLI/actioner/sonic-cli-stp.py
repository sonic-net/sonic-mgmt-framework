#!/usr/bin/python

###########################################################################
#
# Copyright 2019 Dell, Inc.
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
import openconfig_spanning_tree_client
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from openconfig_spanning_tree_client.rest import ApiException
import urllib3

urllib3.disable_warnings()

cust_to_oc_URI  = {
  "PVST":{
    'get_custom_stp': 'get_openconfig_spanning_tree_ext_stp_pvst',
    'get_custom_stp_vlan': 'get_openconfig_spanning_tree_ext_stp_pvst_vlan',
    'get_custom_stp_vlan_interfaces_interface':
    'get_openconfig_spanning_tree_ext_stp_pvst_vlan',
    'get_custom_stp_counters': 'get_openconfig_spanning_tree_ext_stp_pvst',
    'get_custom_stp_counters_vlan': 'get_openconfig_spanning_tree_ext_stp_pvst_vlan',
    'get_custom_stp_inconsistentports': 'get_openconfig_spanning_tree_ext_stp_pvst',
    'get_custom_stp_inconsistentports_vlan': 'get_openconfig_spanning_tree_ext_stp_pvst_vlan',
    'patch_custom_stp_vlan_config_enable':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_spanning_tree_enable',
    'patch_custom_stp_vlan_config_forwarding_delay':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_forwarding_delay',
    'patch_custom_stp_vlan_config_hello_time':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_hello_time',
    'patch_custom_stp_vlan_config_max_age':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_max_age',
    'patch_custom_stp_vlan_config_bridge_priority':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_bridge_priority',
    'delete_custom_stp_vlan_config_forwarding_delay':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_forwarding_delay',
    'delete_custom_stp_vlan_config_hello_time':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_hello_time',
    'delete_custom_stp_vlan_config_max_age':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_max_age',
    'delete_custom_stp_vlan_config_bridge_priority':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_bridge_priority',
    'patch_custom_stp_vlan_interfaces_interface_config_cost':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_cost',
    'patch_custom_stp_vlan_interfaces_interface_config_port_priority':
    'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_port_priority',
    'delete_custom_stp_vlan_interfaces_interface_config_cost':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_cost',
    'delete_custom_stp_vlan_interfaces_interface_config_port_priority':
    'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_port_priority'
                          ''
  },
  "RAPID_PVST":{
    'get_custom_stp': 'get_openconfig_spanning_tree_stp_rapid_pvst',
    'get_custom_stp_vlan': 'get_openconfig_spanning_tree_stp_rapid_pvst_vlan',
    'get_custom_stp_vlan_interfaces_interface':
    'get_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface',
    'get_custom_stp_counters': 'get_openconfig_spanning_tree_stp_rapid_pvst',
    'get_custom_stp_counters_vlan': 'get_openconfig_spanning_tree_stp_rapid_pvst_vlan',
    'get_custom_stp_inconsistentports': 'get_openconfig_spanning_tree_stp_rapid_pvst',
    'get_custom_stp_inconsistentports_vlan': 'get_openconfig_spanning_tree_stp_rapi_pvst_vlan',
    'patch_custom_stp_vlan_config_enable':
    'patch_openconfig_spanning_tree_ext_stp_rapid_pvst_vlan_config_spanning_tree_enable',
    'patch_custom_stp_vlan_config_forwarding_delay':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_forwarding_delay',
    'patch_custom_stp_vlan_config_hello_time':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_hello_time',
    'patch_custom_stp_vlan_config_max_age':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_max_age',
    'patch_custom_stp_vlan_config_bridge_priority':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_bridge_priority',
    'delete_custom_stp_vlan_config_forwarding_delay':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_forwarding_delay',
    'delete_custom_stp_vlan_config_hello_time':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_hello_time',
    'delete_custom_stp_vlan_config_max_age':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_max_age',
    'delete_custom_stp_vlan_config_bridge_priority':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_bridge_priority',
    'patch_custom_stp_vlan_interfaces_interface_config_cost':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_cost',
    'patch_custom_stp_vlan_interfaces_interface_config_port_priority':
    'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_port_priority',
    'delete_custom_stp_vlan_interfaces_interface_config_cost':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_cost',
    'delete_custom_stp_vlan_interfaces_interface_config_port_priority':
    'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_port_priority'
    }
}

def generate_body(func, args):
    body = None
    keypath = []

    if func.__name__ == 'get_openconfig_spanning_tree_stp_rapid_pvst' or func.__name__ == 'get_openconfig_spanning_tree_ext_stp_pvst':
        keypath = []
    elif func.__name__ == 'get_openconfig_spanning_tree_stp_rapid_pvst_vlan' or func.__name__ == 'get_openconfig_spanning_tree_ext_stp_pvst_vlan':
       if (len(args) > 2):
        keypath = [ int(args[2]) ]
    elif func.__name__ == 'get_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface' or func.__name__ == 'get_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface':
       if (len(args) > 2):
        keypath = [ int(args[2]), args[3] ]
    elif func.__name__ == 'get_openconfig_spanning_tree_stp_interfaces':
        keypath = []
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_spanning_tree_enable' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_rapid_pvst_vlan_config_spanning_tree_enable':
        keypath = [ int(args[1]) ]
        if (len(args) > 2):
            if args[2] == "True":
                body = { "openconfig-spanning-tree-ext:spanning-tree-enable": True }
            elif args[2] == "False":
                body = { "openconfig-spanning-tree-ext:spanning-tree-enable": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_hello_time' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_hello_time':
       keypath = [ int(args[1]) ]
       body = { "openconfig-spanning-tree:hello-time": int(args[2]) }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_hello_time' or func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_hello_time':
       keypath = [ int(args[1]) ]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_forwarding_delay' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_forwarding_delay':
       keypath = [ int(args[1]) ]
       body = { "openconfig-spanning-tree:forwarding-delay": int(args[2]) }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_forwarding_delay' or func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_forwarding_delay':
       keypath = [ int(args[1]) ]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_max_age' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_max_age':
       keypath = [ int(args[1]) ]
       body = { "openconfig-spanning-tree:max-age": int(args[2]) }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_max_age' or func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_max_age':
       keypath = [ int(args[1]) ]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_bridge_priority' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_config_bridge_priority':
       keypath = [ int(args[1]) ]
       body = { "openconfig-spanning-tree:bridge-priority": int(args[2]) }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_config_bridge_priority' or func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_config_bridge_priority':
       keypath = [ int(args[1]) ]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_global_config_enabled_protocol':
       keypath = []
       if (len(args) > 1):
          if args[1] == 'pvst':
             body = { "openconfig-spanning-tree:enabled-protocol": ['PVST'] }
          else:
             body = { "openconfig-spanning-tree:enabled-protocol": ['RAPID_PVST'] }
       else:
          body = { "openconfig-spanning-tree:enabled-protocol": ['PVST'] }
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_global_config_bpdu_filter':
       if args[1] == "True":
            body = { "openconfig-spanning-tree:bpdu-filter": True }
       elif args[1] == "False":
            body = { "openconfig-spanning-tree:bpdu-filter": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_spanning_tree_enable':
       keypath = [args[1]]
       if args[2] == "True":
            body = { "openconfig-spanning-tree-ext:spanning-tree-enable": True }
       elif args[2] == "False":
            body = { "openconfig-spanning-tree-ext:spanning-tree-enable": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_uplink_fast':
       keypath = [args[1]]
       if args[2] == "True":
            body = { "openconfig-spanning-tree-ext:uplink-fast": True }
       elif args[2] == "False":
            body = { "openconfig-spanning-tree-ext:uplink-fast": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_portfast':
       keypath = [args[1]]
       if args[2] == "True":
            body = { "openconfig-spanning-tree-ext:portfast": True }
       elif args[2] == "False":
            body = { "openconfig-spanning-tree-ext:portfast": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_interfaces_interface_config_bpdu_filter':
       keypath = [args[1]]
       if args[2] == "True":
            body = { "openconfig-spanning-tree:bpdu-filter": True }
       elif args[2] == "False":
            body = { "openconfig-spanning-tree:bpdu-filter": False }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_interfaces_interface_config_bpdu_filter':
       keypath = [args[1]]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_interfaces_interface_config_bpdu_guard':
       keypath = [args[1]]
       if args[2] == "True":
            body = { "openconfig-spanning-tree:bpdu-guard": True }
       elif args[2] == "False":
            body = { "openconfig-spanning-tree:bpdu-guard": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_bpdu_guard_port_shutdown':
       keypath = [args[1]]
       if args[2] == "True":
          body = { "openconfig-spanning-tree-ext:bpdu-guard-port-shutdown": True }
       elif args[2] == "False":
          body = { "openconfig-spanning-tree-ext:bpdu-guard-port-shutdown": False }
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_interfaces_interface_config_guard':
       keypath = [args[1]]
       body = { "openconfig-spanning-tree:guard": args[2] }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_cost':
       keypath = [args[1]]
       body = { "openconfig-spanning-tree-ext:cost": int(args[2]) }
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_interfaces_interface_config_port_priority':
       keypath = [args[1]]
       body = { "openconfig-spanning-tree-ext:port-priority": int(args[2]) }
    elif func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_interfaces_interface_config_cost':
       keypath = [args[1]]
    elif func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_interfaces_interface_config_port_priority':
       keypath = [args[1]]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_interfaces_interface_config_link_type':
       keypath = [args[1]]
       body = { "openconfig-spanning-tree:link-type": args[2] }
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_interfaces_interface_config_link_type':
       keypath = [args[1]]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_interfaces_interface_config_edge_port':
       keypath = [args[1]]
       if args[2] == "True":
           body = { "openconfig-spanning-tree:edge-port": "EDGE_ENABLE" }
       elif args[2] == "False":
           body = { "openconfig-spanning-tree:edge-port": "EDGE_DISABLE" }
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_cost' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_cost':
       keypath = [int(args[1]), args[2]]
       body = { "openconfig-spanning-tree:cost": int(args[3])}
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_cost' or func.__name__== 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_cost':
       keypath = [int(args[1]), args[2]]
    elif func.__name__ == 'patch_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_port_priority' or func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_port_priority':
       keypath = [int(args[1]), args[2]]
       body = { "openconfig-spanning-tree:port-priority": int(args[3])}
    elif func.__name__ == 'delete_openconfig_spanning_tree_stp_rapid_pvst_vlan_interfaces_interface_config_port_priority' or func.__name__ == 'delete_openconfig_spanning_tree_ext_stp_pvst_vlan_interfaces_interface_config_port_priority':
       keypath = [int(args[1]), args[2]]
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_global_config_forwarding_delay':
       body = {"openconfig-spanning-tree-ext:forwarding-delay": int(args[1])}
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_global_config_hello_time':
       body = {"openconfig-spanning-tree-ext:hello-time": int(args[1])}
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_global_config_max_age':
       body = {"openconfig-spanning-tree-ext:max-age": int(args[1])}
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_global_config_bridge_priority':
       body = {"openconfig-spanning-tree-ext:bridge-priority": int(args[1])}
    elif func.__name__ == 'patch_openconfig_spanning_tree_ext_stp_global_config_rootguard_timeout':
       body = {"openconfig-spanning-tree-ext:rootguard-timeout": int(args[1])}

    return keypath,body

def getId(item):
    return item['vlan-id']

def stp_mode_get(aa):
    keypath = []
    stp_mode = None
    stp_resp = None
    try:
        stp_resp = getattr(aa,"get_openconfig_spanning_tree_stp_global_config_enabled_protocol")(*keypath)
        if not stp_resp:
           print (" Failed to get STP mode")
           return stp_resp,stp_mode

        stp_resp = aa.api_client.sanitize_for_serialization(stp_resp)

        if stp_resp['openconfig-spanning-tree:enabled-protocol'][0] == "openconfig-spanning-tree-ext:PVST":
           stp_mode = "PVST"
        elif stp_resp['openconfig-spanning-tree:enabled-protocol'][0] == "openconfig-spanning-tree-types:RAPID_PVST":
           stp_mode = "RAPID_PVST"

    except ApiException as e:
        error_body = json.loads(e.body)
        if "ietf-restconf:errors" in error_body and 'error' in error_body["ietf-restconf:errors"]:
            error = error_body["ietf-restconf:errors"]["error"][0]
            if "error-message" in error and error["error-message"]:
                print "%Error: "+ error["error-message"] + " or STP not enabled"
                return stp_resp,stp_mode
            
        print "%Error: Transaction Failure"

    return stp_resp,stp_mode

def run(args):

    c = openconfig_spanning_tree_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_spanning_tree_client.OpenconfigSpanningTreeApi(api_client=openconfig_spanning_tree_client.ApiClient(configuration=c))

    oc_func = None
    if "custom" in args[0]:
        stp_resp, stp_mode = stp_mode_get(aa)
        if stp_mode is None:
            return;
        oc_func = cust_to_oc_URI[stp_mode][args[0]]
    else:
        oc_func = args[0]

    if oc_func is None:
        return

    func = eval(oc_func, globals(), openconfig_spanning_tree_client.OpenconfigSpanningTreeApi.__dict__)

    # create a body block
    keypath, body = generate_body(func, args)
    api_response = None

    try:
        if body is not None:
            api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
            api_response = getattr(aa,func.__name__)(*keypath)

        api_response = aa.api_client.sanitize_for_serialization(api_response)

        if not api_response:
            if "get_" not in func.__name__:
                print "Success"
            return
        else:
            if args[0] == 'get_custom_stp' or args[0] == 'get_custom_stp_counters' or args[0] == 'get_custom_stp_inconsistentports':
                # Sort based on VLAN id
                if stp_mode == "RAPID_PVST":
                   vlan_list = api_response['openconfig-spanning-tree:rapid-pvst']
                else:
                   vlan_list = api_response['openconfig-spanning-tree-ext:pvst']

                if 'vlan' in vlan_list:
                    tup = vlan_list['vlan']
                    vlan_list['vlan'] = sorted(tup, key=getId)

            if args[0] == 'get_custom_stp' or args[0] == 'get_custom_stp_vlan' or args[0] == 'get_custom_stp_vlan_interfaces_interface':
                # add stp mode/protocols to the response
                api_response.update(stp_resp)
                keypath = []
                # add stp interfaces to the response
                if args[0] == 'get_custom_stp_vlan_interfaces_interface':
                    if (len(args) > 2):
                        keypath = [ args[3] ]
                        stp_intf_response = getattr(aa,"get_openconfig_spanning_tree_stp_interfaces_interface")(*keypath)
                else:
                    stp_intf_response = getattr(aa,"get_openconfig_spanning_tree_stp_interfaces")(*keypath)

                stp_intf_response = aa.api_client.sanitize_for_serialization(stp_intf_response)
                if stp_intf_response:
                    api_response.update(stp_intf_response)

            elif args[0] == 'get_custom_stp_counters' or args[0] == 'get_custom_stp_counters_vlan': 
                # add stp mode/protocols to the response
                api_response.update(stp_resp)
            elif args[0] == 'get_custom_stp_inconsistentports' or args[0] == 'get_custom_stp_inconsistentports_vlan': 
                # add stp mode/protocols to the response
                api_response.update(stp_resp)
                keypath = []
                # add stp interfaces to the response
                stp_global_response = getattr(aa,"get_openconfig_spanning_tree_stp_global_config")(*keypath)
                stp_global_response = aa.api_client.sanitize_for_serialization(stp_global_response)
                api_response.update(stp_global_response)

            show_cli_output(args[1], api_response)
            return

    except ApiException as e:
        error_body = json.loads(e.body)
        if "ietf-restconf:errors" in error_body and 'error' in error_body["ietf-restconf:errors"]:
            error = error_body["ietf-restconf:errors"]["error"][0]
            if "error-message" in error and error["error-message"]:
                print "%Error: "+ error["error-message"]
                return
            
        print "%Error: Transaction Failure"

    return

if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1:])
