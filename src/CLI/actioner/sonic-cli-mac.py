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

## TODO: Update with network instance l2 yang

import sys
import time
import json
import ast
import openconfig_interfaces_client #update with network instance client
from rpipe_utils import pipestr
from openconfig_interfaces_client.rest import ApiException 
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

# Update with network instance API
def generate_body(func, args):
    body = None
    if func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_description':
       keypath = [ args[0] ]
       body = { "openconfig-interfaces:description": args[1] }
    elif func.__name__ == 'patch_openconfig_interfaces_interfaces_interface_config_enabled':
       keypath = [ args[0] ]
       if args[1] == "True":
           body = { "openconfig-interfaces:enabled": True }
       else:
           body = { "openconfig-interfaces:enabled": False }
    else:
       body = {}

    return keypath,body

def run(func, args):

    c = openconfig_interfaces_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_interfaces_client.OpenconfigInterfacesApi(api_client=openconfig_interfaces_client.ApiClient(configuration=c))

    # create a body block
    #keypath, body = generate_body(func, args)

    try:
        if func.__name__ == 'get_openconfig_interfaces_interfaces':
            if args[1] == 'show':
                api_response ={"1":{"VLAN":"1" , "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet0"},
                               "2":{"VLAN":"2" , "MAC":"20:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet4"},
                               "3":{"VLAN":"3" , "MAC":"30:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet8"},
                               "4":{"VLAN":"4" , "MAC":"40:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet12"}}

            elif args[1] == 'mac-addr':
                api_response = {"1":{"VLAN":"1" , "MAC":args[2], "Type":"dynamic", "Interface":"Ethernet0"}}

            elif args[1] == 'vlan':
                api_response = {args[2]:{"VLAN":args[2] , "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet0"}}
 
            elif args[1] == 'interface':
                api_response = {"1":{"VLAN":"1", "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":args[2]}}

            elif args[1] == 'static':
                if args[2] == 'address':
                    api_response = {"1":{"VLAN":"1" , "MAC":args[3], "Type":"static", "Interface":"Ethernet0"}}
                elif args[2] == 'vlan':
                    api_response = {args[3]:{"VLAN":args[3] , "MAC":"90:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet0"}}
                elif args[2] == 'interface':
                    api_response = {"1":{"VLAN":"1" , "MAC":"90:b1:1c:f4:9d:83", "Type":"static", "Interface":args[3]}}
                else:
                    api_response ={"1":{"VLAN":"1" , "MAC":"90:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet0"},
                                   "2":{"VLAN":"2" , "MAC":"20:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet4"},
                                   "3":{"VLAN":"3" , "MAC":"30:b1:1c:f4:9d:83", "Type":"static", "Interface":"Ethernet8"}}

            elif args[1] == 'dynamic':
                if args[2] == 'address':
                    api_response = {"1":{"VLAN":"1" , "MAC":args[3], "Type":"dynamic", "Interface":"Ethernet0"}}
                elif args[2] == 'vlan':
                    api_response = {args[3]:{"VLAN":args[3] , "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet0"}}
                elif args[2] == 'interface':
                    api_response = {"1":{"VLAN":"1" , "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":args[3]}}
                else:
                    api_response ={"1":{"VLAN":"1" , "MAC":"90:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet0"},
                                   "2":{"VLAN":"2" , "MAC":"20:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet4"},
                                   "3":{"VLAN":"3" , "MAC":"30:b1:1c:f4:9d:83", "Type":"dynamic", "Interface":"Ethernet8"}}
            elif args[1] == 'count':
                api_response = {"1":{"all-mac":"1","dynamic-mac":"1","static-mac":"1","total-mac":"3"}}
            
            show_cli_output(args[0], api_response)
            return

    except ApiException as e:
        #print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func.__name__, e))
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
        else:
            print "Failed"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_interfaces_client.OpenconfigInterfacesApi.__dict__) #Update with network instance API

    run(func, sys.argv[2:])
