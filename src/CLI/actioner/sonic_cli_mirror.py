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
import traceback
import json
import collections
import re
import cli_client as cc
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

        if args.destination is not '':
            if args.destination != "erspan":
                entry["dst_port"] = args.destination

        if args.source is not '':
            entry["src_port"] = args.source

        if args.direction is not '':
            entry["direction"] = args.direction

        if args.destination == "erspan":
            if args.dst_ip is not '':
                entry["dst_ip"] = args.dst_ip

            if args.src_ip is not '':
                entry["src_ip"] = args.src_ip

            if args.dscp is not '':
                entry["dscp"] = int(args.dscp)

            if args.ttl is not '':
                entry["ttl"] = int(args.ttl)

            if args.gre is not '':
                entry["gre_type"] = args.gre

            if args.queue is not '':
                entry["queue"] = int(args.queue)

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

def show(args):
    body = None
    aa = cc.ApiClient()

    # Get mirror-session-info
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

    if args.session is not '':
        session_found = "0"
        for session in session_list:
            if session['name'] == args.session:
                session_list = [ session ]
                session_found = "1"
                break
        for session in session_status:
            if session['name'] == args.session:
                session_status = [ session ]
                break
        if session_found == "0":
            print("No sessions configured")
            return

    final_dict = {}
    final_dict['session_list'] = session_list
    final_dict['session_status'] = session_status
    show_cli_output(args.renderer, final_dict)
    return


def run(func, in_args=None):
    try:
        parser = argparse.ArgumentParser(description='Handles MIRROR commands', prog='MIRROR')
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

        args = parser.parse_args(args=in_args)
        if func == "show_mirror_session":
            show(args)
            return

        api_response = invoke(func, args)
        if api_response.ok():
            # Get Command Output
            response = api_response.content
            if response is None:
                print("Success")
            else:
                print("Failure")
        else:
            #error response
            print(api_response.error_message())

    except:
            traceback.print_exc()


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

