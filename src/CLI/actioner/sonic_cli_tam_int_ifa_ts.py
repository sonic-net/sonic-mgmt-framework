import sys
import json
import collections
import re
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector
import urllib3
urllib3.disable_warnings()

def invoke_api(func, args):
    body = None
    api = cc.ApiClient()

    # Set/Get the rules of all INT IFA TS table entries.
    if func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FEATURE_TABLE')
       resp = api.get(path)
       return api.get(path)
    elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table':
        if (len(args) == 2) and (args[1] != "all"):
           path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[1])
           return api.get(path)
        else:
           path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE')
           return api.get(path)

    elif func == 'patch_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table_tam_int_ifa_ts_feature_table_list_enable':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FEATURE_TABLE/TAM_INT_IFA_TS_FEATURE_TABLE_LIST={name}/enable', name='feature')
       if args[0] == 'enable':
           body = { "sonic-tam-int-ifa-ts:enable": True }
       else:
           body = { "sonic-tam-int-ifa-ts:enable": False }
       return api.patch(path, body)

    elif func == 'patch_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
       bodydict = {"name": args[0], "acl-table-name": args[1], "acl-rule-name": args[2]}
       body = { "sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE_LIST": [ bodydict ] }
       return api.patch(path, body)

    elif func == 'delete_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
        if (len(args) == 1) and (args[0] == "all"):
           path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST')
           return api.delete(path)
        elif (len(args) == 1):
           path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
           return api.delete(path)


    else:
       body = {}

    return api.cli_not_implemented(func)

def get_tam_ifa_ts_status(args):
    api_response = {}
    api = cc.ApiClient()

    path = cc.Path('/restconf/data/sonic-tam:sonic-tam/TAM_DEVICE_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['device'] = response.content['sonic-tam:TAM_DEVICE_TABLE']['TAM_DEVICE_TABLE_LIST']

    path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FEATURE_TABLE')
    response = api.get(path)                                                                               
    if response.ok():                                                                                      
        if response.content:                                                                               
            api_response['feature'] = response.content['sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FEATURE_TABLE']['TAM_INT_IFA_TS_FEATURE_TABLE_LIST']


    path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE')
    response = api.get(path)
    if response.ok():
        if response.content:
            api_response['flow'] = response.content['sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE']['TAM_INT_IFA_TS_FLOW_TABLE_LIST']

    show_cli_output("show_tam_ifa_ts_status.j2", api_response)


def get_tam_int_ifa_ts_supported(args):
    api_response = {}

    # connect to APPL_DB
    app_db = ConfigDBConnector()
    app_db.db_connect('APPL_DB')

    key = 'SWITCH_TABLE:switch'
    data = app_db.get(app_db.APPL_DB, key, 'tam_int_ifa_ts_supported')

    if data and data == 'True':
        api_response['feature'] = data
    else:
        api_response['feature'] = 'False'

    show_cli_output("show_tam_ifa_ts_feature_supported.j2", api_response)

def get_tam_int_ifa_ts_flow_stats(args):
    api_response = {}
    api = cc.ApiClient()

    # connect to COUNTERS_DB
    counters_db = ConfigDBConnector()
    counters_db.db_connect('COUNTERS_DB')

    if len(args) == 1 and args[0] != "all":
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
    else:
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE')

    response = api.get(path)

    if response.ok():
        if response.content:
            if len(args) == 1 and args[0] != "all":
                api_response = response.content['sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE_LIST']
            else:
                api_response = response.content['sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE']['TAM_INT_IFA_TS_FLOW_TABLE_LIST']

            for i in range(len(api_response)):
                api_response[i]['Packets'] = 0
                api_response[i]['Bytes'] = 0
                if "acl-table-name" not in api_response[i] and "acl-rule-name" not in api_response[i]:
                  return

                acl_counter_key = 'COUNTERS:' + api_response[i]['acl-table-name'] + ':' + api_response[i]['acl-rule-name']
                flow_stats = counters_db.get_all(counters_db.COUNTERS_DB, acl_counter_key)
                if flow_stats is not None:
			api_response[i]['Packets'] = flow_stats['Packets']
			api_response[i]['Bytes'] = flow_stats['Bytes']

    show_cli_output("show_tam_int_ifa_ts_flow_stats.j2", api_response)


def run(func, args):
    if func == 'get_tam_ifa_ts_status':
        get_tam_ifa_ts_status(args)
        return
    elif func == 'get_tam_int_ifa_ts_supported':
        get_tam_int_ifa_ts_supported(args)
        return
    elif func == 'get_tam_int_ifa_ts_flow_stats':
	get_tam_int_ifa_ts_flow_stats(args)
        return

    response = invoke_api(func, args)
    if response.ok():
        if response.content is not None:
            # Get Command Output
            api_response = response.content
            if 'sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE' in api_response:
                value = api_response['sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE'] 
                if 'TAM_INT_IFA_TS_FLOW_TABLE_LIST' in value:
                    tup = value['TAM_INT_IFA_TS_FLOW_TABLE_LIST']
                else:
                    api_response = None

            if api_response is None:
                print "Failed"
            elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_list_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
                show_cli_output(args[1], api_response)
            else:
                return
    else:
            api_response = response.content
            if "ietf-restconf:errors" in api_response:
                 err = api_response["ietf-restconf:errors"]
                 if "error" in err:
                     errList = err["error"]

                     errDict = {}
                     for dict in errList:
                         for k, v in dict.iteritems():
                              errDict[k] = v

                     if "error-message" in errDict:
                         print "%Error: " + errDict["error-message"]
                         return
                     print "%Error: Transaction Failure"
                     return
            print response.error_message()
            print "%Error: Transaction Failure"

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]

    if func == 'get_tam_ifa_ts_status':
        get_tam_ifa_ts_status(sys.argv[2:])
    elif func == 'get_tam_int_ifa_ts_supported':
        get_tam_int_ifa_ts_supported(sys.argv[2:])
    elif func == 'get_tam_int_ifa_ts_flow_stats':
	get_tam_int_ifa_ts_flow_stats(sys.argv[2:])
    else: 
        run(func, sys.argv[2:])

