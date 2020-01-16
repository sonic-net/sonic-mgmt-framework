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
from scripts.render_cli import show_cli_output
import cli_client as cc
from crypt import crypt
import base64
import os

import urllib3
urllib3.disable_warnings()


plugins = dict()

def hashed_pw(pw):
    salt = base64.b64encode(os.urandom(6), './')
    return crypt(pw, '$6$' + salt)
    
def util_capitalize(value):
    for key,val in value.items():
        temp = key.split('_')
        alt_key = ''
        for i in temp:
        	alt_key = alt_key + i.capitalize() + ' '
        value[alt_key]=value.pop(key)
    return value

def system_state_key_change(value):
    if 'motd_banner' in value:
        value.pop('motd_banner')
    if 'login_banner' in value:
        value.pop('login_banner')
    return util_capitalize(value)


def memory_key_change(value):
    if 'physical' in value:
        value['Total']=value.pop('physical')
    if 'reserved' in value:
        value['Used']=value.pop('reserved')
    return value

def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func


def call_method(name, args):
    method = plugins[name]
    return method(args)

def invoke(func, args):
    body = None
    aa = cc.ApiClient()
    if func == 'get_openconfig_system_system_state':
        path = cc.Path('/restconf/data/openconfig-system:system/state')
	return aa.get(path)

    elif func == 'get_openconfig_system_system_clock':
        path = cc.Path('/restconf/data/openconfig-system:system/clock')
	return aa.get(path)
        
    elif func == 'get_openconfig_system_system_memory':
        path = cc.Path('/restconf/data/openconfig-system:system/memory')
	return aa.get(path)
        
    elif func == 'get_openconfig_system_system_cpus':
        path = cc.Path('/restconf/data/openconfig-system:system/cpus')
	return aa.get(path)

    elif func == 'get_openconfig_system_system_processes':
        path = cc.Path('/restconf/data/openconfig-system:system/processes')
	return aa.get(path)

    elif func == 'patch_openconfig_system_system_aaa_authentication_users_user':
        body =  { "openconfig-system:user": [{"username": args[0],
					     "config": {
 							 "username": args[0],
        						 "password": "",
        						 "password-hashed": hashed_pw(args[1]),
                                                         "ssh-key": "",
                                                         "role": args[2]
                                                        }
                                             }
                                           ]
                }
        path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/users/user={username}',username=args[0])
	return aa.patch(path,body)
    elif func == 'delete_openconfig_system_system_aaa_authentication_users_user':
        path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/users/user={username}',username=args[0])
        return aa.delete(path)


def run(func, args):
    api_response = invoke(func,args)
    if api_response.ok():
        if api_response.content is not None:
            response = api_response.content
            if 'openconfig-system:state' in response.keys():
                value = response['openconfig-system:state']
                if value is None:
                    return
                show_cli_output(sys.argv[2], system_state_key_change(value))

            elif 'openconfig-system:memory' in response.keys():
                value = response['openconfig-system:memory']
                if value is None:
                    return
		show_cli_output(sys.argv[2], memory_key_change(value['state']))
            elif 'openconfig-system:cpus' in response.keys():
                value = response['openconfig-system:cpus']
                if value is None:
                    return
                show_cli_output(sys.argv[2], value['cpu'])
            elif 'openconfig-system:processes' in response.keys():
                value = response['openconfig-system:processes']
		if 'pid' not in sys.argv:
                    if value is None:
                    	return
	    	    show_cli_output(sys.argv[2],value['process'])
	        else:
		    for proc in value['process']:
		        if proc['pid'] == int(sys.argv[3]):
		            show_cli_output(sys.argv[2],util_capitalize(proc['state']))
		            return
    else:
        print(api_response.error_message())


if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])

