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
import json
import collections
import re
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    #######################################
    # Configure  MCLAG Domain Table - START
    #######################################

    #[un]configure local IP Address
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_source_ip' or
        func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_source_ip'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}/source_ip', domain_id=args[0])

        if (func.startswith("patch") is True):
            body = {
                "sonic-mclag:source_ip": args[1]
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #[un]configure Peer IP Address
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_peer_ip' or
        func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_peer_ip'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}/peer_ip', domain_id=args[0])

        if (func.startswith("patch") is True):
            body = {
                "sonic-mclag:peer_ip": args[1]
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #[un]configure Peer Link
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_peer_link' or
        func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_peer_link'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}/peer_link', domain_id=args[0])

        if (func.startswith("patch") is True):
            body = {
                "sonic-mclag:peer_link": args[1]
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #[un]configure Keepalive interval 
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_keepalive_interval' or
        func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_keepalive_interval'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}/keepalive_interval', domain_id=args[0])

        if (func.startswith("patch") is True):
            body = {
                "sonic-mclag:keepalive_interval": int(args[1])
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #configure session Timeout
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_session_timeout' or
        func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list_session_timeout'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}/session_timeout', domain_id=args[0])

        if (func.startswith("patch") is True):
            body = {
                "sonic-mclag:session_timeout": int(args[1])
            }
            return aa.patch(keypath, body)
        else:
            return aa.delete(keypath)

    #delete MCLAG Domain 
    if (func == 'delete_sonic_mclag_sonic_mclag_mclag_domain_mclag_domain_list'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_DOMAIN/MCLAG_DOMAIN_LIST={domain_id}', domain_id=args[0])
        return aa.delete(keypath)

    

    #######################################
    # Configure  MCLAG Domain Table - END
    #######################################


    #######################################
    # Configure  MCLAG Interface Table - START
    #######################################
    if (func == 'patch_sonic_mclag_sonic_mclag_mclag_interface_mclag_interface_list'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_INTERFACE/MCLAG_INTERFACE_LIST={domain_id},{if_name}',
                domain_id=args[0], if_name=args[1])
        body = {
            "sonic-mclag:MCLAG_INTERFACE_LIST": [
            {
                "domain_id":int(args[0]),
                "if_name":args[1],
                "if_type":"PortChannel"
            }
          ]
        }
        return aa.patch(keypath, body)

    if (func == 'delete_sonic_mclag_sonic_mclag_mclag_interface_mclag_interface_list'):
        keypath = cc.Path('/restconf/data/sonic-mclag:sonic-mclag/MCLAG_INTERFACE/MCLAG_INTERFACE_LIST={domain_id},{if_name}',
                domain_id=(args[0]), if_name=args[1])
        return aa.delete(keypath)

    #######################################
    # Configure  MCLAG Domain Table - END
    #######################################


    else:
        print("%Error: not implemented")
        exit(1)

def run(func, args):
    try:
        api_response = invoke(func, args)

        if api_response.ok():
            print "ok...."
            response = api_response.content
            if response is None:
                print "Success"
            elif 'sonic-mclag:sonic-mclag' in response.keys():
                value = response['sonic-mclag:sonic-mclag']
                if value is None:
                    print("Success")
                else:
                    print ("Failed")
            
        else:
            #error response
            print api_response.error_message()

    except:
            # system/network error
            print "%Error: Transaction Failure"


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

