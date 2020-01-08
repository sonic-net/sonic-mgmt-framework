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

import argparse
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

    if func == 'create_mirror_session':
        keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session/MIRROR_SESSION/MIRROR_SESSION_LIST={name}',
                name=args.session)
        body=collections.defaultdict(dict)
        entry = {
                "name": args.session,
                }

        if args.destination is not None:
            entry["dst_port"] = args.destination

        if args.source is not None:
            entry["src_port"] = args.source

        if args.direction is not None:
            entry["direction"] = args.direction

        if args.dst_ip is not None:
            entry["dst_ip"] = args.dst_ip

        if args.src_ip is not None:
            entry["src_ip"] = args.src_ip

        if args.dscp is not None:
            entry["dscp"] = int(args.dscp)

        if args.ttl is not None:
            entry["ttl"] = int(args.ttl)

        if args.gre is not None:
            entry["gre_type"] = args.gre

        if args.queue is not None:
            entry["queue"] = args.queue

        body["MIRROR_SESSION_LIST"]=[entry]

        return aa.patch(keypath, body)

    # Remove mirror session
    if func == 'delete_mirror_session':
        keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session/MIRROR_SESSION/MIRROR_SESSION_LIST={name}',
                name=args.session)
        return aa.delete(keypath)

    else:
        print("%Error: not implemented")
        exit(1)

def config(args):
    try:
        api_response = invoke(args.config, args)

        if api_response.ok():
            # Get Command Output
            response = api_response.content
            if response is None:
                print("Success")
            else:
                print("Failure")
        else:
            #error response
            print("Failed. Invalid mirror configuration. ")

    except:
            # system/network error
            raise

def show(args):
    try:
        # Get the rules of all ACL table entries.
        body = None
        aa = cc.ApiClient()
        if args.show == 'show_mirror_session':
            # Get mirror-session-info
            if args.session is not None:
                keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session/MIRROR_SESSION/MIRROR_SESSION_LIST={name}',
                        name=args.session)
            else:
                keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session')
            response = aa.get(keypath)
            gbl_oper_dict = {}
            session_list = 0
            if response.ok() and 'sonic-mirror-session:sonic-mirror-session' in response.content.keys():
                value = response['sonic-mirror-session:sonic-mirror-session']
                if 'MIRROR_SESSION' in value.keys():
                    list = value['MIRROR_SESSION']
                    if 'MIRROR_SESSION_LIST' in list.keys():
                        session_list = list['MIRROR_SESSION_LIST']
                    else:
                        print("No sessions configured")
                        return
                else:
                    print("No sessions configured")
                    return
            elif response.ok() and 'sonic-mirror-session:MIRROR_SESSION_LIST' in response.content.keys():
                session_list = response['sonic-mirror-session:MIRROR_SESSION_LIST']
            else:
                print("No sessions configured")
                return

            # Retrieve mirror session status.
            if args.session is not None:
                keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session/MIRROR_SESSION_TABLE/MIRROR_SESSION_TABLE_LIST={name}',
                        name=args.session)
            else:
                keypath = cc.Path('/restconf/data/sonic-mirror-session:sonic-mirror-session/MIRROR_SESSION_TABLE')
            response = aa.get(keypath)
            session_status = 0
            if response.ok() and 'sonic-mirror-session:MIRROR_SESSION_TABLE' in response.content.keys():
                value = response['sonic-mirror-session:MIRROR_SESSION_TABLE']
                if 'MIRROR_SESSION_TABLE_LIST' in value.keys():
                    session_status = value['MIRROR_SESSION_TABLE_LIST']
                else:
                    print("Session state info not found")
                    return
            elif response.ok() and 'sonic-mirror-session:MIRROR_SESSION_TABLE_LIST' in response.content.keys():
                session_status = response['sonic-mirror-session:MIRROR_SESSION_TABLE_LIST']
            else:
                print("Session state info not found")
                return
            final_dict = {}
            final_dict['session_list'] = session_list
            final_dict['session_status'] = session_status
            show_cli_output(args.renderer, final_dict)
    except:
            # system/network error
            raise

    return

if __name__ == '__main__':
    pipestr().write(sys.argv)

    parser = argparse.ArgumentParser(description='Handles MIRROR commands',
                                     version='1.0.0',
                                     formatter_class=argparse.RawTextHelpFormatter,
                                     epilog="""
Examples:
    mirror -config -deviceid value
    mirror -config -collector collectorname -iptype ipv4/ipv6 -ip ipaddr -port value
    mirror -clear -device_id
    mirror -clear -collector collectorname
    mirror -show -device_id
    mirror -show -collector collectorname
""")

    parser.add_argument('-clear', '--clear', type=str, help='Clear mirror information')
    parser.add_argument('-show', '--show', type=str, help='Show mirror information')
    parser.add_argument('-renderer', '--renderer', type=str, help='Show mirror information')
    parser.add_argument('-config', '--config', type=str, help='Config mirror information')
    parser.add_argument('-session', '--session', type=str, help='mirror session name')
    parser.add_argument('-destination', '--destination', help='destination port')
    parser.add_argument('-source', '--source', type=str, help='mirror source port')
    parser.add_argument('-direction', '--direction', type=str, help='mirror direction')
    parser.add_argument('-dst_ip', '--dst_ip', help='ERSPAN destination ip address')
    parser.add_argument('-src_ip', '--src_ip', help='ERSPAN source ip address')
    parser.add_argument('-dscp', '--dscp', help='ERSPAN dscp')
    parser.add_argument('-gre', '--gre', help='ERSPAN gre')
    parser.add_argument('-ttl', '--ttl', help='ERSPAN ttl')
    parser.add_argument('-queue', '--queue', help='ERSPAN ttl')

    args = parser.parse_args()
    if args.config:
        config(args)
    elif args.show:
        show(args)


