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
import ast
import openconfig_network_instance_client 
from rpipe_utils import pipestr
from openconfig_network_instance_client.rest import ApiException
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

IDENTIFIER='BGP'
NAME1='bgp'

plugins = dict()

def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func


def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    keypath = []
    body = None
    # Get the rules of all ACL table entries.
    if func.__name__ == 'patch_sonic_bgp_global_sonic_bgp_global_bgp_globals_bgp_globals_list':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "sonic-bgp-global:vrf": args[0] }
    elif func.__name__ == 'delete_sonic_bgp_global_sonic_bgp_global_bgp_globals_bgp_globals_list':
        keypath = [ args[0], IDENTIFIER, NAME1 ]

    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_as':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:as": int(args[1]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_router_id':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:router-id": args[1] }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_enabled':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:enabled": True if args[1] == 'True' else False }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_restart_time':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:restart-time": int(args[1]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_stale_routes_time':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:stale-routes-time": args[1] }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config_allow_multiple_as':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:allow-multiple-as": True if args[1] == 'True' else False }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_always_compare_med':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:always-compare-med": True if args[1] == 'True' else False }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_ignore_as_path_length':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:ignore-as-path-length": True if args[1] == 'True' else False }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_external_compare_router_id':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:external-compare-router-id": True if args[1] == 'True' else False }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_external_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:external-route-distance": int(args[1]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_internal_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:internal-route-distance": int(args[1]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_local_as':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:local-as": int(args[2]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_local_as':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:local-as": int(args[2]) }
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_config':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:afi-safi-name": args[1] }
    elif func.__name__ == 'patch_openconfig_network_instance1348121867':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
    elif func.__name__ == 'patch_openconfig_network_instance1543452951':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
    elif func.__name__ == 'patch_openconfig_network_instance1717438887':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:peer-group": args[2] }
    elif func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_router_id' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_enabled' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_restart_time' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_stale_routes_time' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config_allow_multiple_as' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_always_compare_med' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_ignore_as_path_length' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_external_compare_router_id' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_external_route_distance' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_internal_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
    elif func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_local_as' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_local_as' or \
         func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_config':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
    else:
        body = {}

    return keypath,body

def run(func, args):

    c = openconfig_network_instance_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_network_instance_client.OpenconfigNetworkInstanceApi(api_client=openconfig_network_instance_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else:
           api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            if 'sonic_bgp_global:sonic_bgp_global' in api_response:
                value = api_response['sonic_bgp_global:sonic_bgp_global']
                if 'WRED_PROFILE' in value:
                    tup = value['WRED_PROFILE']

            if api_response is None:
                print("Failed")
            else:
                if func.__name__ == 'get_sonic_bgp_global_sonic_bgp_global_bgp_globals_bgp_globals_list':
                     show_cli_output(args[1], api_response)
                elif func.__name__ == 'get_sonic_bgp_global_sonic_bgp_global':
                     show_cli_output(args[0], api_response)
    except ApiException as e:
        print("Exception when calling oc_bgp_global_client->%s : %s\n" %(func.__name__, e))

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_network_instance_client.OpenconfigNetworkInstanceApi.__dict__)

    run(func, sys.argv[2:])
