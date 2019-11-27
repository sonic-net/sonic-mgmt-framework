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

SFLOW_DEFAULT_PORT = 6343

def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func

def call_method(name, args):
    method = plugins[name]
    return method(args)

def get_sflow_admin_state(resp):
    if 'sonic-sflow:sonic-sflow' not in resp:
        return 'down'
    resp = resp['sonic-sflow:sonic-sflow']
    if ('SFLOW' in resp and
        'SFLOW_LIST' in resp['SFLOW'] and
        'admin_state' in resp['SFLOW']['SFLOW_LIST'][0]):
        return resp['SFLOW']['SFLOW_LIST'][0]['admin_state']
    return 'down'

def get_sflow_polling_interval(resp):
    if 'sonic-sflow:sonic-sflow' not in resp:
        return 'default'
    resp = resp['sonic-sflow:sonic-sflow']
    if ('SFLOW' in resp and
        'SFLOW_LIST' in resp['SFLOW'] and
        'polling_interval' in resp['SFLOW']['SFLOW_LIST'][0]):
        return resp['SFLOW']['SFLOW_LIST'][0]['polling_interval']
    return 'default'

def get_sflow_agent_id(resp):
    if 'sonic-sflow:sonic-sflow' not in resp:
        return 'default'
    resp = resp['sonic-sflow:sonic-sflow']
    if ('SFLOW' in resp and
        'SFLOW_LIST' in resp['SFLOW'] and
        'agent_id' in resp['SFLOW']['SFLOW_LIST'][0]):
        return resp['SFLOW']['SFLOW_LIST'][0]['agent_id']
    return 'default'

def get_collector_list(resp):
    if 'sonic-sflow:sonic-sflow' not in resp:
        return []
    resp = resp['sonic-sflow:sonic-sflow']
    if ('SFLOW_COLLECTOR' in resp and
        'SFLOW_COLLECTOR_LIST' in resp['SFLOW_COLLECTOR']):
        return resp['SFLOW_COLLECTOR']['SFLOW_COLLECTOR_LIST']
    return []

def get_session_list(resp, table_name):
    if ('sonic-sflow:sonic-sflow' in resp and
        table_name in resp['sonic-sflow:sonic-sflow'] and
        'SFLOW_SESSION_LIST' in resp['sonic-sflow:sonic-sflow'][table_name]):
        return resp['sonic-sflow:sonic-sflow'][table_name]['SFLOW_SESSION_LIST']
    return []

def get_all_sflow_info(aa):
    keypath = []
    resp = getattr(aa, 'get_sonic_sflow_sonic_sflow')(*keypath)
    resp = aa.api_client.sanitize_for_serialization(resp)
    return resp

def get_all_sflow_sess_info(aa):
    keypath = []
    resp = getattr(aa, 'get_sonic_sflow_sonic_sflow_sflow_session')(*keypath)
    resp = aa.api_client.sanitize_for_serialization(resp)
    return resp

def generate_body(func, args):
    body = None
    keypath = []
    port = SFLOW_DEFAULT_PORT
    if len(args) == 3:
       port = int(args[2])
    if func.__name__ == 'put_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list':
       keypath = [args[0]]
       body = {  "sonic-sflow:SFLOW_COLLECTOR_LIST": [
          {
              "collector_name": args[0],
              "collector_ip": args[1],
              "collector_port": port
          }] }
    elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list':
	keypath = [args[0]]
    elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
    	keypath = [args[0]]
	body = {
	  "sonic-sflow:sample_rate": int(args[1])
	  }
    elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
    	keypath = [args[0]]
    elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_admin_state':
        keypath = [args[0]]
	body = {
	  "sonic-sflow:admin_state": args[1]
	  }
    elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_admin_state':
        keypath = ['global']
        body = {
	  "sonic-sflow:admin_state": args[0]
	  }
    elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id':
        keypath = ['global']
        body = {
	  "sonic-sflow:agent_id": args[0]
	  }
    elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval':
        keypath = ['global']
        body = {
	  "sonic-sflow:polling_interval": int(args[0])
	  }
    elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval':
        keypath = ['global']
    elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id':
        keypath = ['global']

    return keypath, body;

def print_exception(e):
    if e.body != "":
        body = json.loads(e.body)
        if "ietf-restconf:errors" in body:
            err = body["ietf-restconf:errors"]
            if "error" in err:
                errDict = {}
                for dict in err["error"]:
                    for k, v in dict.iteritems():
                        errDict[k] = v
                if "error-message" in errDict:
                    print("Error: " + errDict["error-message"])
                    return
    print("failed")
    return

def getId(item):
    prfx = "Ethernet"
    ifname = item['ifname']
    if ifname.startswith(prfx):
        return int(ifname[len(prfx):])
    return ifname

def run(func, args):
    try:
        c = sonic_sflow_client.Configuration()
        c.verify_ssl = False
        aa = sonic_sflow_client.SonicSflowApi(api_client=sonic_sflow_client.ApiClient(configuration=c))

        resp = get_all_sflow_info(aa)
        cresp = None

        if resp is None:
            print("Can't get sFlow information")
            return

        # sFlow show commands
        if func.__name__ == 'get_sonic_sflow_sonic_sflow':
            sflow_info = {'sflow' : { 'admin_state' : get_sflow_admin_state(resp), 'polling_interval' : get_sflow_polling_interval(resp),
                                      'agent_id' : get_sflow_agent_id(resp)}}
            sflow_col_lst = get_collector_list(resp)
            sflow_info['col_info'] = {}
            sflow_info['col_info']['col_cnt'] = len(sflow_col_lst)
            sflow_info['col_info']['col_lst'] = sflow_col_lst
            show_cli_output(sys.argv[2], sflow_info)
            return
        elif func.__name__ == 'get_sonic_sflow_sonic_sflow_sflow_session_table':
            sess_lst = get_session_list(resp, 'SFLOW_SESSION_TABLE')
            sess_lst = sorted(sess_lst, key=getId)
            show_cli_output(sys.argv[2], sess_lst)
            return

        # sFlow collector config commands
        keypath, body = generate_body(func, args)
        if func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_admin_state' or \
           func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id' or \
           func.__name__ == 'put_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list' or \
           func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval':
            cresp = getattr(aa, func.__name__)(*keypath, body=body)
        elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list':
            name = args[0]
            sflow_col_lst = get_collector_list(resp)
            for col in sflow_col_lst:
                if name in col['collector_name']:
                    cresp = getattr(aa,func.__name__)(*keypath)
        elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval' or \
             func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id':
            cresp = getattr(aa,func.__name__)(*keypath)

        # sFlow session config commands
        elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_admin_state':
            cresp = getattr(aa, func.__name__)(*keypath, body = body)
        elif func.__name__ == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
            cresp = getattr(aa, func.__name__)(*keypath, body = body)
        elif func.__name__ == 'delete_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
            cresp = getattr(aa, func.__name__)(*keypath)
    except ApiException as e:
        print_exception(e)
    return

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), sonic_sflow_client.SonicSflowApi.__dict__)
    run(func, sys.argv[2:])
