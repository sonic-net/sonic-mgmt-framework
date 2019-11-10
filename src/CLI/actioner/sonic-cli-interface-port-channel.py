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
import openconfig_interfaces_client
import openconfig_lacp_client
import sonic_portchannel_client
from sonic_portchannel_client.api.sonic_portchannel_api import SonicPortchannelApi
from sonic_portchannel_client.rest import ApiException
from openconfig_interfaces_client.rest import ApiException
from rpipe_utils import pipestr
from openconfig_lacp_client.rest import ApiException
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
    

def generate_key(func, args):
    keypath = []
    if func.__name__ == "get_sonic_portchannel_sonic_portchannel_lag_table" or func.__name__ == "get_openconfig_lacp_lacp_interfaces' or func.__name__ == 'get_openconfig_interfaces_interfaces":
        keypath = []
    elif func.__name__ == "get_sonic_portchannel_sonic_portchannel_lag_table_lag_table_list" or func.__name__ == "get_openconfig_lacp_lacp_interfaces_interface" or func.__name__ == "get_openconfig_interfaces_interfaces_interface_state_counters":
        keypath = [args[0]]

    return keypath

def print_exception(e):
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
    
    
def get_lag_data():

    api_response = {}

    try:        
        
        c = sonic_portchannel_client.Configuration()
        c.verify_ssl = False
        aa = sonic_portchannel_client.SonicPortchannelApi(api_client=sonic_portchannel_client.ApiClient(configuration=c))
        
        if sys.argv[1] == "get_all_portchannels":
            portchannel_func = 'get_sonic_portchannel_sonic_portchannel_lag_table'
        else :
            portchannel_func = 'get_sonic_portchannel_sonic_portchannel_lag_table_lag_table_list'

        func = eval(portchannel_func, globals(), sonic_portchannel_client.SonicPortchannelApi.__dict__)
        args = sys.argv[2:]

        keypath = generate_key(func, args)
        
        if len(keypath) != 0:
           api_response = getattr(aa,func.__name__)(*keypath)
        else :
           api_response = getattr(aa,func.__name__)()


        if api_response is None:
            print ("Failure in getting portchannel data")
        else:
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            #print "-----------------------", api_response
            output = {}
            if 'sonic-portchannel:LAG_TABLE' not in api_response.keys():
                output['sonic-portchannel:LAG_TABLE'] = {}
                output['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST'] = api_response['sonic-portchannel:LAG_TABLE_LIST']
            else:
                output = api_response
        
    except ApiException as e:
        #print("Exception when calling SonicPortchannelApi->%s : %s\n" %(func.__name__, e))
        print_exception(e)       
    
    return output

def get_lacp_data():

    api_response1 = {}
    
    try:
        c1 = openconfig_lacp_client.Configuration()
        c1.verify_ssl = False
        aa1 = openconfig_lacp_client.OpenconfigLacpApi(api_client=openconfig_lacp_client.ApiClient(configuration=c1))
        
        # create a body block
        if sys.argv[1] == "get_all_portchannels":
            lacp_func = 'get_openconfig_lacp_lacp_interfaces'
        else :
            lacp_func = 'get_openconfig_lacp_lacp_interfaces_interface'

        func1 = eval(lacp_func, globals(), openconfig_lacp_client.OpenconfigLacpApi.__dict__)
        args = sys.argv[2:]

        keypath1 = generate_key(func1, args)
        
        if len(keypath1) != 0:
            api_response1 = getattr(aa1,func1.__name__)(*keypath1)
        else :
            api_response1 = getattr(aa1,func1.__name__)()

        if api_response1 is None:
            print ("Failure in getting LACP data")
        else:
            # Get Command Output
            api_response1 = aa1.api_client.sanitize_for_serialization(api_response1)
            resp = {}
            if 'openconfig-lacp:interfaces' not in api_response1.keys():
                resp['openconfig-lacp:interfaces'] = {}
                resp['openconfig-lacp:interfaces']['interface'] = api_response1['openconfig-lacp:interface']
            else:
                 resp = api_response1
            #print "------------------------------------------------", resp
                
    except ApiException as e:
        #print("Exception when calling OpenconfigLacpApi->%s : %s\n" %(func.__name__, e))
        print_exception(e)       
    
    return resp

    
def get_counters(api_response):

    try:
        c3 = openconfig_interfaces_client.Configuration()
        c3.verify_ssl = False
        aa3 = openconfig_interfaces_client.OpenconfigInterfacesApi(api_client=openconfig_interfaces_client.ApiClient(configuration=c3))


        # create a body block
        func3 = eval('get_openconfig_interfaces_interfaces_interface_state_counters', globals(), openconfig_interfaces_client.OpenconfigInterfacesApi.__dict__)
        args = sys.argv[2:]

        response = {}
        if 'sonic-portchannel:LAG_TABLE' not in api_response.keys():
            response['sonic-portchannel:LAG_TABLE'] = {}
            response['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST'] = api_response['sonic-portchannel:LAG_TABLE_LIST']
        else:
            response = api_response
        
        for po_intf in response['sonic-portchannel:LAG_TABLE']['LAG_TABLE_LIST']:
            resp = getattr(aa3,func3.__name__)(po_intf['lagname'])
            if resp is None:
                print ("Failure in getting PortChannel counters data")
            else:
                resp = aa3.api_client.sanitize_for_serialization(resp)
                po_intf["counters"] = resp


    except ApiException as e:
        #print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func.__name__, e))
        print_exception(e)       
    

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

    show_cli_output(template_file, response)


if __name__ == '__main__':

    pipestr().write(sys.argv)
    run()       

