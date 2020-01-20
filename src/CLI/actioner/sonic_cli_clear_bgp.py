#!/usr/bin/python
###########################################################################
#
# Copyright 2020 Dell, Inc.
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
   asnval = None
   address = ""
   vrfname = ""
   prefixip = ""
   ifname = ""
   pg = ""
   af = ""
   cinall = ""
   clearall = ""
   external = ""
   cin = ""
   cout = ""
   soft = ""
   asn, asnval = args[0].split("=")
   nipv4, nipv4ip = args[1].split("=")
   nipv6, nipv6ip = args[2].split("=")
   i = 6
   for arg in args[6:]:
        if "vrf" == arg:
           vrfname = args[i+1]
        elif "prefix" == arg:
           prefixip = args[i+1]
        elif "interface" == arg:
           ifname = args[i+1]
        elif "peer-group" == arg:
           pg = args[i+1]
        elif "ipv4" == arg:
           af = "IPv4"
        elif "ipv6" == arg:
           af = "IPv6"
        elif "*" == arg:
           if len(args) > 7:
              cinall = True
           else:
              clearall = True
        elif "external" == arg:
           external = True
        elif "in" == arg:
           cin = True
        elif "out" == arg:
           cout = True
        elif "soft" == arg:
           soft = True
        else:
           pass
        i = i + 1
   if nipv4ip != "":
      address = nipv4ip
   if nipv6ip != "":
      address = nipv6ip
   keypath = cc.Path('/restconf/operations/sonic-bgp-clear:clear-bgp')
   if clearall == True:
      body = {"sonic-bgp-clear:input": { "clear-all": clearall}}
   elif asnval != "":
      body = {"sonic-bgp-clear:input": { "vrf-name": vrfname, "family": af, "all": cinall, "address": address, "interface": ifname, "asn": int(asnval), "prefix": prefixip, "peer-group": pg, "external": external, "in": cin, "out": cout, "soft": soft}}
   else:
      body = {"sonic-bgp-clear:input": { "vrf-name": vrfname, "family": af, "all": cinall, "address": address, "interface": ifname, "prefix": prefixip, "peer-group": pg, "external": external, "in": cin, "out": cout, "soft": soft}}
   return api.post(keypath, body)

def run(func, args):
    if func == 'clear_bgp':
        response = clear_bgp_api(args)
        if response.ok():
            if response.content is not None:
                # Get Command Output
                api_response = response.content
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
