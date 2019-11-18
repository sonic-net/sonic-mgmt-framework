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
import cli_client as cc
from rpipe_utils import pipestr
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

    
def invoke_api(func, args=[]):
    api = cc.ApiClient()

    if func == 'get_sonic_portchannel_sonic_portchannel_lag_table':
        path = cc.Path('/restconf/data/sonic-portchannel:sonic-portchannel/LAG_TABLE')
        return api.get(path)

    if func == 'get_sonic_portchannel_sonic_portchannel_lag_table_lag_table_list':
        path = cc.Path('/restconf/data/sonic-portchannel:sonic-portchannel/LAG_TABLE/LAG_TABLE_LIST={lagname}', lagname=args[0])
        return api.get(path)

    if func == 'get_openconfig_lacp_lacp_interfaces':
        path = cc.Path('/restconf/data/openconfig-lacp:lacp/interfaces')
        return api.get(path)
        
    if func == 'get_openconfig_lacp_lacp_interfaces_interface':
        path = cc.Path('/restconf/data/openconfig-lacp:lacp/interfaces/interface={name}', name=args[0])
        return api.get(path)        
            
    if func == 'get_openconfig_interfaces_interfaces_interface_state_counters':
        path = cc.Path('/restconf/data/openconfig-interfaces:interfaces/interface={name}/state/counters', name=args[0])
        return api.get(path)

    #print "invoke_api done"
    return api.cli_not_implemented(func)
    

def get_lag_data():

    api_response = {}
    output = {}

    try:  
        if sys.argv[1] == "get_all_portchannels":
            portchannel_func = 'get_sonic_portchannel_sonic_portchannel_lag_table'
        else :
            portchannel_func = 'get_sonic_portchannel_sonic_portchannel_lag_table_lag_table_list'
            
        args = sys.argv[2:]
        
        response = invoke_api(portchannel_func, args)
        if response.ok():
            if response.content is not None:
                # Get Command Output
                api_response = response.content
                #print "--------------------------------------------------", api_response
                #api_response = aa.api_client.sanitize_for_serialization(api_response)
                #print "-----------------------", api_response
                if 'sonic-portchannel:LAG_TABLE' not in api_response.keys():
                    output['sonic-portchannel:LAG_TABLE'] = {}
                    output['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST'] = api_response['sonic-portchannel:LAG_TABLE_LIST']
                else:
                    output = api_response        
        
    except Exception as e:
        print("Exception when calling get_lag_data : %s\n" %(e))
    
    return output

def get_lacp_data():

    api_response1 = {}
    resp = {}
    
    try:
        if sys.argv[1] == "get_all_portchannels":
            lacp_func = 'get_openconfig_lacp_lacp_interfaces'
        else :
            lacp_func = 'get_openconfig_lacp_lacp_interfaces_interface'

        args = sys.argv[2:]
        
        response = invoke_api(lacp_func, args)
        if response.ok():
            if response.content is not None:
                # Get Command Output
                api_response1 = response.content
                #api_response1 = aa1.api_client.sanitize_for_serialization(api_response1)
                if 'openconfig-lacp:interfaces' not in api_response1.keys():
                    resp['openconfig-lacp:interfaces'] = {}
                    resp['openconfig-lacp:interfaces']['interface'] = api_response1['openconfig-lacp:interface']
                else:
                     resp = api_response1
                #print "------------------------------------------------", resp        
                
    except Exception as e:
        print("Exception when calling get_lacp_data : %s\n" %(e))
    
    return resp

    
def get_counters(api_response):

    try:
        #print api_response        
        response = {}
        if 'sonic-portchannel:LAG_TABLE' not in api_response.keys():
            response['sonic-portchannel:LAG_TABLE'] = {}
            response['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST'] = api_response['sonic-portchannel:LAG_TABLE_LIST']
        else:
            response = api_response
        
        for po_intf in response['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST']:        
            resp = invoke_api('get_openconfig_interfaces_interfaces_interface_state_counters', po_intf['lagname'])
        if resp.ok():
            if resp.content is not None:
                # Get Command Output
                resp = resp.content
                #resp = aa3.api_client.sanitize_for_serialization(resp)
                po_intf["counters"] = resp

    except Exception as e:
        print("Exception when calling get_counters : %s\n" %(e))
    

def run():
    
    api_response = get_lag_data()
    api_response1 = get_lacp_data()
    get_counters(api_response)
    

    # Combine Outputs
    response = {"portchannel": api_response, "lacp": api_response1}
    #print response
    
    if sys.argv[1] == "get_all_portchannels":
        template_file = sys.argv[2]
    else:
        template_file = sys.argv[3]

    #print response
    show_cli_output(template_file, response)


if __name__ == '__main__':

    pipestr().write(sys.argv)
    run()       

