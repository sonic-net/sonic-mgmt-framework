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
import openconfig_network_instance_client
from  openconfig_network_instance_client.rest import ApiException
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
    # Get the all vrfs.
    if func.__name__ == 'get_list_openconfig_network_instance_network_instances_network_instance':
       keypath = []

    # Get a specific vrf. 
    elif func.__name__ == 'get_openconfig_network_instance_network_instances_network_instance':
       keypath = [ args[0] ]

    # Configure management vrf. maybe just set mgmt here without args[0] 
    elif func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance' :
       keypath = [ args[0] args[1] args[2] ]
       body = { "openconfig-network-instance:config": {
                   "name": args[0],
                   "type": args[1],
                   "enabled": args[2]
                   "description": ""
                 }
              }

    # Delete management vrf.
    elif func.__name__ == 'delete_openconfig_network_instance_network_instances_network_instance':
       keypath = [ args[0] args[1] args[2] ]
       body = { "openconfig-network-instance:config": {
                   "name": args[0],
                   "type": args[1],
                   "enabled": args[2]
                   "description": ""
                 }
              }
    
    else:
        body = {}

    if body is not None:
        body = json.dumps(body,ensure_ascii=False, indent=4, separators=(',', ':'))
        return keypath, ast.literal_eval(body)
    else:
        return keypath, body


def run(func, args):

    c = openconfig_network_instance_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_network_instance_client.OpenconfigNetworkInstanceApi(api_client=openconfig_network_instance_client.ApiClient(configuration=c))

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
            if 'openconfig_network_instancenetwork_instance' in response.keys():
                value = response['openconfig_network_instancenetwork_instance']
                if value is None:
                    print("Success")
                else:
                    print ("Failed")
            else:
                print("Failed")

    except ApiException as e:
        print("Exception when calling OpenconfigNetworkInstanceApi->%s : %s\n" %(func.__name__, e))
        if e.body != "":
            body = json.loads(e.body)
            if "ietf-restconf:errors" in body:
                 err = body["ietf-restconf:errors"]
                 if "error" in err:
                     errList = err["error"]

                     errDict = {}
                     for dict in errList:
                         for k, v in dict.iteritems():
                              errDict[k] = v

                     if "error-message" in errDict:
                         print "%Error: " + errDict["error-message"]
                         return
                     print "%Error: Transaction Failure"
                     return
            print "%Error: Transaction Failure"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    #pdb.set_trace()
    func = eval(sys.argv[1], globals(), openconfig_network_instance_client.OpenconfigNetworkInstanceApi.__dict__)
    run(func, sys.argv[2:])
                                                                                          
