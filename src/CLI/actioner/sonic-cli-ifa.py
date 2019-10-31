import sys
import time
import json
import ast
import sonic_ifa_client
from sonic_ifa_client.rest import ApiException
from scripts.render_cli import show_cli_output

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

def generate_body(func, args):
    body = None
    keypath = []

    # Set/Get the rules of all IFA table entries.
    if func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_device_table':
       keypath = []
    elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table':
       keypath = []
    elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_collector_table':
       keypath = []
    elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_collector_table_tam_collector_table_list':
       keypath = [args[0]]
    elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table':
       keypath = []
    elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       keypath = [args[0]]
    elif func.__name__ == 'patch_sonic_ifa_sonic_ifa_tam_device_table_tam_device_table_list_deviceid':
       keypath = [ args[0] ]
       body = { "sonic-ifa:deviceid": int(args[1]) }
    elif func.__name__ == 'delete_sonic_ifa_sonic_ifa_tam_device_table_tam_device_table_list_deviceid':
       keypath = [args[0]]
    elif func.__name__ == 'patch_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table_tam_int_ifa_feature_table_list_enable':
       keypath = [ args[0] ]
       if args[1] == 'True':
           body = { "sonic-ifa:enable": True }
       else:
           body = { "sonic-ifa:enable": False }
    elif func.__name__ == 'patch_list_sonic_ifa_sonic_ifa_tam_collector_table_tam_collector_table_list':
       keypath = [ ]
       body = {
           "sonic-ifa:TAM_COLLECTOR_TABLE_LIST": [
              {
                  "name": args[0], "ipaddress-type": args[1], "ipaddress": args[2], "port": int(args[3])
              }
           ]
       }
    elif func.__name__ == 'delete_sonic_ifa_sonic_ifa_tam_collector_table_tam_collector_table_list':
       keypath = [ args[0] ]
    elif func.__name__ == 'patch_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       keypath = [ args[0] ]
       body = {
                  "sonic-ifa:TAM_INT_IFA_FLOW_TABLE_LIST": [
                       {
                           "name": args[0], "acl-rule-name": args[1], "acl-table-name": args[2], "sampling-rate": int(args[3]), "collector-name":args[4]
                       }
                  ]
              }
    elif func.__name__ == 'delete_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
       keypath = [ args[0] ]
    else:
       body = {}

    return keypath,body

def run(func, args):

    c = sonic_ifa_client.Configuration()
    c.verify_ssl = False
    aa = sonic_ifa_client.SonicIfaApi(api_client=sonic_ifa_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)

    try:
        if body is not None:
           api_response = getattr(aa,func.__name__)(*keypath, body=body)
        else :
           api_response = getattr(aa,func.__name__)(*keypath)

        if api_response is None:
            print ("Success")
        else:
            # Get Command Output
            api_response = aa.api_client.sanitize_for_serialization(api_response)
            if 'sonic-ifa:sonic-ifa' in api_response:
                value = api_response['sonic-ifa:sonic-ifa']
                if 'TAM_DEVICE_TABLE' in value:
                    tup = value['TAM_DEVICE_TABLE']
                elif 'TAM_INT_IFA_FEATURE_TABLE' in value:
                    tup = value['TAM_INT_IFA_FEATURE_TABLE']
                elif 'TAM_COLLECTOR_TABLE' in value:
                    tup = value['TAM_COLLECTOR_TABLE']
                elif 'TAM_INT_IFA_FLOW_TABLE' in value:
                    tup = value['TAM_INT_IFA_FLOW_TABLE']
                else:
                    api_response = None

            if api_response is None:
                print("Failed")
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_device_table':
                show_cli_output(args[0], api_response)
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_feature_table':
                show_cli_output(args[0], api_response)
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_collector_table':
                show_cli_output(args[0], api_response)
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_collector_table_tam_collector_table_list':
                show_cli_output(args[1], api_response)
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table':
                show_cli_output(args[0], api_response)
            elif func.__name__ == 'get_sonic_ifa_sonic_ifa_tam_int_ifa_flow_table_tam_int_ifa_flow_table_list':
                show_cli_output(args[1], api_response)
            else:
                return
    except ApiException as e:
        print("Exception when calling get_sonic_ifa ->%s : %s\n" %(func.__name__, e))
        if e.body != "":
            body = json.loads(e.body)
            if "ietf-restconf:errors" in body:
                 err = body["ietf-restconf:errors"]
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
            print "%Error: Transaction Failure"
        else:
            print "Failed"

if __name__ == '__main__':

    func = eval(sys.argv[1], globals(), sonic_ifa_client.SonicIfaApi.__dict__)

    run(func, sys.argv[2:])

