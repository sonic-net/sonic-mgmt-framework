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
import urllib3
urllib3.disable_warnings()

def invoke_api(func, args):
    body = None
    api = cc.ApiClient()

    # Set/Get aaa configuration
    if func == 'patch_openconfig_system_augments_system_aaa_authentication_config_failthrough':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/config/failthrough')
       body = { "openconfig-system-augments:failthrough": args[0] }
       return api.patch(path, body)
    elif func == 'patch_openconfig_system_system_aaa_authentication_config_authentication_method':
       path = cc.Path('/restconf/data/openconfig-system:system/aaa/authentication/config/authentication-method')
       body = { "openconfig-system-augments:authentication-method": (args[0]) }
       return api.patch(path, body)
    else:
       body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content

            if api_response is None:
                print("Failed")
            else:
                return
    else:
        print response.error_message()

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])

