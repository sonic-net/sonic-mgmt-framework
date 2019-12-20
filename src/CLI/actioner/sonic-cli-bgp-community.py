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
        elif "no-peer" == arg:
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

def generate_community_expanded_body(args):
    community_member = []
    match_options = "ANY"
    for arg in args[5:]:
        if "all" == arg:
           match_options = "ALL"
        elif "any" == arg:
           match_options = "ANY"
        else:
           member = "REGEX:"+arg
           community_member.append(member)
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

def generate_extcommunity_expanded_body(args):
    extcommunity_member = []
    match_options = "ANY"
    for arg in args[5:]:
        if "all" == arg:
           match_options = "ALL"
        elif "any" == arg:
           match_options = "ANY"
        else:
           extcommunity_member.append("REGEX:"+arg)

    body = {"openconfig-bgp-policy:ext-community-sets":{"ext-community-set":[{"ext-community-set-name": args[4],
            "config":{"ext-community-set-name":args[4],"ext-community-member":extcommunity_member,"match-set-options": match_options}}]}}
    return body

def generate_community_standard_delete_keypath(args):
    member_exits = 0
    community_member = ''
    for arg in args[6:]:
        member_exits = 1
        if "local-AS" == arg:
           community_member += "NO_EXPORT_SUBCONFED"
           community_member += ","
        elif  "no-peer" == arg:
           community_member += "NOPEER"
           community_member += ","
        elif "no-export" == arg:
           community_member += "NO_EXPORT"
           community_member += ","
        elif "no-advertise" == arg:
           community_member += "NO_ADVERTISE"
           community_member += ","
        else:
           community_member += arg
           community_member += ","
    if member_exits:
       community_member = community_member[:-1]
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}/config/community-member={members}',community_list_name=args[5], members=community_member)
    else:
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}',community_list_name=args[5])
    return keypath

def generate_community_expanded_delete_keypath(args):
    member_exits = 0
    community_member = ''
    for arg in args[6:]:
        member_exits = 1
        community_member += "REGEX:"+arg
        community_member += ","

    if member_exits:
       community_member = community_member[:-1]
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}/config/community-member={members}',community_list_name=args[5], members=community_member)
    else:
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={community_list_name}',community_list_name=args[5])
    return keypath

def generate_extcommunity_standard_delete_keypath(args):
    member_exits = 0
    community_member = ''
    i = 6
    for arg in args[6:]:
        member_exits = 1
        if "soo" == arg:
           community_member += "route-origin:"+args[i+1]
           community_member += ","
        elif "rt" == arg:
           community_member += "route-target:"+args[i+1]
           community_member += ","
        i = i + 1
    if member_exits:
       community_member = community_member[:-1]
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={community_list_name}/config/ext-community-member={members}',community_list_name=args[5], members=community_member)
    else:
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={community_list_name}',community_list_name=args[5])
    return keypath

def generate_extcommunity_expanded_delete_keypath(args):
    member_exits = 0
    community_member = ''
    for arg in args[6:]:
        member_exits = 1
        community_member += "REGEX:"+arg
        community_member += ","

    if member_exits:
       community_member = community_member[:-1]
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={community_list_name}/config/ext-community-member={members}',community_list_name=args[5], members=community_member)
    else:
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={community_list_name}',community_list_name=args[5])
    return keypath

def generate_aspath_delete_keypath(args):
    paths_exits = 0
    paths = ''
    for arg in args[1:]:
        paths_exits = 1
        paths += arg
        paths += ","

    if paths_exits:
       paths = paths[:-1]
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={as_path_set_name}/config/as-path-set-member={path}',as_path_set_name=args[0], path=paths)
    else:
       keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={as_path_set_name}',as_path_set_name=args[0])
    return keypath


def invoke(func, args):
    body = None
    keypath = None
    aa = cc.ApiClient()

    #bgp-community-standard commands
    if func == 'bgp_community_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = generate_community_standard_body(args)
        return aa.patch(keypath,body)

    #bgp-community-expanded commands
    elif func == 'bgp_community_expanded':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        body = generate_community_expanded_body(args)
        return aa.patch(keypath,body)

    # Remove the bgp-community-standard set.
    elif func == 'bgp_community_standard_delete':
        keypath = generate_community_standard_delete_keypath(args)
        return aa.delete(keypath)

    # Remove the bgp-community-expanded set.
    elif func == 'bgp_community_expanded_delete':
        keypath = generate_community_expanded_delete_keypath(args)
        return aa.delete(keypath)

    #bgp-extcommunity-standard commands
    elif func == 'bgp_extcommunity_standard':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = generate_extcommunity_standard_body(args)
        return aa.patch(keypath,body)

    #bgp-extcommunity-expanded commands
    elif func == 'bgp_extcommunity_expanded':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        body = generate_extcommunity_expanded_body(args)
        return aa.patch(keypath,body)

    # Remove the bgp-extcommunity-standard set.
    elif func == 'bgp_extcommunity_standard_delete':
        keypath = generate_extcommunity_standard_delete_keypath(args)
        return aa.delete(keypath)

    # Remove the bgp-extcommunity-expanded set.
    elif func == 'bgp_extcommunity_expanded_delete':
        keypath = generate_extcommunity_expanded_delete_keypath(args)
        return aa.delete(keypath)

    # bgp-as-path-list command
    elif func == 'bgp_as_path_list':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets')
        body = {"openconfig-bgp-policy:as-path-sets":{"as-path-set":[{"as-path-set-name": args[0],"config":{"as-path-set-name":args[0],
                "as-path-set-member":[args[1]]}}]}}
        return aa.patch(keypath,body)

    # Remove the bgp-as-path-list set.
    elif func == 'bgp_as_path_list_delete':
        keypath = generate_aspath_delete_keypath(args)
        return aa.delete(keypath)

    elif func == 'bgp_community_show_all':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets')
        return aa.get(keypath)

    elif func == 'bgp_community_show_specific':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/community-sets/community-set={name}', name=args[1])
        return aa.get(keypath)

    elif func == 'bgp_ext_community_show_all':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets')
        return aa.get(keypath)

    elif func == 'bgp_ext_community_show_specific':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/ext-community-sets/ext-community-set={name}', name=args[1])
        return aa.get(keypath)

    elif func == 'bgp_aspath_show_specific':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets/as-path-set={name}', name=args[1])
        return aa.get(keypath)

    elif func == 'bgp_aspath_show_all':
        keypath = cc.Path('/restconf/data/openconfig-routing-policy:routing-policy/defined-sets/openconfig-bgp-policy:bgp-defined-sets/as-path-sets')
        return aa.get(keypath)
    else:
    	return aa.cli_not_implemented(func)


def run(func, args):

  try:
    response = invoke(func,args)

    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if api_response is None:
                print("Failed")
                return 
	    #print api_response
	    show_cli_output(args[0], api_response)
    else:
        print response.error_message()
	return
  except Exception as e:
    print "%Error: " + str(e)

  return


if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])

