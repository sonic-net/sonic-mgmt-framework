#!/usr/bin/python
import sys
import time
import json
import ast
from rpipe_utils import pipestr
from collections import OrderedDict
from scripts.render_cli import show_cli_output
import sonic_sflow_client
from sonic_sflow_client.api.sonic_sflow_api import SonicSflowApi  
from sonic_sflow_client.rest import ApiException


import urllib3
urllib3.disable_warnings()
plugins = dict()


def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func

def call_method(name, args):
    method = plugins[name]
    return method(args)

def get_sflow():
    pass

def get_sflow_interface():
    pass

def run(func, args):
    c = sonic_sflow_client.Configuration()
    c.verify_ssl = False
    aa = sonic_sflow_client.SonicSflowApi(api_client=sonic_sflow_client.ApiClient(configuration=c))
    keypath = []
    api_response = getattr(aa,'get_sonic_sflow_sonic_sflow')(*keypath)
    print(api_response)
    api_response = getattr(aa, 'get_sonic_sflow_sonic_sflow_sflow_session')(*keypath)
    print(api_response)
    keypath = ['col1']
    body = {  "sonic-sflow:SFLOW_COLLECTOR_LIST": [
               {
                   "collector_name": 'col1',
                   "collector_ip": "5.5.5.5",
                   "collector_port": 54321
               }] }

    api_response = getattr(aa, 'patch_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list')(*keypath, body=body)
    print(api_response)
    keypath = []
    api_response = getattr(aa,'get_sonic_sflow_sonic_sflow')(*keypath)
    print(api_response)
    # create a body block
    if (func.__name__ == 'get_sflow'):
        sflow_info = {'sflow' : {'admin_state' : 'enabled', 'polling-interval' : 20, 'agent-id' : 'default'}}
    else:
	sflow_info = {}
        sflow_info['sflow'] = OrderedDict()
        for i in range(30):
            sflow_info['sflow']['Ethernet'+str(i)] = {'admin_state' : 'enabled', 'sampling_rate' : 4000}
    show_cli_output(sys.argv[2], sflow_info)
    return

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), sonic_sflow_client.SonicSflowApi.__dict__)
    run(func, sys.argv[2:])
