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
from routemap_openconfig_to_restconf import restconf_map 


def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None

    op, attr = func.split('_', 1)
    uri = restconf_map[attr]

    if op == 'patch':
        if attr == 'openconfig_routing_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_config_policy_result':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-routing-policy:policy-result" : "ACCEPT_ROUTE" if "permit" == args[2] else "REJECT_ROUTE" }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_next_hop':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:set-next-hop" : args[2] }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_local_pref':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:set-local-pref" : int(args[2]) }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_config_set_route_origin':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:set-route-origin" : args[2].upper() }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_as_path_prepend_config_asn':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:asn" : int(args[2]) }
            return api.patch(keypath, body)  
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config_method':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:method" : args[2] }
            return api.patch(keypath, body) 
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_community_config_options':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:options" : args[2] }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy3674057445':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:communities" : [args[2]] }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_config_method':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:method" : args[2] }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy_routing_policy_policy_definitions_policy_definition_statements_statement_actions_bgp_actions_set_ext_community_config_options':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:options" : args[2] }
            return api.patch(keypath, body)
        elif attr == 'openconfig_bgp_policy2318914281':
            keypath = cc.Path(uri, name=args[0], name1=args[1])
            body = { "openconfig-bgp-policy:communities" : ["route-target:"+args[3]] if "rt" == args[2] else ["route-origin:"+args[3]] }
            return api.patch(keypath, body)
    elif op == 'delete':
        keypath = cc.Path(uri,
                name=args[0], name1=args[1])
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

