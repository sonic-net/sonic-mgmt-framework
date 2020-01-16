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
import netaddr
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output
from bgp_openconfig_to_restconf_map import restconf_map

def clear_bgp_api(args):
   api = cc.ApiClient()
   keypath = []
   body = None
   print len(args)
   print args
   asn, asnval = args[0].split("=")
   nipv4, nipv4ip = args[1].split("=")
   nipv6, nipv6ip = args[2].split("=")
   cmd = "{\"sonic-bgp-clear:input\": { "
   i = 6
   for arg in args[6:]:
        if "vrf" == arg:
           cmd = cmd + "\"vrf\": " + args[i+1] + ", "
        elif "prefix" == arg:
           cmd = cmd + "\"prefix\": " + args[i+1] + ", "
        elif "interface" == arg:
           cmd = cmd + "\"interface\": " + args[i+1] + ", "
        elif "peer-group" == arg:
           cmd = cmd + "\"peer-group\": " + args[i+1] + ", "
        elif "ipv4" == arg:
           cmd = cmd + "\"family\": \"IPv4\", "
        elif "ipv6" == arg:
           cmd = cmd + "\"family\": \"IPv6\", "
        elif "*" == arg:
           if len(args) > 7:
              cmd = cmd + "\"all\": true, "
           else:
              cmd = cmd + "\"clear-all\": true, "
        elif "external" == arg:
           cmd = cmd + "\"external\": true, "
        elif "in" == arg:
           cmd = cmd + "\"in\": true, "
        elif "out" == arg:
           cmd = cmd + "\"out\": true, "
        elif "soft" == arg:
           cmd = cmd + "\"soft\": true, "
        else:
           pass
        i = i + 1

   if asnval:
      cmd = cmd + "\"asn\": " + asnval + ", "
   elif nipv4ip:
      cmd = cmd + "\"address\": " + nipv4ip + ", "
   elif nipv6ip:
      cmd = cmd + "\"address\": " + nipv6ip + ", "
   else:
      pass
   cmd = cmd[:-2]
   cmd = cmd + "}}"
   print "CMD"
   print cmd
   keypath = cc.Path('/restconf/operations/sonic-bgp-clear:clear-bgp')
   print "KEYPATH"
   print keypath
   body = { "sonic-bgp-clear:input":{"clear-all": true} }
   print "BODY"
   print body
   return aa.post(keypath,body)

def run(func, args):
    if func == 'clear_bgp':
        response = clear_bgp_api(args)
        if response.ok():
            if response.content is not None:
                # Get Command Output
                api_response = response.content
                print(api_response)
                if api_response is None:
                    print("Failed")
                    sys.exit(1)
        else:
            print response.error_message()
            sys.exit(1)
    else:
       return

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])
