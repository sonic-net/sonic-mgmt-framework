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
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

def generate_community_standard_body(args):
    community_member = []
    match_options = "ANY"
    for arg in args[5:]:
        if "local-AS" == arg:
           community_member.append("NO_EXPORT_SUBCONFED")
        elif  "no-peer" == arg:
           community_member.append("NOPEER")
        elif "no-export" == arg:
           community_member.append("NO_EXPORT")
        elif "no-advertise" == arg:
           community_member.append("NO_ADVERTISE")
        elif "all" == arg:
           match_options = "ALL"
        elif "any" == arg:
           match_options = "ANY"
        else:
           community_member.append(arg)

    body = {"openconfig-bgp-policy:community-sets":{"community-set":[{"community-set-name": args[4],
            "config":{"community-set-name":args[4],"community-member":community_member,
            "match-set-options":match_options}}]}}

    return body

def generate_extcommunity_standard_body(args):
    extcommunity_member = []
    match_options = "ANY"
    i = 5
    for arg in args[5:]:
        if "all" == arg:
           match_options = "ALL"
        elif "any" == arg:
           match_options = "ANY"
        elif "soo" == arg:
           extcommunity_member.append("route-origin:"+args[i+1])
        elif "rt" == arg:
           extcommunity_member.append("route-target:"+args[i+1])
        i = i + 1

    body = {"openconfig-bgp-policy:ext-community-sets":{"ext-community-set":[{"ext-community-set-name": args[4],
            "config":{"ext-community-set-name":args[4],"ext-community-member":extcommunity_member,"match-set-options": match_options}}]}}
    return body

def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    #bgp-community-standard commands
    if func == 'bgp_community_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = generate_community_standard_body(args)
        return aa.patch(keypath,body)

    #bgp-community-expanded commands
    elif func == 'bgp_community_expanded':
        community_member = ["REGEX:"+args[1]]
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = {"openconfig-bgp-policy:community-sets":{"community-set":[{"community-set-name": args[0],"config":{"community-set-name":args[0],
                "community-member":community_member,"match-set-options":"ANY"}}]}}
        return aa.patch(keypath,body)

    # Remove the bgp-community-standard set.
    elif func == 'bgp_community_standard_delete':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}',community_list_name=args[0])
        return aa.delete(keypath)

    # Remove the bgp-community-expanded set.
    elif func == 'bgp_community_expanded_delete':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}',community_list_name=args[0])
        return aa.delete(keypath)

    #bgp-extcommunity-standard commands
    elif func == 'bgp_extcommunity_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = generate_extcommunity_standard_body(args)
        return aa.patch(keypath,body)

    #bgp-extcommunity-expanded commands
    elif func == 'bgp_extcommunity_expanded':
        community_member = ["REGEX:"+args[1]]
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = {"openconfig-bgp-policy:ext-community-sets":{"ext-community-set":[{"ext-community-set-name": args[0],
                "config":{"ext-community-set-name":args[0],"ext-community-member":community_member,"match-set-options": "ANY"}}]}}
        return aa.patch(keypath,body)

    # Remove the bgp-extcommunity-standard set.
    elif func == 'bgp_extcommunity_standard_delete':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={extcommunity_list_name}',extcommunity_list_name=args[0])
        return aa.delete(keypath)

    # Remove the bgp-extcommunity-expanded set.
    elif func == 'bgp_extcommunity_expanded_delete':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={extcommunity_list_name}',extcommunity_list_name=args[0])
        return aa.delete(keypath)

    # bgp-as-path-list command
    elif func == 'bgp_as_path_list':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets')
        body = {"openconfig-bgp-policy:as-path-sets":{"as-path-set":[{"as-path-set-name": args[0],"config":{"as-path-set-name":args[0],
                "as-path-set-member":[args[1]]}}]}}
        return aa.patch(keypath,body)

    # Remove the bgp-as-path-list set.
    elif func == 'bgp_as_path_list_delete':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={as_path_set_name}',as_path_set_name=args[0])
        return aa.delete(keypath)

    else:
        return body


def run(func, args):
    try:
        api_response = invoke(func,args)
        return
    except:
            # system/network error
            print "Error: Transaction Failure"



if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])

