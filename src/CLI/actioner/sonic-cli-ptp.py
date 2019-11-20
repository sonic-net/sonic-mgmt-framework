#!/usr/bin/python

import sys
import swsssdk
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector

import urllib3
urllib3.disable_warnings()

PTP_CLOCK = 'PTP_CLOCK'
PTP_PORT = 'PTP_PORT|GLOBAL'
PTP_GLOBAL = 'GLOBAL'


def port_state_to_str(state_num):
    outval = ""
    if state_num == "1":
        outval = "initializing"
    if state_num == "2":
        outval = "faulty"
    if state_num == "3":
        outval = "disabled"
    if state_num == "4":
        outval = "listening"
    if state_num == "5":
        outval = "pre_master"
    if state_num == "6":
        outval = "master"
    if state_num == "7":
        outval = "passive"
    if state_num == "8":
        outval = "uncalibrated"
    if state_num == "9":
        outval = "slave"

    return outval


if __name__ == '__main__':
    pipestr().write(sys.argv)
    db = swsssdk.SonicV2Connector(host='127.0.0.1')
    db.connect(db.STATE_DB)

    config_db = ConfigDBConnector()
    if config_db is None:
        sys.exit()
    config_db.connect()
    if sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_default_ds':
        raw_data = db.get_all(db.STATE_DB, "PTP_CLOCK|GLOBAL")
        if not raw_data:
            raw_data = config_db.get_entry(PTP_CLOCK, PTP_GLOBAL)
            api_response = {}
            api_response['ietf-ptp:default-ds'] = raw_data
        else:
            api_response = {}
            api_inner_response = {}
            api_clock_quality_response = {}

            for key, val in raw_data.items():
                if key == "clock-class" or key == "clock-accuracy" or key == "offset-scaled-log-variance":
                    api_clock_quality_response[key] = val
                else:
                    api_inner_response[key] = val

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
        raw_data = db.get_all(db.STATE_DB, "PTP_PORT|GLOBAL|Ethernet" + sys.argv[3])
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
        data = {}
        data['tc-delay-mechanism'] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    elif sys.argv[1] == 'add_port':
        data = {}
        data['enable'] = '1'
        config_db.set_entry(PTP_PORT, sys.argv[2], data)
    elif sys.argv[1] == 'del_port':
        config_db.set_entry(PTP_PORT, sys.argv[2], None)
    else:
        data = {}
        data[sys.argv[1]] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
    db.close(db.STATE_DB)
