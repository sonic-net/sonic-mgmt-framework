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
import sonic_vlan_client
from rpipe_utils import pipestr
from sonic_vlan_client.rest import ApiException
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()

vlanDict = {}
plugins = dict()

class ifInfo:
    ifModeDict = {}
    oper_status = "down"

    def __init__(self, ifModeDict):
        self.ifModeDict = ifModeDict
    
    def asdict(self):
        return {'vlanMembers':self.ifModeDict, 'oper_status':self.oper_status}

def register(func):
    plugins[func.__name__] = func
    return func


def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    body = None
    if func.__name__ == 'get_sonic_vlan_sonic_vlan_state':
        keypath = []
    else:
       body = {}

    return keypath,body

def getVlanId(key):
    try:
        return int(key)
    except ValueError:
        return key

def updateVlanToIntfMap(vlanTuple, vlanId):
    for dict in vlanTuple:
        if "ifname" in dict:
            ifName = dict["ifname"]
        if "name" in dict:
            vlanName = dict["name"]
            if vlanId:
                if vlanName != vlanId:
                    continue
            vlanName = vlanName[len("Vlan"):]
        
        if "tagging_mode" in dict:
            ifMode = dict["tagging_mode"]
     
        if not vlanName:
            return

        if vlanDict.get(vlanName) == None:
            ifModeDict = {}

            if ifMode and ifName:
                ifModeDict[ifName] = ifMode
            ifData = ifInfo(ifModeDict)
            vlanDict[vlanName] = ifData
        else:
            ifData = vlanDict.get(vlanName)
            existingifDict = ifData.ifModeDict
            if ifMode and ifName:
                existingifDict[ifName] = ifMode

def updateVlanInfoMap(vlanTuple, vlanId):
    for dict in vlanTuple:
        if "name" in dict:
            vlanName = dict["name"]
            if vlanId:
                if vlanName != vlanId:
                    continue
            vlanName = vlanName[len("Vlan"):]
        if not vlanName:
            return

        operStatus = "down"
        if "oper_status" in dict:
            operStatus = dict["oper_status"]

        if vlanDict.get(vlanName) == None:
            ifModeDict = {}

            ifData = ifInfo(ifModeDict)
            ifData.oper_status = operStatus
            vlanDict[vlanName] = ifData
        else:
            ifData = vlanDict.get(vlanName)
            ifData.oper_status = operStatus

def run(func, args):

    c = sonic_vlan_client.Configuration()
    c.verify_ssl = False
    aa = sonic_vlan_client.SonicVlanApi(api_client=sonic_vlan_client.ApiClient(configuration=c))

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
            if 'sonic-vlan:sonic-vlan-state' in api_response:
                value = api_response['sonic-vlan:sonic-vlan-state']
                if 'VLAN_MEMBER_TABLE' in value:
                    vlanMemberTup = value['VLAN_MEMBER_TABLE']
                    updateVlanToIntfMap(vlanMemberTup, args[0])
                if 'VLAN_TABLE' in value:
                    vlanTup = value['VLAN_TABLE']
                    updateVlanInfoMap(vlanTup, args[0])
            if api_response is None:
                print("Failed")
            else:
                vDict = {}
                for key, val in vlanDict.iteritems():
                    vDict[key] = vlanDict[key].asdict()
                for key, val in vDict.iteritems():
                    sortMembers = collections.OrderedDict(sorted(val['vlanMembers'].items(), key=lambda t: t[1]))
                    val['vlanMembers'] = sortMembers
                vDictSorted = collections.OrderedDict(sorted(vDict.items(), key = lambda t: getVlanId(t[0])))
                if func.__name__ == 'get_sonic_vlan_sonic_vlan_state':
                     show_cli_output(args[1], vDictSorted)
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
    func = eval(sys.argv[1], globals(), sonic_vlan_client.SonicVlanApi.__dict__)

    run(func, sys.argv[2:])
