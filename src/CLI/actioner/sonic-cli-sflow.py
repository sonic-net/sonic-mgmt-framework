#!/usr/bin/python
import sys
import time
import json
import ast
from rpipe_utils import pipestr
from collections import OrderedDict
from scripts.render_cli import show_cli_output
import cli_client as cc


import urllib3
urllib3.disable_warnings()

def get_session_list(resp, table_name):
    if ('sonic-sflow:sonic-sflow' in resp and
        table_name in resp['sonic-sflow:sonic-sflow'] and
        'SFLOW_SESSION_TABLE_LIST' in resp['sonic-sflow:sonic-sflow'][table_name]):
        return resp['sonic-sflow:sonic-sflow'][table_name]['SFLOW_SESSION_TABLE_LIST']
    return []

def get_sflow_info(resp):
    if ('sonic-sflow:sonic-sflow' in resp):
	return resp 
    return {u'sonic-sflow:sonic-sflow': {u'SFLOW': {u'SFLOW_LIST': [{u'sflow_key': u'global', u'admin_state': u'down'}]}}}

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
                    print("% Error: " + errDict["error-message"])
                    return
    else:
        print("% Error: Transaction failure.")
    return

def getId(item):
    prfx = "Ethernet"
    ifname = item['ifname']
    if ifname.startswith(prfx):
        return int(ifname[len(prfx):])
    return ifname

def invoke_api(func, args=[]):
    api = cc.ApiClient()

    if func == 'put_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW_COLLECTOR/SFLOW_COLLECTOR_LIST={collector_name}',
		collector_name=args[0])
        body = {  "sonic-sflow:SFLOW_COLLECTOR_LIST": [
          {
              "collector_name": args[0],
              "collector_ip": args[1]
          }] }
        if len(args) == 3:
            body["sonic-sflow:SFLOW_COLLECTOR_LIST"][0].update({"collector_port":int(args[2])})
	return api.put(path, body)

    elif func == 'delete_sonic_sflow_sonic_sflow_sflow_collector_sflow_collector_list':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW_COLLECTOR/SFLOW_COLLECTOR_LIST={collector_name}',
		collector_name=args[0])
	return api.delete(path)
    elif func == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW_SESSION/SFLOW_SESSION_LIST={ifname}/sample_rate',
		ifname=args[0])
	body = {
                "sonic-sflow:sample_rate": int(args[1])
            }
	return api.patch(path, body)
    elif func == 'delete_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_sample_rate':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW_SESSION/SFLOW_SESSION_LIST={ifname}/sample_rate',
		ifname=args[0])
	return api.delete(path)
    elif func == 'patch_sonic_sflow_sonic_sflow_sflow_session_sflow_session_list_admin_state':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW_SESSION/SFLOW_SESSION_LIST={ifname}/admin_state',
		ifname=args[0])
        keypath = [args[0]]
	body = {
                "sonic-sflow:admin_state": args[1]
            }
	return api.patch(path, body)
    elif func == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_admin_state':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST={sflow_key}/admin_state', sflow_key='global')
	body = { "sonic-sflow:admin_state": args[0] }
	return api.patch(path, body)
    elif func == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST={sflow_key}/agent_id', sflow_key='global')
        body = {
                "sonic-sflow:agent_id": args[0]
            }
	return api.patch(path, body)
    elif func == 'patch_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST={sflow_key}/polling_interval', sflow_key='global')
        body = {
                "sonic-sflow:polling_interval": int(args[0])
            }
	return api.patch(path, body)
    elif func == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_polling_interval':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST={sflow_key}/polling_interval', sflow_key='global')
	return api.delete(path)
    elif func == 'delete_sonic_sflow_sonic_sflow_sflow_sflow_list_agent_id':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow/SFLOW/SFLOW_LIST={sflow_key}/agent_id', sflow_key='global')
	return api.delete(path)
    elif func == 'get_sonic_sflow_sonic_sflow_sflow_session_table':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow')
	return api.get(path)
    elif func == 'get_sonic_sflow_sonic_sflow':
	path = cc.Path('/restconf/data/sonic-sflow:sonic-sflow')
	return api.get(path)

    return api.cli_not_implemented(func)

def run(func, args):
    try:
	response = invoke_api(func, args)

	if response.ok() is False:
            print response.error_message()
            return

	if func == 'get_sonic_sflow_sonic_sflow_sflow_session_table':
            sess_lst = get_session_list(response.content, 'SFLOW_SESSION_TABLE')
            sess_lst = sorted(sess_lst, key=getId)
            show_cli_output(sys.argv[2], sess_lst)

        elif func == 'get_sonic_sflow_sonic_sflow':
            resp = get_sflow_info(response.content)
            show_cli_output(sys.argv[2], resp)

    except Exception as e:
        print("% Error: Internal error: " + str(e))

    return

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])
