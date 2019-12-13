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
    if args[1] != "bgp":
        community_member.append(args[1])
    if "local-AS" in args or "no-peer" in args:
        community_member.append("NOPEER")
    if "no-export" in args:
        community_member.append("NO_EXPORT")
    if "no-advertise" in args:
        community_member.append("NO_ADVERTISE")
    if "all" in args:
         match_options = "ALL"
    else:
        match_options = "ANY"

    body = {"openconfig-bgp-policy:community-sets":{"community-set":[{"community-set-name": args[0],
            "config":{"community-set-name":args[0],"community-member":community_member,
            "match-set-options":match_options}}]}}

    return body

def generate_extcommunity_standard_body(args):
    extcommunity_member = []
    if 'all' in args:
        match_options = "ALL"
    else:
        match_options = "ANY"

    if args[1] == "soo":
        extcommunity_member.append("route-original:"+args[7])
    else:
        extcommunity_member.append("route-target:"+args[7])

    body = {"openconfig-bgp-policy:ext-community-sets":{"ext-community-set":[{"ext-community-set-name": args[0],
            "config":{"ext-community-set-name":args[0],"ext-community-member":extcommunity_member,"match-set-options": match_options}}]}}
    return body

def invoke(func, args):
    body = None
    aa = cc.ApiClient()
    if func == 'bgp_community_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = generate_community_standard_body(args)
        return aa.patch(keypath,body)

    elif func == 'bgp_community_expanded':
        community_member = ["REGEX:"+args[1]]
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = {"openconfig-bgp-policy:community-sets":{"community-set":[{"community-set-name": args[0],"config":{"community-set-name":args[0],
                "community-member":community_member,"match-set-options":"ANY"}}]}}
        return aa.patch(keypath,body)


    elif func == 'bgp_extcommunity_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = generate_extcommunity_standard_body(args)
        return aa.patch(keypath,body)

    elif func == 'bgp_extcommunity_expanded':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = {"openconfig-bgp-policy:ext-community-sets":{"ext-community-set":[{"ext-community-set-name": args[0],
                "config":{"ext-community-set-name":args[0],"ext-community-member":[args[1]],"match-set-options": "ANY"}}]}}
        return aa.patch(keypath,body)

    elif func == 'bgp_as_path_list':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets')
        body = {"openconfig-bgp-policy:as-path-sets":{"as-path-set":[{"as-path-set-name": args[0],"config":{"as-path-set-name":args[0],
                "as-path-set-member":[args[1]]}}]}}
        return aa.patch(keypath,body)

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

