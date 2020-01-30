#!/usr/bin/python

import sys
import os
import json
import collections
import re
import cli_client as cc
from scripts.render_cli import show_cli_output

## Run an external command
def run_config_cmd(data):
    aa = cc.ApiClient()
    keypath = cc.Path('/restconf/operations/sonic-kdump:kdump-config')
    body = { "sonic-kdump:input":data}

    api_response = aa.post(keypath, body)
    if api_response.ok():
        response = api_response.content
        if response is not None and 'sonic-kdump:output' in response:
            print(response['sonic-kdump:output']['result'])
    else:
        print(api_response.error_message())

## Run an external command
def run_show_cmd(data):
    aa = cc.ApiClient()
    keypath = cc.Path('/restconf/operations/sonic-kdump:kdump-state')
    body = { "sonic-kdump:input":data}

    api_response = aa.post(keypath, body)
    if api_response.ok():
        response = api_response.content
        if response is not None and 'sonic-kdump:output' in response:
            print(response['sonic-kdump:output']['result'])
    else:
        print(api_response.error_message())

## Run a kdump 'show' command
def kdump_show_cmd(cmd):
    run_show_cmd({})

## Display kdump status
def cmd_show_status():
    run_show_cmd({"Param":"status"})

## Display kdump memory
def cmd_show_memory():
    run_show_cmd({"Param":"memory"})

## Display kdump num_dumps
def cmd_show_num_dumps():
    run_show_cmd({"Param":"num_dumps"})

## Display kdump files
def cmd_show_files():
    run_show_cmd({"Param":"files"})

## Display kdump log
def cmd_show_log(record, lines=None):
    if lines is None:
        run_show_cmd({"Param":"log %s" % record})
    else:
        run_show_cmd({"Param":"log %s %s" % (record, lines)})

## Enable kdump
def cmd_enable():
    run_config_cmd({"Enabled":True, "Num_Dumps":0, "Memory":""})

## Disable kdump
def cmd_disable():
    run_config_cmd({"Enabled":False, "Num_Dumps":0, "Memory":""})

## Set memory allocated for kdump
def cmd_set_memory(memory):
    run_config_cmd({"Enabled":False, "Num_Dumps":0, "Memory":memory})

## Set max numbers of kernel core files
def cmd_set_num_dumps(num_dumps):
    run_config_cmd({"Enabled":False, "Num_Dumps":num_dumps, "Memory":""})

## Main function
def run(func, args):

    argv = []
    argv.append("sonic_cli_kdump")
    argv.append(func)
    for x in args:
        argv.append(x)

    if len(argv) == 3:
        if argv[1] == 'show' and argv[2] == 'kdump':
            cmd_show_status()
        elif argv[1] == 'kdump' and argv[2] == 'enable':
            cmd_enable()
        elif argv[1] == 'no' and argv[2] == 'kdump':
            cmd_disable()
    elif len(argv) == 4:
        if argv[1] == 'show' and argv[2] == 'kdump' and argv[3] == 'status':
            cmd_show_status()
        elif argv[1] == 'show' and argv[2] == 'kdump' and argv[3] == 'memory':
            cmd_show_memory()
        elif argv[1] == 'show' and argv[2] == 'kdump' and argv[3] == 'num_dumps':
            cmd_show_num_dumps()
        elif argv[1] == 'show' and argv[2] == 'kdump' and argv[3] == 'files':
            cmd_show_files()
        elif argv[1] == 'show' and argv[2] == 'kdump' and argv[3] == 'log':
            if len(argv[4:]) == 1:
                cmd_show_log(argv[4])
            else:
                cmd_show_log(argv[4], argv[5])
        elif argv[1] == 'kdump' and argv[2] == 'memory':
            cmd_set_memory(argv[3])
        elif argv[1] == 'kdump' and argv[2] == 'num_dumps':
            cmd_set_num_dumps(int(argv[3]))
        elif argv[1] == 'no' and argv[2] == 'kdump' and argv[3] == 'memory':
            cmd_set_memory("0M-2G:256M,2G-4G:320M,4G-8G:384M,8G-:448M")
        elif argv[1] == 'no' and argv[2] == 'kdump' and argv[3] == 'num_dumps':
            cmd_set_num_dumps(int(3))
    elif len(argv) == 5 and argv[1] == 'show':
        if argv[2] == 'kdump' and argv[3] == 'log':
            cmd_show_log(argv[4])
    elif len(argv) == 6 and argv[1] == 'show':
        if argv[2] == 'kdump' and argv[3] == 'log':
            cmd_show_log(argv[4], argv[5])
    elif len(argv) == 5 and argv[1] == 'do':
        if argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'status':
            cmd_show_status()
        elif argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'memory':
            cmd_show_memory()
        elif argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'num_dumps':
            cmd_show_num_dumps()
        elif argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'files':
            cmd_show_files()
        elif argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'log':
            if len(argv[5:]) == 1:
                cmd_show_log(argv[5])
            else:
                cmd_show_log(argv[5], argv[6])
    elif len(argv) == 6 and argv[1] == 'do':
        if argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'log':
            cmd_show_log(argv[5])
    elif len(argv) == 7 and argv[1] == 'do':
        if argv[2] == 'show' and argv[3] == 'kdump' and argv[4] == 'log':
            cmd_show_log(argv[5], argv[6])

if __name__ == '__main__':
    run(sys.argv[1], sys.argv[2:])
