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

import urllib3
urllib3.disable_warnings()


def invoke_api(func, args=[]):
    api = cc.ApiClient()

    if func == 'patch_openconfig_nat_nat_instances_instance_config_enable':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/enable', id=args[0])
        if args[1] == "True":
           body = { "openconfig-nat:enable": True }
        else:
           body = { "openconfig-nat:enable": False }
        return api.patch(path,body)

    elif func == 'patch_openconfig_nat_nat_instances_instance_config_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/timeout', id=args[0])
        body = { "openconfig-nat:timeout":  int(args[1]) }
        return api.patch(path, body)

    elif func == 'patch_openconfig_nat_nat_instances_instance_config_tcp_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/tcp-timeout', id=args[0])
        body = { "openconfig-nat:tcp-timeout":  int(args[1]) }
        return api.patch(path, body)

    elif func == 'patch_openconfig_nat_nat_instances_instance_config_udp_timeout':
        path = cc.Path('/restconf/data/openconfig-nat:nat/instances/instance={id}/config/udp-timeout', id=args[0])
        body = { "openconfig-nat:udp-timeout":  int(args[1]) }
        return api.patch(path, body)

    return api.cli_not_implemented(func)

def run(func, args):   

    try:
       args.insert(0,"0")  # NAT instance 0
       response = invoke_api(func, args)    
       if response.ok():
           if response.content is not None:
               # Get Command Output
               api_response = response.content
            
               if api_response is None:
                  print("Failed")
               else:
                  show_cli_output(args[0], api_response)
       else:
           print response.error_message()

        
    except Exception as e:
        print("Exception when calling OpenconfigNatApi->%s : %s\n" %(func, e))

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

