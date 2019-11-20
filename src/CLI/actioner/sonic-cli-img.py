#!/usr/bin/python
import sys
import time
import json
import ast
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

def prompt(msg):
    prompt_msg = msg + " [y/N]:";

    x = raw_input(prompt_msg)
    while x != "y" and  x != "N":
        print ("Invalid input, expected [y/N]")
        x = raw_input(prompt_msg)
    if x == "N":
        exit(1)

def prompt_confirm(func, args):
    msg = ""
    if func == 'rpc_sonic_image_management_image_remove':
        if  len(args) > 0:
                msg = ("Remove image " +  args[0] + "?")
        else:
                msg = ("Remove images which are not current and next, continue?")
        prompt(msg)

def validate_imagename(args, command):
    if len(args) < 1:
        if command != "remove":
            print('ERROR: Image name not provided.')
            exit(1)
        else:
            return None
    return {"sonic-image-management:input":{"imagename":args[0]}}
   


def invoke_api(func, args):
    body = None

    api = cc.ApiClient()    

    if func.startswith("rpc") == True:

        if func == 'rpc_sonic_image_management_image_install':
            body= validate_imagename(args, 'install')
            path = cc.Path('/restconf/operations/sonic-image-management:image-install')
            return api.post(path, body)

        if func == 'rpc_sonic_image_management_image_remove':
            body =validate_imagename(args, 'remove')
            path = cc.Path('/restconf/operations/sonic-image-management:image-remove')
            return api.post(path,body)

        if func == 'rpc_sonic_image_management_image_default':
            body =validate_imagename(args, 'default')
            path = cc.Path('/restconf/operations/sonic-image-management:image-default')
            return api.post(path,body)

    if func ==  'get_sonic_image_management_sonic_image_management':
        path = cc.Path('/restconf/data/sonic-image-management:sonic-image-management')
        return api.get(path)

def run(func, args):
  prompt_confirm(func, args)  
  try:
       api_response = invoke_api(func, args)
       if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            else:
                if 'sonic-image-management:output' in response:
                    status =response["sonic-image-management:output"]
                    if status["status"] != 0:
                        print status["status-detail"]
                else:
                    jOut = eval(json.dumps(response))
                    show_cli_output(args[0], jOut)
       else:
            #error response
            print api_response.error_message()

  except:
            # system/network error
            print "%Error: Transaction Failure"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])
