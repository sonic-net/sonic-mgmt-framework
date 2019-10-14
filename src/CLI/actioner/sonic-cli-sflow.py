#!/usr/bin/python
import sys
import time
import json
import ast
from collections import OrderedDict
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

def get_sflow():
    pass

def get_sflow_interface():
    pass

def run(func, args):

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
    func = eval(sys.argv[1], globals(), {})
    run(func, sys.argv[2:])
