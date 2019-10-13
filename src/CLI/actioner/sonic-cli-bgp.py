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
from rpipe_utils import pipestr
from openconfig_network_instance_client.rest import ApiException
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
    # Get the rules of all ACL table entries.
    if func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_config_router_id':
	    keypath = [args[0], 0]
    else:
       body = {}

    return keypath,body

def run(func, args):

    c = openconfig_network_instance_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_network_instance_client.OpenconfigNetworkInstanceApi(api_client=openconfig_network_instance_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    if func.__name__ == 'patch_openconfig_network_instance_network_instances_network_instance_config_router_id':
        print args[0], args[1], args[2], args[3]
        return

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_network_instance_client.OpenconfigNetworkInstanceApi.__dict__)

    run(func, sys.argv[2:])
