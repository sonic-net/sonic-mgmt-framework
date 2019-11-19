#!/usr/bin/python

import sys
import cli_client as cc

def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    # Gather tech support information into a compressed file
    if func == 'rpc_sonic_show_techsupport_sonic_show_techsupport_info':
        keypath = cc.Path('/restconf/operations/sonic-show-techsupport:sonic-show-techsupport-info')
        if args is None:
            body = {"sonic-show-techsupport:input":{"date": None}}
        else:
            body = {"sonic-show-techsupport:input":{"date": args[0]}}
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
                print ("ERROR: No output file generated")

	    else:
		if func == 'rpc_sonic_show_techsupport_sonic_show_techsupport_info':
                    if not response['sonic-show-techsupport:output']:
                        print("ERROR: no output file available")
                        return
                    elif response['sonic-show-techsupport:output'] is None:
                        print("ERROR: No output file available")
		        return
                    output_file_object = response['sonic-show-techsupport:output']
                    if output_file_object['output-filename'] is None:
                        print("ERROR: No output file available")
                        return
                    output_filename = output_file_object['output-filename']
                    if len(output_filename) is 0:
                        print("Invalid input: Incorrect DateTime format")
                    else:
                        print("Output stored in:  " + output_filename)
		else:
                    print("ERROR: Python: Show Techsupport parsing Failed: Invalid function")
        else:
            #error response
            print api_response.error_message()

    except:
        print("An exception occurred while attempting to gather the requested "
              "information via a remote procedure call. The failing RPC function "
              " is: %s\n" %(func))

if __name__ == '__main__':
    if len(sys.argv) == 3:
	    run(sys.argv[1], sys.argv[2:])
    else:
            run(sys.argv[1], None)
