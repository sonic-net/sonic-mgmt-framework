#!/usr/bin/python
import sys
import time
import json
import ast
import re
from scripts.render_cli import show_cli_output
from rpipe_utils import pipestr
import cli_client as cc
import urllib3

api = cc.ApiClient()
urllib3.disable_warnings()
plugins = dict()


def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func

def call_method(name, args):
    method = plugins[name]
    return method(args)

def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    # Gather tech support information into a compressed file
    if func == 'rpc_sonic_show_techsupport_sonic_show_techsupport_info':
        keypath = cc.Path('/restconf/operations/sonic-show-techsupport:sonic-show-techsupport-info')
        body = {}
        if len(args) < 2:
            body = {"sonic-show-techsupport:input":{"date": None}}
        else:
            body = {"sonic-show-techsupport:input":{"date": args[1]}}
        return aa.post(keypath, body)
    else:
        print("%Error: not implemented")
        exit(1)


def run(func, args):

    try:
        api_response = invoke(func, args)

        if api_response.ok():
            response = api_response.content

	    if response is None:
                print ("ERROR: No output file generated: \n"
                       "Invalid input: Incorrect DateTime format")

	    else:
		if func == 'rpc_sonic_show_techsupport_sonic_show_techsupport_info':
                    if not response['sonic-show-techsupport:output']:
                        print("ERROR: No top level output object")
                        show_cli_output(sys.argv[2], "")
                        return
                    elif response['sonic-show-techsupport:output'] is None:
                        print("ERROR: No second level output object")
		        show_cli_output(sys.argv[2], "")
		        return
                    output_file = response['sonic-show-techsupport:output']
		    show_cli_output(sys.argv[2], output_file)
		else:
                    print("ERROR: Python: Show Techsupport parsing Failed: Invalid function")
        else:
            #error response
            print api_response.error_message()

    except:
        print("Exception calling SonicShowTechsupportApi->%s\n" %(func))

if __name__ == '__main__':
    pipestr().write(sys.argv)
    run(sys.argv[1], sys.argv[2:])
