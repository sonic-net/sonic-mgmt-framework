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
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

IDENTIFIER='BGP'
NAME1='bgp'

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None

    if func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_as':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/as', 
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:as": int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_router_id':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:router-id": args[1] }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_enabled':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:enabled": True if args[1] == 'True' else False }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_restart_time':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:restart-time": int(args[1]) }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_stale_routes_time':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:stale-routes-time": args[1] }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config_allow_multiple_as':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:allow-multiple-as": True if args[1] == 'True' else False }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_always_compare_med':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:always-compare-med": True if args[1] == 'True' else False }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_ignore_as_path_length':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:ignore-as-path-length": True if args[1] == 'True' else False }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_external_compare_router_id':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:external-compare-router-id": True if args[1] == 'True' else False }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_external_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:external-route-distance": int(args[1]) }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_internal_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
        body = { "openconfig-network-instance:internal-route-distance": int(args[1]) }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_local_as':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:local-as": int(args[2]) }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_local_as':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
        body = { "openconfig-network-instance:local-as": int(args[2]) }
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_config':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/config',
		name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "openconfig-network-instance:afi-safi-name": args[1] }
        return api.patch(keypath, body)
	
    elif func == 'patch_openconfig_network_instance1348121867':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/use-multiple-paths/ebgp/config/maximum-paths',
		name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance1543452951':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/use-multiple-paths/ibgp/config/maximum-paths',
		name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance1717438887':
	keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/dynamic-neighbor-prefixes/dynamic-neighbor-prefix={prefix}/config/peer-group', 
		name=args[0], identifier=IDENTIFIER, name1=NAME1, prefix=args[1])
        body = { "openconfig-network-instance:peer-group": args[2] }
        return api.patch(keypath, body)
    elif func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_router_id' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_enabled' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_restart_time' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_stale_routes_time' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config_allow_multiple_as' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_always_compare_med' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_ignore_as_path_length' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_external_compare_router_id' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_external_route_distance' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_internal_route_distance':
        keypath = [ args[0], IDENTIFIER, NAME1 ]
    elif func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_local_as' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_local_as' or \
         func == 'delete_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_config':
        keypath = [ args[0], IDENTIFIER, NAME1, args[1] ]
    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if api_response is None:
                print("Failed")
    else:
        print response.error_message()

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

