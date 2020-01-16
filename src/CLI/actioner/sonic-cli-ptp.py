import sys
import base64
import struct
import socket
import os.path
from os import path
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


def node_addr_type(address):
    try:
        socket.inet_pton(socket.AF_INET, address)
    except socket.error:
        try:
            socket.inet_pton(socket.AF_INET6, address)
        except socket.error:
            return "mac"

        return "ipv6"

    return "ipv4"


def decode_base64(string):
    my_bin = base64.b64decode(string)
    return "%02x%02x%02x.%02x%02x.%02x%02x%02x" % struct.unpack("BBBBBBBB", my_bin)


def get_unicast_table(aa, instance_num, port_num):
    tmp_keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}',
                          instance_number=instance_num)
    tmp_response = aa.get(tmp_keypath)
    if tmp_response is None:
        return 0, "None"

    found = 0
    if tmp_response.ok():
        response = tmp_response.content

        if response is not None and response != {}:
            for i in response['ietf-ptp:instance-list']:
                if 'port-ds-list' in i:
                    for j in i['port-ds-list']:
                        if 'port-number' in j:
                            if j['port-number'] == int(port_num):
                                found = 1
                                if 'ietf-ptp-ext:unicast-table' in j:
                                    if j['ietf-ptp-ext:unicast-table'] == '':
                                        return found, "None"
                                    else:
                                        return found, j['ietf-ptp-ext:unicast-table']

    return found, "None"


def get_port_num(interface):
    if 'Ethernet' in interface:
        port_num = interface.replace("Ethernet", "")
    if 'Vlan' in interface:
        port_num = str(int(interface.replace("Vlan", "")) + 1000)
    return port_num


def invoke(func, args):
    rc = None
    body = None
    aa = cc.ApiClient()

    if func == 'patch_ietf_ptp_ptp_instance_list_default_ds_domain_number':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/domain-number',
                          instance_number=args[0])
        body = {"ietf-ptp:domain-number": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_default_ds_two_step_flag':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/two-step-flag',
                          instance_number=args[0])
        if args[1] == "enable":
            body = {"ietf-ptp:two-step-flag": True}
        else:
            body = {"ietf-ptp:two-step-flag": False}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_default_ds_priority1':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/priority1',
                          instance_number=args[0])
        body = {"ietf-ptp:priority1": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_default_ds_priority2':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/priority2',
                          instance_number=args[0])
        body = {"ietf-ptp:priority2": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_port_ds_list_log_announce_interval':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:log-announce-interval',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:log-announce-interval": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_port_ds_list_announce_receipt_timeout':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:announce-receipt-timeout',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:announce-receipt-timeout": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_port_ds_list_log_sync_interval':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:log-sync-interval',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:log-sync-interval": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'patch_ietf_ptp_ptp_instance_list_port_ds_list_log_min_delay_req_interval':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:log-min-delay-req-interval',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:log-min-delay-req-interval": int(args[1])}
        rc = aa.patch(keypath, body)
    elif func == 'clock-type':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:clock-type',
                          instance_number=args[0])

        body = {"ietf-ptp-ext:clock-type": args[1]}
        rc = aa.patch(keypath, body)
    elif func == 'network-transport':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:network-transport',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:network-transport": args[1]}
        rc = aa.patch(keypath, body)
    elif func == 'unicast-multicast':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:unicast-multicast',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:unicast-multicast": args[1]}
        rc = aa.patch(keypath, body)
    elif func == 'domain-profile':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:domain-profile',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:domain-profile": args[1]}
        rc = aa.patch(keypath, body)
    elif func == 'udp6-scope':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds/ietf-ptp-ext:udp6-scope',
                          instance_number=args[0])
        body = {"ietf-ptp-ext:udp6-scope": int(args[1], 0)}
        rc = aa.patch(keypath, body)
    elif func == 'add_master_table':
        port_num = get_port_num(args[1])
        found, uc_tbl = get_unicast_table(aa, args[0], port_num)

        if not found:
            print("%Error: " + args[1] + " has not been added")
            sys.exit()

        nd_list = []
        if uc_tbl != 'None':
            nd_list = uc_tbl.split(',')
        if args[2] in nd_list:
            # entry already exists
            sys.exit()
        if len(nd_list) == 1 and nd_list[0] == '':
            nd_list = []
        if len(nd_list) > 0 and node_addr_type(nd_list[0]) != node_addr_type(args[2]):
            print("%Error: Mixed address types not allowed")
            sys.exit()
        if len(nd_list) >= 8:
            print("%Error: maximum 8 nodes")
            sys.exit()
        nd_list.append(args[2])
        value = ','.join(nd_list)
        args[2] = value
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/port-ds-list={port_number}/ietf-ptp-ext:unicast-table',
                          instance_number=args[0], port_number=port_num)
        body = {"ietf-ptp-ext:unicast-table": args[2]}
        rc = aa.patch(keypath, body)
    elif func == 'del_master_table':
        port_num = get_port_num(args[1])
        found, uc_tbl = get_unicast_table(aa, args[0], port_num)

        nd_list = []
        if uc_tbl != 'None':
            nd_list = uc_tbl.split(',')
        if args[2] not in nd_list:
            # entry doesn't exists
            sys.exit()

        nd_list.remove(args[2])
        value = ','.join(nd_list)
        args[2] = value
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/port-ds-list={port_number}/ietf-ptp-ext:unicast-table',
                          instance_number=args[0], port_number=port_num)
        body = {"ietf-ptp-ext:unicast-table": args[2]}
        rc = aa.patch(keypath, body)
    elif func == 'post_ietf_ptp_ptp_instance_list_port_ds_list_port_state':
        port_num = get_port_num(args[1])
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/port-ds-list={port_number}/underlying-interface',
                          instance_number=args[0], port_number=port_num)
        body = {"ietf-ptp:underlying-interface": args[1]}
        rc = aa.patch(keypath, body)
    elif func == 'delete_ietf_ptp_ptp_instance_list_port_ds_list':
        port_num = get_port_num(args[1])
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/port-ds-list={port_number}',
                          instance_number=args[0], port_number=port_num)
        rc = aa.delete(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list_time_properties_ds':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/time-properties-ds',
                          instance_number=args[0])
        rc = aa.get(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list_parent_ds':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/parent-ds',
                          instance_number=args[0])
        rc = aa.get(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list_port_ds_list':
        port_num = get_port_num(args[1])
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/port-ds-list={port_number}',
                          instance_number=args[0], port_number=port_num)
        rc = aa.get(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list_default_ds':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/default-ds',
                          instance_number=args[0])
        rc = aa.get(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}',
                          instance_number=args[0])
        rc = aa.get(keypath)
    elif func == 'get_ietf_ptp_ptp_instance_list_current_ds':
        keypath = cc.Path('/restconf/data/ietf-ptp:ptp/instance-list={instance_number}/current-ds',
                          instance_number=args[0])
        rc = aa.get(keypath)
    else:
        print("%Error: not implemented")
        exit(1)

    return rc


def run(func, args):
    if not path.exists('/proc/bcm/ksync/stats'):
        print("%Error: PTP feature not supported")
        sys.exit(-1)

    api_response = invoke(func, args)
    if api_response is None:
        return

    if api_response.ok():
        response = api_response.content
        if response is None:
            if func != 'network-transport':
                print("Success")
        else:
            # Get Command Output
            if func == 'get_ietf_ptp_ptp_instance_list_default_ds':
                if not response == {}:
                    if 'clock-identity' in response['ietf-ptp:default-ds']:
                        response['ietf-ptp:default-ds']['clock-identity'] = decode_base64(response['ietf-ptp:default-ds']['clock-identity'])
                show_cli_output(args[1], response)
            elif func == 'get_ietf_ptp_ptp_instance_list_port_ds_list':
                show_cli_output(args[2], response)
            elif func == 'get_ietf_ptp_ptp_instance_list_parent_ds':
                if not response == {}:
                    if 'parent-port-identity' in response['ietf-ptp:parent-ds']:
                        response['ietf-ptp:parent-ds']['parent-port-identity']['clock-identity'] = decode_base64(response['ietf-ptp:parent-ds']['parent-port-identity']['clock-identity'])
                    if 'grandmaster-identity' in response['ietf-ptp:parent-ds']:
                        response['ietf-ptp:parent-ds']['grandmaster-identity'] = decode_base64(response['ietf-ptp:parent-ds']['grandmaster-identity'])
                show_cli_output(args[1], response)
            elif func == 'get_ietf_ptp_ptp_instance_list_time_properties_ds':
                show_cli_output(args[1], response)
            elif func == 'get_ietf_ptp_ptp_instance_list':
                show_cli_output(args[1], response)
            elif func == 'get_ietf_ptp_ptp_instance_list_current_ds':
                show_cli_output(args[1], response)
            else:
                return
    else:
        response = api_response.content
        if "ietf-restconf:errors" in response:
            err = response["ietf-restconf:errors"]
            if "error" in err:
                errList = err["error"]

                errDict = {}
                for err_list_dict in errList:
                    for k, v in err_list_dict.iteritems():
                        errDict[k] = v

                        if "error-message" in errDict:
                            print("%Error: " + errDict["error-message"])
                            sys.exit(-1)
                print("%Error: Transaction Failure")
                sys.exit(-1)
        print(api_response.error_message())
        print("%Error: Transaction Failure")
        sys.exit(-1)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    # pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])
