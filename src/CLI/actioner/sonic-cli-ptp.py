#!/usr/bin/python

import sys
import swsssdk
import socket
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector

import urllib3
urllib3.disable_warnings()

PTP_CLOCK = 'PTP_CLOCK'
PTP_PORT = 'PTP_PORT|GLOBAL'
PTP_GLOBAL = 'GLOBAL'


def node_addr_type(address):
    try:
        socket.inet_pton(socket.AF_INET, address)
    except:
        try:
            socket.inet_pton(socket.AF_INET6, address)
        except:
            return "mac"

        return "ipv6"

    return "ipv4"


def port_state_to_str(state_num):
    port_state_tbl = {"0": "none", "1": "initializing", "2": "faulty", "3": "disabled", "4": "listening",
                      "5": "pre_master", "6": "master", "7": "passive", "8": "uncalibrated", "9": "slave"}
    return port_state_tbl[state_num]


def get_attrib(c, attrib):
    attrib_val = c.get(attrib, 'None')
    return attrib_val


def check_network_transport_allowed(c, network_transport, unicast_multicast):
    domain_profile = get_attrib(c, 'domain-profile')
    domain_number = int(get_attrib(c, 'domain-number'))
    clock_type = get_attrib(c, 'clock-type')
    if domain_profile == 'G.8275.x':
        if network_transport == "L2":
            print "%Error: L2 not supported with G.8275.2"
            return 0
        if unicast_multicast == "multicast":
            print "%Error: multicast not supported with G.8275.2"
            return 0
        if domain_number < 44 or domain_number > 63:
            print "%Error: domain must be in range 44-63 with G.8275.2"
            return 0
        if clock_type == 'P2P_TC' or clock_type == 'E2E_TC':
            print "%Error: transparent-clock not supported with G.8275.2"
            return 0
    return 1


def check_domain_number_allowed(c, domain_number):
    domain_profile = get_attrib(c, 'domain-profile')
    if domain_profile == 'G.8275.x':
        if domain_number < 44 or domain_number > 63:
            print "%Error: domain must be in range 44-63 with G.8275.2"
            return 0
    return 1


def check_clock_type_allowed(c, clock_type):
    domain_profile = get_attrib(c, 'domain-profile')
    if domain_profile == 'G.8275.x':
        if clock_type == 'P2P_TC' or clock_type == 'E2E_TC':
            print "%Error: transparent-clock not supported with G.8275.2"
            return 0
    return 1


def check_master_table_allowed(c):
    unicast_multicast = get_attrib(c, 'unicast-multicast')
    if unicast_multicast == 'multicast':
        print "%Error: master-table is not needed in with multicast transport"
        return 0
    return 1


def check_domain_profile_allowed(c, domain_profile):
    network_transport = get_attrib(c, 'network-transport')
    unicast_multicast = get_attrib(c, 'unicast-multicast')
    domain_number = int(get_attrib(c, 'domain-number'))
    if domain_profile == 'G.8275.2':
        if network_transport == "L2":
            print "%Error: G.8275.2 not supported with L2 transport"
            return 0
        if unicast_multicast == "multicast":
            print "%Error: G.8275.2 not supported with multicast transport"
            return 0
        if domain_number < 44 or domain_number > 63:
            print "%Error: domain must be in range 44-63 with G.8275.2"
            return 0
    return 1


if __name__ == '__main__':
    pipestr().write(sys.argv)
    db = swsssdk.SonicV2Connector(host='127.0.0.1')
    db.connect(db.STATE_DB)

    config_db = ConfigDBConnector()
    if config_db is None:
        sys.exit()
    config_db.connect()

    clock_global = db.get_all(db.STATE_DB, "PTP_CLOCK|GLOBAL")

    if sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_default_ds':
        raw_data = clock_global
        if not raw_data:
            sys.exit()
        else:
            api_response = {}
            api_inner_response = {}
            api_clock_quality_response = {}

            for key, val in raw_data.items():
                if key == "clock-class" or key == "clock-accuracy" or key == "offset-scaled-log-variance":
                    api_clock_quality_response[key] = val
                else:
                    api_inner_response[key] = val

                if bool(api_clock_quality_response):
                    api_inner_response["clock-quality"] = api_clock_quality_response
                api_response['ietf-ptp:default-ds'] = api_inner_response

        show_cli_output(sys.argv[3], api_response)

        raw_data = db.get_all(db.STATE_DB, "PTP_CURRENTDS|GLOBAL")
        if not raw_data:
            sys.exit()
        for key, val in raw_data.items():
            if key == "mean-path-delay":
                print("%-21s %s") % ("Mean Path Delay", val)
            if key == "steps-removed":
                print("%-21s %s") % ("Steps Removed", val)
            if key == "offset-from-master":
                print("%-21s %s") % ("Ofst From Master", val)
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_time_properties_ds':
        raw_data = db.get_all(db.STATE_DB, "PTP_TIMEPROPDS|GLOBAL")
        if not raw_data:
            sys.exit()
        api_response = {}
        api_inner_response = {}

        for key, val in raw_data.items():
            if key == "time-traceable" or key == "frequency-traceable" or key == "ptp-timescale" or key == "leap59" or key == "leap61" or key == "current-utc-offset-valid":
                if val == "0":
                    val = "false"
                else:
                    val = "true"
            api_inner_response[key] = val

        api_response['ietf-ptp:time-properties-ds'] = api_inner_response
        show_cli_output(sys.argv[3], api_response)

        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_parent_ds':
        raw_data = db.get_all(db.STATE_DB, "PTP_PARENTDS|GLOBAL")
        if not raw_data:
            sys.exit()
        api_response = {}
        api_inner_response = {}
        api_parent_id_response = {}
        api_gm_response = {}

        for key, val in raw_data.items():
            if key == "parent-stats":
                if val == "0":
                    val = "false"
                else:
                    val = "true"
            if key == "clock-identity" or key == "port-number":
                api_parent_id_response[key] = val
            elif key == "clock-class" or key == "clock-accuracy" or key == "offset-scaled-log-variance":
                api_gm_response[key] = val
            else:
                api_inner_response[key] = val

        api_inner_response["parent-port-identity"] = api_parent_id_response
        api_inner_response["grandmaster-clock-quality"] = api_gm_response
        api_response['ietf-ptp:parent-ds'] = api_inner_response

        show_cli_output(sys.argv[3], api_response)

        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_port_ds_list':
        raw_data = db.get_all(db.STATE_DB, "PTP_PORT|GLOBAL|" + sys.argv[3])
        if not raw_data:
            sys.exit()
        api_response = {}
        api_response_list = []
        api_inner_response = {}

        for key, val in raw_data.items():
            if key == "port-state":
                val = port_state_to_str(val)
            if key == "delay-mechanism":
                if val == "1":
                    val = "e2e"
                if val == "2":
                    val = "p2p"

            api_inner_response[key] = val

        api_response_list.append(api_inner_response)
        api_response['ietf-ptp:port-ds-list'] = api_response_list

        show_cli_output(sys.argv[4], api_response)
        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list':
        raw_data = db.keys(db.STATE_DB, "PTP_PORT|GLOBAL|*")
        if not raw_data:
            sys.exit()
        api_response = {}
        api_response_list = []
        port_ds_dict = {}
        port_ds_list = []
        port_ds_entry = {}
        for key in raw_data:
            port_ds_entry = {}
            port_ds_entry["port-number"] = key.replace("PTP_PORT|GLOBAL|", "")
            state_data = db.get_all(db.STATE_DB, key)

            port_ds_entry["port-state"] = port_state_to_str(state_data["port-state"])
            port_ds_list.append(port_ds_entry)
        port_ds_dict['port-ds-list'] = port_ds_list
        api_response_list.append(port_ds_dict)
        api_response['ietf-ptp:instance_list'] = api_response_list
        show_cli_output(sys.argv[3], api_response)
    elif sys.argv[1] == 'patch_ietf_ptp_ptp_instance_list_default_ds_domain_number':
        data = {}
        if not check_domain_number_allowed(clock_global, sys.argv[3]):
            sys.exit()
        data['domain-number'] = sys.argv[3]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'patch_ietf_ptp_ptp_instance_list_default_ds_priority1':
        data = {}
        data['priority1'] = sys.argv[3]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'patch_ietf_ptp_ptp_instance_list_default_ds_priority2':
        data = {}
        data['priority2'] = sys.argv[3]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'patch_ietf_ptp_ptp_instance_list_default_ds_two_step_flag':
        data = {}
        if sys.argv[3] == "enable":
            data['two-step-flag'] = '1'
        else:
            data['two-step-flag'] = '0'
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'patch_ietf_ptp_ptp_transparent_clock_default_ds_delay_mechanism':
        if sys.argv[2] == 'P2P':
            print "%Error: peer-to-peer is not supported"
            sys.exit()
        data = {}
        data['tc-delay-mechanism'] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'add_port':
        data = {}
        data['enable'] = '1'
        config_db.set_entry(PTP_PORT, sys.argv[2], data)
    elif sys.argv[1] == 'del_port':
        config_db.set_entry(PTP_PORT, sys.argv[2], None)
    elif sys.argv[1] == 'clock-type':
        if sys.argv[2] == 'P2P_TC':
            print "%Error: peer-to-peer-transparent-clock is not supported"
            sys.exit()
        if not check_clock_type_allowed(clock_global, sys.argv[2]):
            sys.exit()
        data = {}
        data[sys.argv[1]] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'domain-profile':
        data = {}
        if sys.argv[2] == 'G.8275.1':
            print "%Error: g8275.1 is not supported"
            sys.exit()
        elif sys.argv[2] == 'G.8275.2':
            data[sys.argv[1]] = 'G.8275.x'
        else:
            data[sys.argv[1]] = sys.argv[2]
        if not check_domain_profile_allowed(clock_global, sys.argv[2]):
            sys.exit()
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'add_master_table':
        tbl = db.get_all(db.STATE_DB, "PTP_PORT|GLOBAL|" + sys.argv[2])
        nd_list = []
        if not tbl:
            print "%Error: " + sys.argv[2] + " has not been added"
            sys.exit()

        uc_tbl = tbl.get('unicast-table', 'None')
        if uc_tbl != 'None':
            nd_list = uc_tbl.split(',')
        if sys.argv[3] in nd_list:
            # entry already exists
            sys.exit()
        if len(nd_list) == 1 and nd_list[0] == '':
            nd_list = []
        if len(nd_list) > 0 and node_addr_type(nd_list[0]) != node_addr_type(sys.argv[3]):
            print "%Error: Mixed address types not allowed"
            sys.exit()
        if len(nd_list) >= 8:
            print "%Error: maximum 8 nodes"
            sys.exit()
        if not check_master_table_allowed(clock_global):
            sys.exit()
        nd_list.append(sys.argv[3])
        value = ','.join(nd_list)
        data = {}
        data['unicast-table'] = value
        config_db.mod_entry("PTP_PORT|GLOBAL", sys.argv[2], data)
    elif sys.argv[1] == 'del_master_table':
        tbl = db.get_all(db.STATE_DB, "PTP_PORT|GLOBAL|" + sys.argv[2])
        nd_list = []
        if tbl:
            uc_tbl = tbl.get('unicast-table', 'None')
            if uc_tbl != 'None':
                nd_list = uc_tbl.split(',')
        if sys.argv[3] not in nd_list:
            # entry doesn't exists
            sys.exit()

        nd_list.remove(sys.argv[3])
        value = ','.join(nd_list)
        data = {}
        data['unicast-table'] = value
        config_db.mod_entry("PTP_PORT|GLOBAL", sys.argv[2], data)
    elif sys.argv[1] == 'network-transport':
        data = {}
        if not check_network_transport_allowed(clock_global, sys.argv[2], sys.argv[3]):
            sys.exit()
        data[sys.argv[1]] = sys.argv[2]
        data["unicast-multicast"] = sys.argv[3]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    else:
        data = {}
        data[sys.argv[1]] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    db.close(db.STATE_DB)
