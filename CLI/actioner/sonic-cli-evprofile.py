from collections import OrderedDict
import sys
import json
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
import cli_client as cc

def run(func, args):
    aa = cc.ApiClient()

    if func == 'get_evprofile':
        keypath = cc.Path('/restconf/operations/sonic-evprofile:get-evprofile')
        response = aa.post(keypath)
        if response.ok():
            api_response = response.content["sonic-get-evprofile:output"]
            show_cli_output("show_evprofile_rpc.j2", api_response)
        else:
            print(response.error_message())
            return 1

    if func == 'set_evprofile':
        keypath = cc.Path('/restconf/operations/sonic-evprofile:set-evprofile')
        body = { "sonic-evprofile:input": { "file-name": args[0]} }
        response = aa.post(keypath, body)
        if not response.ok():
            print(response.error_message())
            return 1

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

