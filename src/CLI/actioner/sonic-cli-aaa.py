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
import urllib3
urllib3.disable_warnings()

def invoke_api(func, args):
    body = None
    api = cc.ApiClient()

    # Set/Get aaa configuration
    body = { "openconfig-system-ext:failthrough": False, "openconfig-system-ext:authentication-method": 'None' }
    failthrough='None'
    authmethod=[]

    # authentication-method is a leaf-list. So patch is not supported. A put opeartion
    # would clear existing other parameters as well. So reading existing contents and
    # trying to change only the user input parameter with a put

    path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/config')
    get_response = api.get(path)
    if get_response.ok():
        if get_response.content:
            api_response = get_response.content
            if 'failthrough' in api_response['openconfig-system:config']:
                body["openconfig-system-ext:failthrough"] = api_response['openconfig-system:config']['failthrough']
            if 'authentication-method' in api_response['openconfig-system:config']:
                body["openconfig-system-ext:authentication-method"] = api_response['openconfig-system:config']['authentication-method']
    if func == 'put_openconfig_system_ext_system_aaa_authentication_config_failthrough':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/config/openconfig-system-ext:failthrough')
       body["openconfig-system-ext:failthrough"] = (args[0] == "True")
       return api.put(path, body)
    elif func == 'put_openconfig_system_system_aaa_authentication_config_authentication_method':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/config/authentication-method')
       # tricky logic: xml sends frist selection and values of both local and tacacs+ params
       # when user selects "local tacacs+", actioner receives "local local tacacs+"
       # when user selects "tacacs+ local", actioner receives "tacacs+ local tacacs+"

       authmethod.append(args[0])
       if len(args) == 3:
           if args[0] == args[1]:
               authmethod.append(args[2])
           else:
               authmethod.append(args[1])
       else:
           pass
       body["openconfig-system-ext:authentication-method"] = authmethod
       return api.put(path, body)
    elif func == 'get_openconfig_system_system_aaa_authentication_config':
       return get_response
    else:
       body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content

            if api_response is None:
                print("%Error: Transaction Failure")
            elif func == 'get_openconfig_system_system_aaa_authentication_config':
                show_cli_output(args[0], api_response)
            else:
                return
    else:
        print(response.error_message())

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])

