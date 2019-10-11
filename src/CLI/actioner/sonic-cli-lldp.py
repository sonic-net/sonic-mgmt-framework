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
import openconfig_lldp_client
from rpipe_utils import pipestr
from openconfig_lldp_client.rest import ApiException
from scripts.render_cli import show_cli_output


import urllib3
urllib3.disable_warnings()
plugins = dict()


def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func

def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    body = None
    if func.__name__ == 'get_openconfig_lldp_lldp_interfaces':
        keypath = []
    elif func.__name__ == 'get_openconfig_lldp_lldp_interfaces_interface':
        keypath = [args[1]]
    else:
       body = {}

    return keypath,body


def run(func, args):
    c = openconfig_lldp_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_lldp_client.OpenconfigLldpApi(api_client=openconfig_lldp_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(aa,func.__name__)(*keypath)
        if api_response is None:
            print ("Success")
        else:
            response = api_response.to_dict()
            if 'openconfig_lldpinterfaces' in response.keys():
                if not response['openconfig_lldpinterfaces']:
                    return
		neigh_list = response['openconfig_lldpinterfaces']['interface']
                if neigh_list is None:
                    return
		show_cli_output(sys.argv[2],neigh_list)
            elif 'openconfig_lldpinterface' in response.keys():
	        neigh = response['openconfig_lldpinterface']#[0]['neighbors']['neighbor']
                if neigh is None:
                    return
		if sys.argv[3] is not None:
		    if neigh[0]['neighbors']['neighbor'][0]['state'] is None:
			print('No LLDP neighbor interface')
		    else:
			show_cli_output(sys.argv[2],neigh)
		else:
	     	    show_cli_output(sys.argv[2],neigh)
            else:
                print("Failed")
    except ApiException as e:
        print("Exception when calling OpenconfigLldpApi->%s : %s\n" %(func.__name__, e))

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_lldp_client.OpenconfigLldpApi.__dict__)
    run(func, sys.argv[2:])
