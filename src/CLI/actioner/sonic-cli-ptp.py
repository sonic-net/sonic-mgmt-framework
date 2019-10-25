#!/usr/bin/python

import sys
import swsssdk
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector

import urllib3
urllib3.disable_warnings()

PTP_CLOCK = 'PTP_CLOCK'
PTP_GLOBAL = 'GLOBAL'

if __name__ == '__main__':
    pipestr().write(sys.argv)
    db = swsssdk.SonicV2Connector(host='127.0.0.1')
    db.connect(db.CONFIG_DB)
    db.connect(db.COUNTERS_DB)

    config_db = ConfigDBConnector()
    if config_db is None:
        sys.exit()
    config_db.connect()
    if sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_default_ds':
        raw_data = config_db.get_entry(PTP_CLOCK, PTP_GLOBAL)
        api_response = {}
        api_response['ietf-ptp:default-ds'] = raw_data
        show_cli_output(sys.argv[3], api_response)
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_time_properties_ds':
        print "Nothing"
        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_parent_ds':
        print "Nothing"
        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list_port_ds_list':
        print "Nothing"
        sys.exit()
    elif sys.argv[1] == 'get_ietf_ptp_ptp_instance_list':
        raw_data = config_db.get_keys(PTP_CLOCK)
        api_response = {}
        api_response_list = []
        port_ds_dict = {}
        port_ds_list = []
        port_ds_entry = {}
        for key in raw_data:
            if "Ethernet" in key:
                port_ds_entry = {}
                port_ds_entry["port-number"] = key
                port_ds_entry["port-state"] = "online"
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
        config_db.set_entry(PTP_CLOCK, sys.argv[2], data)
    elif sys.argv[1] == 'del_port':
        config_db.set_entry(PTP_CLOCK, sys.argv[2], None)
    else:
        data = {}
        data[sys.argv[1]] = sys.argv[2]
        config_db.mod_entry(PTP_CLOCK, PTP_GLOBAL, data)
