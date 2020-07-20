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
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


blocked_fields = {'parent':0, 'used_power':0, 'allocated_power':0, 'temperature':0}

def filter_json_value(value):
    for key,val in value.items():
        if key in blocked_fields:
            del value[key]
        else:
	    temp = key.split('_')
	    alt_key = ''
	    for i in temp:
		alt_key = alt_key + i.capitalize() + ' '
	    value[alt_key]=value.pop(key)

    return value

def get_openconfig_platform_components(*args):
    path = cc.Path('/restconf/data/openconfig-platform:components')
    return cc.ApiClient().get(path)

def run(func, args):
    # lookup and invoke the function name passed by CLI
    response = globals()[func](*args)

    if response.ok():
        api_response = response.content
        if api_response is None:
            return
        else:
            value =  api_response['openconfig-platform:components']['component'][0]['state']
	    if value is None:
                return
	    if 'oper-status' in value:
		temp = value['oper-status'].split(':')
		if temp[len(temp) - 1] is not None:
	            value['oper-status'] = temp[len(temp) - 1]
            show_cli_output(sys.argv[2],filter_json_value(value))

    else:
        print(response.error_message())
        return 1

if __name__ == '__main__':

    pipestr().write(sys.argv)
    #pdb.set_trace()
    func = sys.argv[1]
    run(func, sys.argv[2:])

