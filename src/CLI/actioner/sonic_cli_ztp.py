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
    elif func == 'patch_openconfig_ztp_ztp_run':
        path = cc.Path('/restconf/data/openconfig-ztp:ztp/config')
        body = {
                 "openconfig-ztp:config": {
                 "admin_mode":2
                 }
                }
        api_response = aa.patch(path,body)
        return api_response
    else:
        path = cc.Path('/restconf/data/openconfig-ztp:ztp/config')
        if 'no' in args:
            body = {
                     "openconfig-ztp:config": {
                     "admin_mode":0
                     }
                    }
        else:
            body = {
                     "openconfig-ztp:config": {
                     "admin_mode":1
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
                       show_cli_output(args[0],value)
        else:
            if func == "patch_openconfig_ztp_ztp_run":
                response = api_response.content['ietf-restconf:errors']['error'][0]['error-message']
                print("%s" % response)
    except Exception as ex:
        print ex
        print("%Error: Transaction Failure")

if __name__ == '__main__':

    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])
