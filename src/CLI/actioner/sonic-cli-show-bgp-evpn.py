#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom, Inc.
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

vniDict = {}

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None
    
    if func == 'get_bgp_evpn_vni':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/state',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        return api.get(keypath)

    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            api_response = response.content
            show_cli_output(args[3], api_response)
        else:
            print("Empty response")
    else:
        print(response.error_message())

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

