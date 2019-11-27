#!/usr/bin/python

############################################################################
#
# ts is a tool for handling TAM INT IFA TS  commands.
#
############################################################################

import argparse
import getopt
import json
import os
import re
import sys
import swsssdk
from swsssdk import ConfigDBConnector
from os import path
from scripts.render_cli import show_cli_output

TAM_INT_IFA_FLOW_TS_TABLE_PREFIX = "TAM_INT_IFA_TS_FLOW"
TAM_INT_IFA_TS_FEATURE_TABLE_PREFIX = "TAM_INT_IFA_TS_FEATURE_TABLE"
TAM_DEVICE_TABLE_PREFIX       = "TAM_DEVICE_TABLE"
TAM_COLLECTOR_TABLE_PREFIX    = "TAM_COLLECTOR_TABLE"
ACL_RULE_TABLE_PREFIX         = "ACL_RULE"
ACL_TABLE_PREFIX              = "ACL_TABLE"

class Ts(object):

    def __init__(self):
        # connect CONFIG DB
        self.config_db = ConfigDBConnector()
        self.config_db.connect()

        # connect COUNTER DB
        self.counters_db = ConfigDBConnector()
        self.counters_db.db_connect('COUNTERS_DB')

        # connect APPL DB
        self.app_db = ConfigDBConnector()
        self.app_db.db_connect('APPL_DB')

    def config_enable(self, args):
        """ Enable ifa """
        key = 'feature'
        self.config_db.set_entry(TAM_INT_IFA_TS_FEATURE_TABLE_PREFIX, key, {'enable' :"true"})
        print "Enabled IFA"

        return
    
    def config_disable(self, args):
        """ Disable ifa """
        key = 'feature'
        self.config_db.set_entry(TAM_INT_IFA_TS_FEATURE_TABLE_PREFIX, key, {'enable' :"false"})
        print "Disabled IFA"

        return

    def config_flow(self, args):
        key = TAM_INT_IFA_FLOW_TS_TABLE_PREFIX + '|' + args.flowname
        entry = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        if entry is None:
            if args.acl_table_name:
                self.config_db.mod_entry(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX, args.flowname, {'acl-table-name' : args.acl_table_name})
            if args.acl_rule_name:
                self.config_db.mod_entry(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX, args.flowname, {'acl-rule-name' : args.acl_rule_name})
        else:
            print "Entry Already Exists"
            return False
        return

    def clear_each_flow(self, flowname):
        entry = self.config_db.get_entry(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX, flowname)
        if entry:
           self.config_db.set_entry(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX, flowname, None)
        else:
           print "Entry Not Found"
           return False

        return

    def clear_flow(self, args):
        key = args.flowname
        if key == "all":
            # Get all the flow keys 
            table_data = self.config_db.get_keys(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX)
            if not table_data:
                return True
            # Clear each flow key
            for key in table_data:
                self.clear_each_flow(key)
        else:
            # Clear the specified flow entry
            self.clear_each_flow(key)

        return

    def show_flow(self, args):
        self.get_print_all_ifa_flows(args.flowname)
        return

    def show_status(self):
        # Get data for all keys
        flowtable_keys = self.config_db.get_keys(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX)

        api_response = {}
        key = TAM_INT_IFA_TS_FEATURE_TABLE_PREFIX + '|' + 'feature'
        raw_data_feature = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        api_response['ietf-ts:feature-data'] = raw_data_feature
        api_inner_response = {}
        api_inner_response["num-of-flows"] = len(flowtable_keys)
        api_response['ietf-ts:num-of-flows'] = api_inner_response 
        key = TAM_DEVICE_TABLE_PREFIX + '|' + 'device'
        raw_data_device = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        api_response['ietf-ts:device-data'] = raw_data_device
        show_cli_output("show_status.j2", api_response)

        return

    def get_ifa_flow_stat(self, flowname):
        api_response_stat = {}
        api_response, entryfound = self.get_ifa_flow_info(flowname)
        if entryfound is None:
          return api_response_stat, entryfound

        api_response_stat['flow-name'] = flowname
        if entryfound is not None:
            for k in api_response:
               if k == "ietf-ts:each-flow-data":
                  acl_rule_name = api_response['ietf-ts:each-flow-data']['acl-rule-name']
                  acl_table_name = api_response['ietf-ts:each-flow-data']['acl-table-name']
                  api_response_stat['rule-name'] = acl_rule_name
                  api_response_stat['table-name'] = acl_table_name

        acl_rule_keys = self.config_db.get_keys(ACL_RULE_TABLE_PREFIX)
        for acl_rule_key in acl_rule_keys:
            if acl_rule_key[1] == acl_rule_name:
               acl_counter_key = 'COUNTERS:' + acl_rule_key[0] + ':' + acl_rule_key[1]
               raw_ifa_stats = self.counters_db.get_all(self.counters_db.COUNTERS_DB, acl_counter_key)
               api_response_stat['ietf-ts:ifa-stats'] = raw_ifa_stats

        return api_response_stat, entryfound

    def get_print_all_ifa_stats(self, name):
        stat_dict = {}
        stat_list = []
        if name != 'all':
            api_response, entryfound = self.get_ifa_flow_stat(name)
            if entryfound is not None:
                stat_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX)
            # Get data for all keys
            for k in table_data:
                api_each_stat_response, entryfound = self.get_ifa_flow_stat(k)
                if entryfound is not None:
                    stat_list.append(api_each_stat_response)

        stat_dict['stat-list'] = stat_list
        show_cli_output("show_statistics_flow.j2", stat_dict)
        return

    def show_statistics(self, args):
        self.get_print_all_ifa_stats(args.flowname)
        return


    def get_ifa_flow_info(self, k):
        flow_data = {}
        flow_data['acl-table-name'] = ''
        flow_data['sampling-rate'] = ''
        flow_data['collector'] = ''

        api_response = {}
        key = TAM_INT_IFA_FLOW_TS_TABLE_PREFIX + '|' + k
        raw_flow_data = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        api_response['ietf-ts:flow-key'] = k 
        api_response['ietf-ts:each-flow-data'] = raw_flow_data
        return api_response , raw_flow_data

    def get_print_all_ifa_flows(self, name):
        flow_dict = {}
        flow_list = []
        if name != 'all':
            api_response, entryfound = self.get_ifa_flow_info(name)
            if entryfound is not None:
                flow_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(TAM_INT_IFA_FLOW_TS_TABLE_PREFIX)
            # Get data for all keys
            for k in table_data:
                api_each_flow_response, entryfound = self.get_ifa_flow_info(k)
                if entryfound is not None:
                    flow_list.append(api_each_flow_response)

        flow_dict['flow-list'] = flow_list
        show_cli_output("show_flow.j2", flow_dict)
        return

    def get_ifa_supported_info(self):
        key = 'TAM_INT_IFA_TS_FEATURE_TABLE|feature'
        data = self.config_db.get_all(self.config_db.CONFIG_DB, key)

        if data is None:
           return

        if data['enable'] == "true" :
            print "TAM INT IFA TS Supported - True"
            return True
        elif data['enable'] == "false" :
            print "TAM INT IFA TS Supported - False "
            return False

        return

    def get_ifa_enabled_info(self):
        print "In get_ifa_enabled_info"
        key = 'SWITCH_TABLE:switch'
        data = self.app_db.get(self.app_db.APPL_DB, key, 'ifa_enabled')

        if data and data == 'True':
            return True

        return True


def main():

    parser = argparse.ArgumentParser(description='Handles MoD commands',
                                     version='1.0.0',
                                     formatter_class=argparse.RawTextHelpFormatter,
                                     epilog="""
Examples:
    ts -config -flow flowname -acl_table acl_table_name -acl_rule acl_rule_name
    ts -config -enable
    ts -config -disable
    ts -clear -flow flowname
""")

    parser.add_argument('-clear', '--clear', action='store_true', help='Clear tam_int_ifa information')
    parser.add_argument('-show', '--show', action='store_true', help='Show tam_int_ifa information')
    parser.add_argument('-config', '--config', action='store_true', help='Config tam_int_ifa information')
    parser.add_argument('-enable', '-enable', action='store_true', help='Enable Tam Int Ifa')
    parser.add_argument('-disable', '-disable', action='store_true', help='Disable Tam Int Ifa')
    parser.add_argument('-flow', '--flowname', type=str, help='ifa flow name')
    parser.add_argument('-acl_table_name', '--acl_table_name', type=str, help='ifa acl table name')
    parser.add_argument('-acl_rule_name', '--acl_rule_name', type=str, help='ifa acl rule name')
    parser.add_argument('-status', '--status', action='store_true', help='ifa status')
    parser.add_argument('-supported', '--supported', action='store_true', help='Check if tam-int-ifa supported')
    parser.add_argument('-statistics', '--statistics', action='store_true', help='ifa statistics')
    parser.add_argument('-templ', '--template', action='store_true', help='ifa template to be used')
    parser.add_argument('-showsupported.j2', '--showsupported', action='store_true', help='ifa template to be used')
    parser.add_argument('-showstatus.j2', '--showstatus', action='store_true', help='ifa status to be used')
    parser.add_argument('-showflow.j2', '--showflow', action='store_true', help='ifa flow to be used')
    parser.add_argument('-showstatisticsflow.j2', '--showstatistics', action='store_true', help='ifa statistics to be used')


    args = parser.parse_args()

    ts = Ts()


    if args.config:
        if args.enable:
            ts.config_enable(args)
        elif args.disable:
            ts.config_disable(args)
        elif args.flowname:
            ts.config_flow(args)

    elif args.clear:
        if args.flowname:
          ts.clear_flow(args)
    elif args.show:
        if args.status:
            ts.show_status()
        elif args.statistics:
            ts.show_statistics(args)
        elif args.flowname:
            ts.show_flow(args)
        elif args.supported:
            ts.get_ifa_supported_info()

    sys.exit(0)

if __name__ == "__main__":
    main()
