#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom, Inc.
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
import os
import json
import ast
import subprocess
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output
import saghelper as sh

output = {}

def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None
    
    keypath = cc.Path('/restconf/data/sonic-sag:sonic-sag/SAG/SAG_LIST')
    res = api.get(keypath)
    output["sag"] = res.content
    keypath = cc.Path('/restconf/data/sonic-sag:sonic-sag/SAG_GLOBAL/SAG_GLOBAL_LIST')
    res = api.get(keypath)
    output["global"] = res.content
    if func == 'get_ip_sag':
        output["family"] = "IPv4"
    else:
        output["family"] = "IPv6"
    return output

    return api.cli_not_implemented(func)



def run(func, args, renderer):
    response = invoke_api(func, args)

    api_response = response
    if "sonic-sag:SAG_LIST" in api_response["sag"]:
        api_response["miscmap"] = sh.get_if_master_and_oper(api_response["sag"]["sonic-sag:SAG_LIST"])
    #print(api_response)
    show_cli_output(renderer, api_response)

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[3:], sys.argv[2])

