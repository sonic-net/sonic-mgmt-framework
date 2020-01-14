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


def prompt(msg):
    prompt_msg = msg + " [y/N]:";

    x = raw_input(prompt_msg)
    while x != "y" and  x != "N":
        print ("Invalid input, expected [y/N]")
        x = raw_input(prompt_msg)
    if x == "N":
        exit(1)

def prompt_confirm(func, args):
    if len(args) < 2:
        print("Error:Invalid arguments.")
        exit(1)

    msg = ""
    if (args[0] == "startup-configuration") or ("file://" in args[0]):
        if  len(args) ==3:
            if args[2] == "overwrite":
                msg = ("Clear current config and reload config from " +  args[0] + "? \n"
                "This will result in loss of management connection for a short time.")
            else:
                print("Error:Invalid arguments.")
                exit(1)
        else:
                msg = ("Load config from " + args[0] + "?")
        prompt(msg)

def invoke(func, args):
    body = None

    aa = cc.ApiClient()
    keypath = cc.Path('/restconf/operations/sonic-config-mgmt:copy')
    body = { "sonic-config-mgmt:input":{"source":args[0], "destination":args[1]}}
    if len(args) == 3:
        body["sonic-config-mgmt:input"].update({"overwrite":True})

    return aa.post(keypath, body)

def run(func, args):
    prompt_confirm(func, args)
    try:
        api_response = invoke(func, args)
        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            else:
               status =response["sonic-config-mgmt:output"]
               if status["status"] != 0:
                  print status["status-detail"] 
        else:
            #error response
            print api_response.error_message()

    except:
            # system/network error
            print "%Error: Transaction Failure"

def prompt_confirm_erase(cmds, argv):
    ok = False
    for c in cmds:
        if len(argv) == len(c) - 2:
            if argv == c[2:]:
                prompt(c[0])
                ok = True
    if not ok:
        print("Error: Invalid arguments.")
        exit(1)

def invoke_erase(cmds, argv):
    body = None

    cmd = None
    for c in cmds:
        if len(argv) == len(c) - 2:
            if argv == c[2:]:
                cmd = c[1]
    if not cmd:
        print("Error: Invalid arguments.")
        exit(1)

    aa = cc.ApiClient()
    keypath = cc.Path('/restconf/operations/sonic-config-mgmt:write-erase')
    body = { "sonic-config-mgmt:input":{"subcmd":cmd}}

    return aa.post(keypath, body)

def run_erase(argv):
    cmds = [
        ["Existing switch configuration files except management interface configuration will be removed, continue?", "op_write_erase", "write", "erase"],
        ["Existing switch configuration files will be removed, continue?", "op_write_erase_boot", "write", "erase", "boot"],
        ["All SONiC switch content will be restored to default values, continue?", "op_write_erase_install", "write", "erase", "install"],
        ["Switch configuration erase operation will be cancelled, continue?", "op_no_write_erase", "no", "write", "erase"]
    ]
    prompt_confirm_erase(cmds, argv)
    try:
        api_response = invoke_erase(cmds, argv)
        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            else:
                status = response["sonic-config-mgmt:output"]
                if status["status"] != 0:
                    print status["status-detail"] 
        else:
            #error response
            print api_response.error_message()

    except:
        # system/network error
        print "%Error: Transaction Failure"


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    if sys.argv[1] == "rpc_sonic_config_mgmt_erase":
        run_erase(sys.argv[2:])
    else:
        run(sys.argv[1], sys.argv[2:])

