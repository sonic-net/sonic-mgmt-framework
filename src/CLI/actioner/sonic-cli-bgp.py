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

EXTGLOBAF_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi'
EXTGLOBAF_OCSTRG_LEN=len(EXTGLOBAF_OCSTRG)
DELETE_EXTGLOBAF_OCPREFIX=DELETE_OCPREFIX+EXTGLOBAF_OCSTRG
DELETE_EXTGLOBAF_OCPREFIX_LEN=len(DELETE_EXTGLOBAF_OCPREFIX)

NEIGAF_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi'
NEIGAF_OCSTRG_LEN=len(NEIGAF_OCSTRG)
DELETE_NEIGAF_OCPREFIX=DELETE_OCPREFIX+NEIGAF_OCSTRG
DELETE_NEIGAF_OCPREFIX_LEN=len(DELETE_NEIGAF_OCPREFIX)

EXTNGHAF_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi'
EXTNGHAF_OCSTRG_LEN=len(EXTNGHAF_OCSTRG)
DELETE_EXTNGHAF_OCPREFIX=DELETE_OCPREFIX+EXTNGHAF_OCSTRG
DELETE_EXTNGHAF_OCPREFIX_LEN=len(DELETE_EXTNGHAF_OCPREFIX)

EXTNGH_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor'
EXTNGH_OCSTRG_LEN=len(EXTNGH_OCSTRG)
DELETE_EXTNGH_OCPREFIX=DELETE_OCPREFIX+EXTNGH_OCSTRG
DELETE_EXTNGH_OCPREFIX_LEN=len(DELETE_EXTNGH_OCPREFIX)

PGAF_OCSTRG='openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi'
PGAF_OCSTRG_LEN=len(PGAF_OCSTRG)
DELETE_PGAF_OCPREFIX=DELETE_OCPREFIX+PGAF_OCSTRG
DELETE_PGAF_OCPREFIX_LEN=len(DELETE_PGAF_OCPREFIX)

EXTPGAF_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi'
EXTPGAF_OCSTRG_LEN=len(EXTPGAF_OCSTRG)
DELETE_EXTPGAF_OCPREFIX=DELETE_OCPREFIX+EXTPGAF_OCSTRG
DELETE_EXTPGAF_OCPREFIX_LEN=len(DELETE_EXTPGAF_OCPREFIX)

EXTPG_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group'
EXTPG_OCSTRG_LEN=len(EXTPG_OCSTRG)
DELETE_EXTPG_OCPREFIX=DELETE_OCPREFIX+EXTPG_OCSTRG
DELETE_EXTPG_OCPREFIX_LEN=len(DELETE_EXTPG_OCPREFIX)

EXTGLB_OCSTRG='openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global'
EXTGLB_OCSTRG_LEN=len(EXTGLB_OCSTRG)
DELETE_EXTGLB_OCPREFIX=DELETE_OCPREFIX+EXTGLB_OCSTRG
DELETE_EXTGLB_OCPREFIX_LEN=len(DELETE_EXTGLB_OCPREFIX)

OCEXTPREFIX_PATCH='PATCH'
OCEXTPREFIX_PATCH_LEN=len(OCEXTPREFIX_PATCH)
OCEXTPREFIX_DELETE='DELETE'
OCEXTPREFIX_DELETE_LEN=len(OCEXTPREFIX_DELETE)

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None

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
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_global_use_multiple_paths_ebgp_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/use-multiple-paths/ebgp/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-network-instance:config" : { "allow-multiple-as" : True if args[1] == 'True' else False, "openconfig-bgp-ext:as-set" : True if 'as-set' in args[2:] else False } }
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
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_compare_confed_as_path':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/openconfig-bgp-ext:compare-confed-as-path',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:compare-confed-as-path": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_med_missing_as_worst':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/openconfig-bgp-ext:med-missing-as-worst',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:med-missing-as-worst": True if 'missing-as-worst' in args[1:] else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_route_selection_options_config_med_confed':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/route-selection-options/config/openconfig-bgp-ext:med-confed',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:med-confed": True if 'confed' in args[1:] else False }
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
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_bgp_ext_route_reflector_config_route_reflector_cluster_id':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:bgp-ext-route-reflector/config/route-reflector-cluster-id',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:route-reflector-cluster-id": args[1] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_logging_options_config_log_neighbor_state_changes':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:logging-options/config/log-neighbor-state-changes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:log-neighbor-state-changes": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_route_flap_damping_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:route-flap-damping/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:config" : { "enabled" : True if args[1] == 'True' else False } }
        if len(args) > 2:
            body["openconfig-bgp-ext:config"]["half-life"] = int(args[2])
            if len(args) > 3:
                body["openconfig-bgp-ext:config"]["reuse-threshold"] = int(args[3])
                if len(args) > 4:
                    body["openconfig-bgp-ext:config"]["suppress-threshold"] = int(args[4])
                    if len(args) > 5:
                        body["openconfig-bgp-ext:config"]["max-suppress"] = int(args[5])
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_disable_ebgp_connected_route_check':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:disable-ebgp-connected-route-check',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:disable-ebgp-connected-route-check": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_graceful_shutdown':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:graceful-shutdown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = {"openconfig-bgp-ext:graceful-shutdown": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_fast_external_failover':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:fast-external-failover',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:fast-external-failover": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_network_import_check':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:network-import-check',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:network-import-check": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_bgp_ext_route_reflector_config_allow_outbound_policy':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:bgp-ext-route-reflector/config/allow-outbound-policy',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:allow-outbound-policy": True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_graceful_restart_config_preserve_fw_state':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/graceful-restart/config/openconfig-bgp-ext:preserve-fw-state',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:preserve-fw-state" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_coalesce_time':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:coalesce-time',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:coalesce-time" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_read_quanta':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:read-quanta',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:read-quanta" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_write_quanta':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:write-quanta',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:write-quanta" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_clnt_to_clnt_reflection':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:clnt-to-clnt-reflection',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:clnt-to-clnt-reflection" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_deterministic_med':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:deterministic-med',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:deterministic-med" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_max_med_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:max-med/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:config" : { "time" : int(args[1]), "max-med-val" : int(args[2]) } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_global_defaults_config_ipv4_unicast':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:global-defaults/config/ipv4-unicast',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:ipv4-unicast" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_global_defaults_config_local_preference':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:global-defaults/config/local-preference',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:local-preference" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_global_defaults_config_show_hostname':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:global-defaults/config/show-hostname',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:show-hostname" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_global_defaults_config_shutdown':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:global-defaults/config/shutdown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:shutdown" : True if args[1] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_global_defaults_config_subgroup_pkt_queue_max':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:global-defaults/config/subgroup-pkt-queue-max',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:subgroup-pkt-queue-max" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_config_route_map_process_delay':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/config/openconfig-bgp-ext:route-map-process-delay',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:route-map-process-delay" : int(args[1]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_update_delay_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/openconfig-bgp-ext:update-delay/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1)
        body = { "openconfig-bgp-ext:config" : { "max-delay" : int(args[1]), "establish-wait" : int(args[1]) } }
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
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_network_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:network/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "oc-bgp-ext:config" : { "prefix": args[2], "backdoor": True if args[3] == 'backdoor' else False } }
        if len(args) > 4:
            body["oc-bgp-ext:config"]["policy-name"] = args[4]
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_config_table_map_name':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:table-map-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "openconfig-bgp-ext:table-map-name" : args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_network_config_policy_name':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:network/config/policy-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "openconfig-bgp-ext:policy-name" : args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_aggregate_address_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:aggregate-address/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "oc-bgp-ext:config" : { "prefix": args[2], "as-set": True if 'as-set' in args[3:] else False, "summary-only": True if 'summary-only' in args[3:] else False } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_default_route_distance_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:default-route-distance/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        body = { "oc-bgp-ext:config" : { "external-route-distance" : int(args[2]), "internal-route-distance" : int(args[3]) } }
        return api.patch(keypath, body)

    elif func == 'patch_openconfig_network_instance1348121867' or func == 'delete_openconfig_network_instance1348121867':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/use-multiple-paths/ebgp/config/maximum-paths',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        if func[0:DELETE_OCPREFIX_LEN] == DELETE_OCPREFIX:
            return api.delete(keypath)
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance1543452951' or func == 'delete_openconfig_network_instance1543452951':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/use-multiple-paths/ibgp/config/maximum-paths',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        if func[0:DELETE_OCPREFIX_LEN] == DELETE_OCPREFIX:
            return api.delete(keypath)
        body = { "openconfig-network-instance:maximum-paths": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext3691744053' or func == 'delete_openconfig_bgp_ext3691744053':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/use-multiple-paths/ibgp/config/openconfig-bgp-ext:equal-cluster-length',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        if func[0:DELETE_OCPREFIX_LEN] == DELETE_OCPREFIX:
            return api.delete(keypath)
        body = { "openconfig-bgp-ext:equal-cluster-length": True if 'equal-cluster-length' in args[2:] else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance1717438887' and func == 'delete_openconfig_network_instance1717438887':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/dynamic-neighbor-prefixes/dynamic-neighbor-prefix={prefix}/config/peer-group', 
                name=args[0], identifier=IDENTIFIER, name1=NAME1, prefix=args[1])
        if func[0:DELETE_OCPREFIX_LEN] == DELETE_OCPREFIX:
            return api.delete(keypath)
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
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_ebgp_multihop_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/ebgp-multihop/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:config" : { "enabled" : True if args[2] == 'True' else False, "multihop-ttl" : int(args[3]) } }
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
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_peer_type':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/peer-type',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:peer-type": args[2].upper() }
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
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_timers_config_minimum_advertisement_interval':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/timers/config/minimum-advertisement-interval',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:minimum-advertisement-interval": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_capability_extended_nexthop':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:capability-extended-nexthop',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:capability-extended-nexthop": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_capability_dynamic':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:capability-dynamic',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:capability-dynamic": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_disable_ebgp_connected_route_check':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:disable-ebgp-connected-route-check',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:disable-ebgp-connected-route-check": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_enforce_first_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:enforce-first-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:enforce-first-as": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_local_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/local-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:local-as": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_transport_config_passive_mode':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/transport/config/passive-mode',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:passive-mode": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_auth_password':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/auth-password',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-network-instance:auth-password": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_solo_peer':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:solo-peer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:solo-peer": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_ttl_security_hops':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:ttl-security-hops',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:ttl-security-hops": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_bfd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:bfd',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:bfd" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_dont_negotiate_capability':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:dont-negotiate-capability',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:dont-negotiate-capability" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_enforce_multihop':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:enforce-multihop',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:enforce-multihop" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_override_capability':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:override-capability',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:override-capability" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_peer_port':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:peer-port',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:peer-port" : int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_config_strict_capability_match':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/config/openconfig-bgp-ext:strict-capability-match',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        body = { "openconfig-bgp-ext:strict-capability-match" : True if args[2] == 'True' else False }
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
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_add_paths_config_tx_all_paths':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/add-paths/config/openconfig-bgp-ext:tx-all-paths',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:tx-all-paths" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_add_paths_config_tx_bestpath_per_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/add-paths/config/openconfig-bgp-ext:tx-bestpath-per-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = {  "openconfig-bgp-ext:tx-bestpath-per-as" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_as_override':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:as-override',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:as-override": True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_attribute_unchanged_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:attribute-unchanged/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "as-path" : True if 'as-path' in args[3:] else False, "med" : True if 'med' in args[3:] else False, "next-hop" : True if 'next-hop' in args[3:] else False } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_filter_list_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:filter-list/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config" : { "as-path-set-name" : args[3], "direction" : "OUTBOUND" if args[4] == 'out' else "INBOUND" } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_next_hop_self_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:next-hop-self/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "enabled" : True if args[3] == 'True' else False, "force" : True if 'force' in args[3:] else False } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_prefix_list_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:prefix-list/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config" : { "prefix-set-name" : args[3], "direction" : "OUTBOUND" if args[4] == 'out' else "INBOUND" } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_remove_private_as_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:remove-private-as/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "enabled" : True if args[3] == 'True' else False, "all" : True if 'all' in args[3:] else False,  "replace-as" : True if 'replace-AS' in args[3:] else False} }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_route_reflector_client':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:route-reflector-client',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:route-reflector-client" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_send_community':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:send-community',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:send-community" : args[3].upper() }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_soft_reconfiguration_in':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:soft-reconfiguration-in',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:soft-reconfiguration-in" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_unsuppress_map_name':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:unsuppress-map-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:unsuppress-map-name" : args[3] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_weight':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:weight',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:weight" : int(args[3]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_capability_orf_config_orf_type':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:capability-orf/config/orf-type',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:orf-type" : args[3].upper() }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_config_route_server_client':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:route-server-client',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:route-server-client" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_allow_own_as_config_as_count':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/as-count',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:as-count": int(args[3]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_allow_own_as_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:enabled": True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_allow_own_as_config_origin':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/origin',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:origin": True if 'origin' in args[3:] else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_description':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/description',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:description": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_ebgp_multihop_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/ebgp-multihop/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:config" : { "enabled" : True if args[2] == 'True' else False, "multihop-ttl" : int(args[3]) } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_peer_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/peer-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:peer-as": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_peer_type':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/peer-type',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:peer-type": args[2].upper() }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_timers_config_keepalive_interval':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/timers/config/keepalive-interval',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:keepalive-interval": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_timers_config_hold_time':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/timers/config/hold-time',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:hold-time": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_transport_config_local_address':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/transport/config/local-address',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:local-address": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_capability_extended_nexthop':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:capability-extended-nexthop',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:capability-extended-nexthop": True if args[2] == 'extended-nexthop' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_disable_ebgp_connected_route_check':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:disable-ebgp-connected-route-check',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:disable-ebgp-connected-route-check": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_enforce_first_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:enforce-first-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:enforce-first-as": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_local_as':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/local-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:local-as": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_transport_config_passive_mode':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/transport/config/passive-mode',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:passive-mode": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_auth_password':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/auth-password',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-network-instance:auth-password": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_solo_peer':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:solo-peer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:solo-peer": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_ttl_security_hops':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:ttl-security-hops',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:ttl-security-hops": int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_bfd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:bfd',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:bfd" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_dont_negotiate_capability':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:dont-negotiate-capability',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:dont-negotiate-capability" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_enforce_multihop':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:enforce-multihop',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:enforce-multihop" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_override_capability':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:override-capability',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:override-capability" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_peer_port':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:peer-port',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:peer-port" : int(args[2]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_config_strict_capability_match':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/config/openconfig-bgp-ext:strict-capability-match',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        body = { "openconfig-bgp-ext:strict-capability-match" : True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config": { "afi-safi-name": args[2] } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:enabled": True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_add_paths_config_tx_all_paths':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/add-paths/config/openconfig-bgp-ext:tx-all-paths',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:tx-all-paths" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_as_override':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:as-override',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:as-override": True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_attribute_unchanged_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:attribute-unchanged/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "as-path" : True if 'as-path' in args[3:] else False, "med" : True if 'med' in args[3:] else False, "next-hop" : True if 'next-hop' in args[3:] else False } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_filter_list_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:filter-list/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config" : { "as-path-set-name" : args[3], "direction" : "OUTBOUND" if args[4] == 'out' else "INBOUND" } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_next_hop_self_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:next-hop-self/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "enabled" : True if args[3] == 'True' else False, "force" : True if 'force' in args[3:] else False } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_prefix_list_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:prefix-list/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:config" : { "prefix-set-name" : args[3], "direction" : "OUTBOUND" if args[4] == 'out' else "INBOUND" } }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_remove_private_as_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:remove-private-as/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:config" : { "enabled" : True if args[3] == 'True' else False, "all" : True if 'all' in args[3:] else False,  "replace-as" : True if 'replace-AS' in args[3:] else False} }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_route_reflector_client':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:route-reflector-client',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:route-reflector-client" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_send_community':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:send-community',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:send-community" : args[3].upper() }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_soft_reconfiguration_in':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:soft-reconfiguration-in',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:soft-reconfiguration-in" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_unsuppress_map_name':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:unsuppress-map-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:unsuppress-map-name" : args[3] }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_weight':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:weight',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:weight" : int(args[3]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_allow_own_as_config_as_count':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/as-count',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:as-count" : int(args[3]) }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_allow_own_as_config_enabled':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-network-instance:enabled": True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_capability_orf_config_orf_type':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:capability-orf/config/orf-type',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:orf-type" : args[3].upper() }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_config_route_server_client':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/config/openconfig-bgp-ext:route-server-client',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:route-server-client" : True if args[3] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_allow_own_as_config_origin':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:allow-own-as/config/origin',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        body = { "openconfig-bgp-ext:origin": True if 'origin' in args[3:] else False }
        return api.patch(keypath, body)

    elif attr == 'openconfig_network_instance_network_instances_network_instance_table_connections_table_connection_config_import_policy':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/table-connections/table-connection={src_protocol},{dst_protocol},{address_family}/config/import-policy',
                name=args[0], src_protocol= "STATIC" if 'static' == args[2] else "DIRECTLY_CONNECTED", dst_protocol=IDENTIFIER, address_family=args[1].split('_',1)[0])
        if op == 'patch':
            body = { "openconfig-network-instance:import-policy" : [ args[3] ] }
            return api.patch(keypath, body)
        else:
            return api.delete(keypath)
    elif attr == 'openconfig_network_instance_network_instances_network_instance_table_connections_table_connection_config':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/table-connections/table-connection={src_protocol},{dst_protocol},{address_family}/config',
                name=args[0], src_protocol= "STATIC" if 'static' == args[2] else "DIRECTLY_CONNECTED", dst_protocol=IDENTIFIER, address_family=args[1].split('_',1)[0])
        if op == 'patch':
            body = { "openconfig-network-instance:config" : { "address-family" : args[1].split('_',1)[0] } }
            return api.patch(keypath, body)
        else:
            return api.delete(keypath)
    elif attr == 'openconfig_network_instance_network_instances_network_instance_table_connections_table_connection':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/table-connections/table-connection={src_protocol},{dst_protocol},{address_family}',
                name=args[0], src_protocol= "STATIC" if 'static' == args[2] else "DIRECTLY_CONNECTED", dst_protocol=IDENTIFIER, address_family=args[1].split('_',1)[0])
        if op == 'patch':
            body = { "openconfig-network-instance:table-connection": [ { "config": { "address-family": args[1].split('_',1)[0] } } ] }
            return api.patch(keypath, body)
        else:
            return api.delete(keypath)

    elif op == OCEXTPREFIX_DELETE or op == OCEXTPREFIX_PATCH:
        # PATCH_ and DELETE_ prefixes (all caps) means no swaggar-api string
        if attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_apply_policy_config_import_policy':
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
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv4_unicast_config_default_policy_name':
            # openconfig_bgp_ext841615068
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/config/openconfig-bgp-ext:default-policy-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:default-policy-name" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config':
            # openconfig_network_instance1753955874
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config':
            # openconfig_network_instance985144991
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_neighbors_neighbor_afi_safis_afi_safi_ipv6_unicast_config_default_policy_name':
            # openconfig_bgp_ext2059791605
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/neighbors/neighbor={neighbor_address}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/config/openconfig-bgp-ext:default-policy-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:default-policy-name" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_timers_config_minimum_advertisement_interval':
            # openconfig_network_instance1223315985
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/timers/config/minimum-advertisement-interval',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:minimum-advertisement-interval":  args[2]  }

        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_apply_policy_config_import_policy':
            # openconfig_network_instance1779097864
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/apply-policy/config/import-policy',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:import-policy":  [ args[3] ] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_apply_policy_config_export_policy':
            # openconfig_network_instance251836598
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/apply-policy/config/export-policy',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:export-policy": [ args[3] ] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config':
            # openconfig_network_instance3096500951
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv4_unicast_config_default_policy_name':
            # openconfig_bgp_ext2561500065
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv4-unicast/config/openconfig-bgp-ext:default-policy-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:default-policy-name" : args[3] }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_add_paths_config_tx_bestpath_per_as':
            # openconfig_bgp_ext3272932244
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/add-paths/config/openconfig-bgp-ext:tx-bestpath-per-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:tx-bestpath-per-as" : True if args[3] == 'True' else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_default_policy_name':
            # openconfig_bgp_ext777259601
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/config/openconfig-bgp-ext:default-policy-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:default-policy-name" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config':
            # openconfig_network_instance1753955874
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_ipv6_unicast_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/ipv6-unicast/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config':
            # openconfig_network_instance202630882
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:config" : { "max-prefixes" : int(args[3]), "warning-threshold-pct" : int(args[4]), "prevent-teardown" : True if 'warning-only' in args[4:] else False } }
                if 'restart' in args[4:]:
                   body["openconfig-network-instance:config"]["restart-timer"] = args[-1]
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_max_prefixes':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/max-prefixes',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:max-prefixes" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_warning_threshold_pct':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/warning-threshold-pct',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:warning-threshold-pct" : int(args[3]) }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_prevent_teardown':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/prevent-teardown',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:prevent-teardown" : True if 'warning-only' in args[3] else False }
        elif attr == 'openconfig_network_instance_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_l2vpn_evpn_config_prefix_limit_config_restart_timer':
            # openconfig_network_instance3828573403
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/l2vpn-evpn/prefix-limit/config/restart-timer',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-network-instance:restart-timer" : args[3] }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_attribute_unchanged_config_as_path':
            # openconfig_bgp_ext2045507776
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer_group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:attribute-unchanged/config/as-path',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:as-path" : True if args[3] == 'True' else False }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_attribute_unchanged_config_next_hop':
            # openconfig_bgp_ext2045507776
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer_group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:attribute-unchanged/config/next-hop',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:next-hop" : True if args[3] == 'True' else False }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_filter_list_config_as_path_set_name':
            # openconfig_bgp_ext1021391242
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:filter-list/config/as-path-set-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:as-path-set-name" : args[3] }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_prefix_list_config_prefix_set_name':
            # openconfig_bgp_ext1545829530
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:prefix-list/config/prefix-set-name',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:prefix-set-name" : args[3] }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_remove_private_as_config_enabled':
            # openconfig_bgp_ext2741086768
            keypath = cc.Path(' /restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:remove-private-as/config/enabled',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:enabled" : True if args[3] == 'True' else False }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_peer_groups_peer_group_afi_safis_afi_safi_remove_private_as_config_replace_as':
            # openconfig_bgp_ext1124459141
            keypath = cc.Path(' /restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/peer-groups/peer-group={peer_group_name}/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:remove-private-as/config/replace-as',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:repplace-as" : True if args[3] == 'True' else False }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_default_route_distance_config_external_route_distance':
            # openconfig_bgp_ext1219850592
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:default-route-distance/config/external-route-distance',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:external-route-distance" : int(args[2]) }
        elif attr == 'openconfig_bgp_ext_network_instances_network_instance_protocols_protocol_bgp_global_afi_safis_afi_safi_default_route_distance_config_internal_route_distance':
            # openconfig_bgp_ext1240612726
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}/protocols/protocol={identifier},{name1}/bgp/global/afi-safis/afi-safi={afi_safi_name}/openconfig-bgp-ext:default-route-distance/config/internal-route-distance',
                name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
            if op == OCEXTPREFIX_PATCH:
                body = { "openconfig-bgp-ext:internal-route-distance" : int(args[2]) }
        else:
            return api.cli_not_implemented(func)
        if op == OCEXTPREFIX_PATCH:
            return api.patch(keypath, body)
        else:
            return api.delete(keypath)

    # OC-prefixes can be substring of parent prefixes, so check the longer child prefixes before the parents.
    elif func[0:DELETE_NEIGAF_OCPREFIX_LEN] == DELETE_NEIGAF_OCPREFIX or func[0:DELETE_EXTNGHAF_OCPREFIX_LEN] == DELETE_EXTNGHAF_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{neighbor-address}', '{neighbor_address}').replace('{afi-safi-name}', '{afi_safi_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1], afi_safi_name=args[2])
        return api.delete(keypath)
    elif func[0:DELETE_NEIGHB_OCPREFIX_LEN] == DELETE_NEIGHB_OCPREFIX or func[0:DELETE_EXTNGH_OCPREFIX_LEN] == DELETE_EXTNGH_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{neighbor-address}', '{neighbor_address}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, neighbor_address=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_PGAF_OCPREFIX_LEN] == DELETE_PGAF_OCPREFIX or func[0:DELETE_EXTPGAF_OCPREFIX_LEN] == DELETE_EXTPGAF_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{peer-group-name}', '{peer_group_name}').replace('{afi-safi-name}', '{afi_safi_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1], afi_safi_name=args[2])
        return api.delete(keypath)
    elif func[0:DELETE_PEERGP_OCPREFIX_LEN] == DELETE_PEERGP_OCPREFIX or func[0:DELETE_EXTPG_OCPREFIX_LEN] == DELETE_EXTPG_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{peer-group-name}', '{peer_group_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, peer_group_name=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_GLOBAF_OCPREFIX_LEN] == DELETE_GLOBAF_OCPREFIX or func[0:DELETE_EXTGLOBAF_OCPREFIX_LEN] == DELETE_EXTGLOBAF_OCPREFIX:
        uri = restconf_map[attr]
        keypath = cc.Path(uri.replace('{afi-safi-name}', '{afi_safi_name}'),
               name=args[0], identifier=IDENTIFIER, name1=NAME1, afi_safi_name=args[1])
        return api.delete(keypath)
    elif func[0:DELETE_GLOBAL_OCPREFIX_LEN] == DELETE_GLOBAL_OCPREFIX or func[0:DELETE_EXTGLB_OCPREFIX_LEN] == DELETE_EXTGLB_OCPREFIX:
        keypath = cc.Path(restconf_map[attr],
               name=args[0], identifier=IDENTIFIER, name1=NAME1)
        return api.delete(keypath)

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if api_response is None:
                print("Failed")
                sys.exit(1)
    else:
        print response.error_message()
        sys.exit(1)

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

