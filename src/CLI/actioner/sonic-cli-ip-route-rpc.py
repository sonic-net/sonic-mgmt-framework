from collections import OrderedDict
import netaddr
import sys
import json
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output

def getPrefixAndLen(item):
    ip = netaddr.IPNetwork(item)
    return (ip.value, ip.prefixlen)

def invoke_api(func, args):
    body = None
    vrfname = ""
    prefix = ""
    af = ""
    api = cc.ApiClient()
    i = 0
    for arg in args:
        if "vrf" == arg:
           vrfname = args[i+1]
        elif "ipv4" == arg:
           af = "IPv4"
        elif "ipv6" == arg:
           af = "IPv6"
        elif "prefix" == arg:
           prefix = args[i+1]
        else:
           pass
        i = i + 1

    keypath = cc.Path('/restconf/operations/sonic-ip-show:show-ip-route')
    body = {"sonic-ip-show:input": { "vrf-name": vrfname, "family": af, "prefix": prefix}}
    response = api.post(keypath, body)
    return response

def run(func, args):
    try:
        response = invoke_api(func, args)
        if not response.ok():
            #error response
            print response.error_message()
            exit(1)

        d = response.content['sonic-ip-show:output']['response']
        if len(d) != 0 and "warning" not in d:
           d = json.loads(d)
           routes = d
           keys = sorted(routes,key=getPrefixAndLen)
           temp = OrderedDict()
           for key in keys:
               temp[key] = routes[key]
           d = temp
           show_cli_output(args[0], d)

    except:
        # system/network error
        print "%Error: Transaction Failure"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])
