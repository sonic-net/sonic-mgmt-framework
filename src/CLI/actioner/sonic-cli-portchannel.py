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
import collections
from rpipe_utils import pipestr
import sonic_portchannel_client
from sonic_portchannel_client.api.sonic_portchannel_api import SonicPortchannelApi  
from sonic_portchannel_client.rest import ApiException
import sonic_port_client
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

pcDict = {}
memberDict = {}

plugins = dict()

def register(func):
    plugins[func.__name__] = func
    return func


def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    body = None
    if func.__name__ == 'get_sonic_portchannel_sonic_portchannel_lag_table':
        keypath = []
    else:
       body = {}

    return keypath,body

def run(func, args):

    c = sonic_portchannel_client.Configuration()
    c2 = sonic_port_client.Configuration()
    c.verify_ssl = False
    c2.verify_ssl = False
    aa = sonic_portchannel_client.SonicPortchannelApi(api_client=sonic_portchannel_client.ApiClient(configuration=c))
    aa2 = sonic_port_client.SonicPortApi(api_client=sonic_port_client.ApiClient(configuration=c2))

    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            laglst =[]
            if 'sonic-portchannel:LAG_TABLE' in api_response:
                value = api_response['sonic-portchannel:LAG_TABLE']
                if 'LAG_TABLE_LIST' in value:
                    laglst = value['LAG_TABLE_LIST']
            if api_response is None:
                print("Failed")
            else:
                if func.__name__ == 'get_sonic_portchannel_sonic_portchannel_lag_table':
                    memlst=[]
                    # Get members for all PortChannels
                    api_response_members = getattr(aa,'get_sonic_portchannel_sonic_portchannel_lag_member_table')(*keypath)
                    api_response_members = aa.api_client.sanitize_for_serialization(api_response_members)
                    if 'sonic-portchannel:LAG_MEMBER_TABLE' in api_response_members:
                        memlst = api_response_members['sonic-portchannel:LAG_MEMBER_TABLE']['LAG_MEMBER_TABLE_LIST']
                    for pc_dict in laglst:
                        pc_dict['members']=[]
                        pc_dict['type']="Eth"
                        for mem_dict in memlst:
                            if mem_dict['name'] == pc_dict['lagname']:
                                keypath = [mem_dict['ifname']]
                                pc_dict['members'].append(mem_dict['ifname'])
                    show_cli_output(args[0], laglst)
                else:
                     return
    except ApiException as e:
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
    func = eval(sys.argv[1], globals(), sonic_portchannel_client.SonicPortchannelApi.__dict__)

    run(func, sys.argv[2:])
