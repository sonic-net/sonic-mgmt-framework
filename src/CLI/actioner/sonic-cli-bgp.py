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
from bgp_openconfig_to_restconf_map import restconf_map 

IDENTIFIER='BGP'
NAME1='bgp'

DELETE_OCPREFIX='delete_'
DELETE_OCPREFIX_LEN=len(DELETE_OCPREFIX)

GLOBAL_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_'
GLOBAL_OCSTRG_LEN=len(GLOBAL_OCSTRG)
DELETE_GLOBAL_OCPREFIX=DELETE_OCPREFIX+GLOBAL_OCSTRG
DELETE_GLOBAL_OCPREFIX_LEN=len(DELETE_GLOBAL_OCPREFIX)

NEIGHB_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor'
NEIGHB_OCSTRG_LEN=len(NEIGHB_OCSTRG)
DELETE_NEIGHB_OCPREFIX=DELETE_OCPREFIX+NEIGHB_OCSTRG
DELETE_NEIGHB_OCPREFIX_LEN=len(DELETE_NEIGHB_OCPREFIX)

PEERGP_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group'
PEERGP_OCSTRG_LEN=len(PEERGP_OCSTRG)
DELETE_PEERGP_OCPREFIX=DELETE_OCPREFIX+PEERGP_OCSTRG
DELETE_PEERGP_OCPREFIX_LEN=len(DELETE_PEERGP_OCPREFIX)

GLOBAF_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi'
GLOBAF_OCSTRG_LEN=len(GLOBAF_OCSTRG)
DELETE_GLOBAF_OCPREFIX=DELETE_OCPREFIX+GLOBAF_OCSTRG
DELETE_GLOBAF_OCPREFIX_LEN=len(DELETE_GLOBAF_OCPREFIX)

NEIGAF_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi'
NEIGAF_OCSTRG_LEN=len(NEIGAF_OCSTRG)
DELETE_NEIGAF_OCPREFIX=DELETE_OCPREFIX+NEIGAF_OCSTRG
DELETE_NEIGAF_OCPREFIX_LEN=len(DELETE_NEIGAF_OCPREFIX)

OCEXTPREFIX_PATCH='PATCH'
OCEXTPREFIX_PATCH_LEN=len(OCEXTPREFIX_PATCH)
OCEXTPREFIX_DELETE='DELETE'
OCEXTPREFIX_DELETE_LEN=len(OCEXTPREFIX_DELETE)

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None
    # Override global NAME1 and use vrf-name (temporarily needed for new transformer)
    NAME1=args[0]

    op, attr = func.split('_', 1)

    if func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/as', 
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:as": int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_config_router_id':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/router-id',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:router-id": args[1] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/graceful-restart/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:enabled": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_restart_time':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/graceful-restart/config/restart-time',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:restart-time": int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_stale_routes_time':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/graceful-restart/config/stale-routes-time',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:stale-routes-time": args[1] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config_allow_multiple_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/use-multiple-paths/ebgp/config/allow-multiple-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:allow-multiple-as": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_always_compare_med':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/always-compare-med',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:always-compare-med": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_ignore_as_path_length':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/ignore-as-path-length',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:ignore-as-path-length": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_external_compare_router_id':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/external-compare-router-id',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:external-compare-router-id": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_external_route_distance':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/default-route-distance/config/external-route-distance',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:external-route-distance": int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_default_route_distance_config_internal_route_distance':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/default-route-distance/config/internal-route-distance',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:internal-route-distance": int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:config": { "neighbor-address": args[1] } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:config": { "peer-group-name": args[1] } }
        return api.patch(keypath, body)
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

    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:enabled": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_description':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/description',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:description": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_ebgp_multihop_config_multihop_ttl':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/ebgp-multihop/config/multihop-ttl',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:multihop-ttl": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_ebgp_multihop_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/ebgp-multihop/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:enabled": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_peer_group':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/peer-group',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:peer-group": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_peer_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/peer-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:peer-as": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_timers_config_keepalive_interval':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/timers/config/keepalive-interval',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:keepalive-interval": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_timers_config_hold_time':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/timers/config/hold-time',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:hold-time": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_transport_config_local_address':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/transport/config/local-address',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:local-address": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config": { "afi-safi-name": args[2] } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:enabled": True if args[3] == 'True' else False }
        return api.patch(keypath, body)

    elif op == OCEXTPREFIX_DELETE or op == OCEXTPREFIX_PATCH:
        # PATCH_ and DELETE_ prefixes (all caps) means no swaggar-api string
        if attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_allow_own_as_config_as_count':
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/as-count',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:as-count": int(args[3]) }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_allow_own_as_config_enabled':
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:enabled": True if args[3] == 'True' else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_apply_policy_config_import_policy':
            # openconfig_network_instance3764031561
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/apply-policy/config/import-policy',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:import-policy": [ args[3] ] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_apply_policy_config_export_policy':
            # openconfig_network_instance1837635724
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/apply-policy/config/export-policy',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:export-policy": [ args[3] ] }
        if op == OCEXTPREFIX_PATCH:
            return api.patch(keypath, body)
        else:
            return api.delete(keypath)

    # OC-prefixes can be substring of parent prefixes, so check the longer child prefixes before the parents.
    elif func[0:DELETE_NEIGAF_OCPREFIX_LEN] == DELETE_NEIGAF_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{neighbor-address}', '{neighbor_address}').replace('{afi-safi-name}', '{afi_safi_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        return api.delete(keypath)
    elif func[0:DELETE_NEIGHB_OCPREFIX_LEN] == DELETE_NEIGHB_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{neighbor-address}', '{neighbor_address}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_PEERGP_OCPREFIX_LEN] == DELETE_PEERGP_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{peer-group-name}', '{peer_group_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_GLOBAF_OCPREFIX_LEN] == DELETE_GLOBAF_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{afi-safi-name}', '{afi_safi_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_GLOBAL_OCPREFIX_LEN] == DELETE_GLOBAL_OCPREFIX:
        keypath = cc.Path(restconf_map[attr],
               name=args[0], identifier=IDENTIFIER, name1=NAME1)
        return api.delete(keypath)
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

