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

    # Set/Get the rules of all INT IFA TS table entries.
    if func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FEATURE_TABLE')
       resp = api.get(path)
       return api.get(path)
    elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE')
       return api.get(path)
    elif func == 'get_list_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
       return api.get(path)
    elif func == 'patch_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table_tam_int_ifa_ts_feature_table_list_enable':
       print "IN pathc"
       print args
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FEATURE_TABLE/TAM_INT_IFA_TS_FEATURE_TABLE_LIST={name}/enable', name=args[0])
       if args[1] == 'True':
           body = { "sonic-tam-int-ifa-ts:enable": True }
       else:
           body = { "sonic-tam-int-ifa-ts:enable": False }
       return api.patch(path, body)

    elif func == 'patch_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
       bodydict = {"name": args[0], "acl-table-name": args[1], "acl-rule-name": args[2]}
       body = { "sonic-tam-int-ifa-ts:TAM_INT_IFA_TS_FLOW_TABLE_LIST": [ bodydict ] }
       return api.patch(path, body)

    elif func == 'delete_list_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
       path = cc.Path('/restconf/data/sonic-tam-int-ifa-ts:sonic-tam-int-ifa-ts/TAM_INT_IFA_TS_FLOW_TABLE/TAM_INT_IFA_TS_FLOW_TABLE_LIST={name}', name=args[0])
       return api.delete(path)

    else:
       body = {}

    return api.cli_not_implemented(func)

def run(func, args):
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
                print("Failed")
            elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_feature_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table':
                show_cli_output(args[0], api_response)
            elif func == 'get_list_sonic_tam_int_ifa_ts_sonic_tam_int_ifa_ts_tam_int_ifa_ts_flow_table_tam_int_ifa_ts_flow_table_list':
                show_cli_output(args[1], api_response)
            else:
                return
        else:
            print response.error_message()

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

