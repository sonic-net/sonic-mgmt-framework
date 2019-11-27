#!/usr/bin/python
import sys
import time
import json
import ast
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
import cli_client as cc
import re
import urllib3
urllib3.disable_warnings()

def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    if func == 'get_openconfig_ztp_ztp_state':
        path = cc.Path('/restconf/data/openconfig-ztp:ztp/state')
        return aa.get(path)
    else:
        path = cc.Path('/restconf/data/openconfig-ztp:ztp/config')
        if 'no' in sys.argv:
            body = {
                     "openconfig-ztp:config": {
                     "admin_mode":False
                     }
                    }
        else:
            body = {
                     "openconfig-ztp:config": {
                     "admin_mode":True
                     }
                    }
        return aa.patch(path,body)



def run(func, args):
    try:
        api_response = invoke(func, args)
        if api_response.ok():
            if api_response.content is not None:
                response = api_response.content
                if 'openconfig-ztp:state' in response.keys():
                   value = response['openconfig-ztp:state']
                   if value is not None:
                       show_cli_output(sys.argv[2],value)
    except:
        print("%Error: Transaction Failure")

if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])
