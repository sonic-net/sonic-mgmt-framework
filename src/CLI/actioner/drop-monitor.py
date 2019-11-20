#!/usr/bin/python

############################################################################
#
# drop-monitor is a tool for handling DROP MONITOR Feature commands.
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

TAM_COLLECTOR_TABLE_PREFIX = "TAM_COLLECTOR_TABLE"
SAMPLE_RATE_TABLE = "SAMPLE_RATE_TABLE"
TAM_DROP_MONITOR_AGING_INTERVAL_TABLE = "TAM_DROP_MONITOR_AGING_INTERVAL_TABLE"
TAM_DROP_MONITOR_FLOW_TABLE = "TAM_DROP_MONITOR_FLOW_TABLE"
ACL_RULE_TABLE_PREFIX         = "ACL_RULE"
ACL_TABLE_PREFIX              = "ACL_TABLE"

collectorheader = ['NAME', 'IP TYPE', 'IP', 'PORT']

class DropMon(object):

    def __init__(self):
        # connect CONFIG DB
        self.config_db = ConfigDBConnector()
        self.config_db.connect()

        # connect COUNTERS_DB
        self.counters_db = ConfigDBConnector()
        self.counters_db.db_connect('COUNTERS_DB')

        # connect APPL DB
        self.app_db = ConfigDBConnector()
        self.app_db.db_connect('APPL_DB')


    def config_drop_mon(self, args):
        self.config_db.mod_entry(TAM_DROP_MONITOR_FLOW_TABLE, args.flowname, {'acl-table' : args.acl_table, 'acl-rule' : args.acl_rule, 'collector' : args.dropcollector, 'sample' : args.dropsample})
        return

    def config_drop_mon_aging(self, args):
        self.config_db.mod_entry(TAM_DROP_MONITOR_AGING_INTERVAL_TABLE, "aging", {'aging-interval' : args.aginginterval})
        return

    def config_drop_mon_sample(self, args):
        self.config_db.mod_entry(SAMPLE_RATE_TABLE, args.samplename, {'sampling-rate' : args.rate})
        return

    def clear_single_drop_mon_flow(self, key):
        entry = self.config_db.get_entry(TAM_DROP_MONITOR_FLOW_TABLE, key)
        if entry:
            self.config_db.set_entry(TAM_DROP_MONITOR_FLOW_TABLE, key, None)
        else:
            return False
        return

    def clear_drop_mon_flow(self, args):
        key = args.flowname
        if key == "all":
            # Get all the flow keys
            table_data = self.config_db.get_keys(TAM_DROP_MONITOR_FLOW_TABLE)
            if not table_data:
                return True
            # Clear each flow key
            for key in table_data:
                self.clear_single_drop_mon_flow(key)
        else:
            # Clear the specified flow entry
            self.clear_single_drop_mon_flow(key)

        return

    def clear_drop_mon_sample(self, args):
        key = args.samplename
        entry = self.config_db.get_entry(SAMPLE_RATE_TABLE, key)
        if entry:
            self.config_db.set_entry(SAMPLE_RATE_TABLE, key, None)
        else:
            print "Entry Not Found"
            return False
        return

    def clear_drop_mon_aging_int(self, args):
        key = "aging" 
        entry = self.config_db.get_entry(TAM_DROP_MONITOR_AGING_INTERVAL_TABLE, key)
        if entry:
            self.config_db.set_entry(TAM_DROP_MONITOR_AGING_INTERVAL_TABLE, key, None)
        else:
            return False
        return

    def show_flow(self, args):
        self.get_print_all_dropmon_flows(args.flowname)
        return

    def get_dropmon_flow_stat(self, flowname): 
        api_response_stat = {}
        api_response, entryfound = self.get_dropmon_flow_info(flowname)
        api_response_stat['flow-name'] = flowname 
        if entryfound is not None:
            for k in api_response:
               if k == "ietf-ts:each-flow-data":
                  acl_rule = api_response['ietf-ts:each-flow-data']['acl-rule'] 
                  acl_table = api_response['ietf-ts:each-flow-data']['acl-table']
                  api_response_stat['rule-name'] = acl_rule
                  api_response_stat['table-name'] = acl_table

        acl_rule_keys = self.config_db.get_keys(ACL_RULE_TABLE_PREFIX)
        for acl_rule_key in acl_rule_keys:
            if acl_rule_key[1] == acl_rule:
               acl_counter_key = 'COUNTERS:' + acl_rule_key[0] + ':' + acl_rule_key[1]
               raw_dropmon_stats = self.counters_db.get_all(self.counters_db.COUNTERS_DB, acl_counter_key)
               api_response_stat['ietf-ts:dropmon-stats'] = raw_ifa_stats

        return api_response_stat, entryfound

    def get_print_all_dropmon_stats(self, name):
        stat_dict = {}
        stat_list = []
        if name != 'all':
            api_response, entryfound = self.get_dropmon_flow_stat(name)
            if entryfound is not None:
                stat_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(TAM_DROP_MONITOR_FLOW_TABLE)
            # Get data for all keys
            for k in table_data:
                api_each_stat_response, entryfound = self.get_dropmon_flow_stat(k)
                if entryfound is not None:
                    stat_list.append(api_each_stat_response)

        stat_dict['stat-list'] = stat_list
        show_cli_output("show_statistics_flow.j2", stat_dict)
        return

    def show_statistics(self, args):
        self.get_print_all_dropmon_stats(args.flowname)
        return

    def show_aging_interval(self, args):
        key = "aging"
        entry = self.config_db.get_entry(TAM_DROP_MONITOR_AGING_INTERVAL_TABLE, key)
        if entry:
            print "Aging interval : {}".format(entry['aging-interval'])
        return

    def show_sample(self, args):
        self.get_print_all_sample(args.samplename)
        return

    def get_dropmon_flow_info(self, k):
        flow_data = {}
        flow_data['acl-table-name'] = ''
        flow_data['sampling-rate'] = ''
        flow_data['collector'] = ''

        api_response = {}
        key = TAM_DROP_MONITOR_FLOW_TABLE + '|' + k
        raw_flow_data = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        if raw_flow_data:
            sample = raw_flow_data['sample']
            rate = self.config_db.get_entry(SAMPLE_RATE_TABLE, sample)
            raw_flow_data['sample'] = rate['sampling-rate']
        api_response['ietf-ts:flow-key'] = k
        api_response['ietf-ts:each-flow-data'] = raw_flow_data
        return api_response , raw_flow_data

    def get_print_all_dropmon_flows(self, name):
        flow_dict = {}
        flow_list = []
        if name != 'all':
            api_response, entryfound = self.get_dropmon_flow_info(name)
            if entryfound is not None:
                flow_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(TAM_DROP_MONITOR_FLOW_TABLE)
            # Get data for all keys
            for k in table_data:
                api_each_flow_response, entryfound = self.get_dropmon_flow_info(k)
                if entryfound is not None:
                    flow_list.append(api_each_flow_response)

        flow_dict['flow-list'] = flow_list
        show_cli_output("show_drop_monitor_flow.j2", flow_dict)
        return

    def get_sample_info(self, k):
        sample_data = {}
        sample_data['sampling-rate'] = ''

        api_response = {}
        key = SAMPLE_RATE_TABLE + '|' + k
        raw_sample_data = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        api_response['ietf-ts:sample-key'] = k
        api_response['ietf-ts:each-sample-data'] = raw_sample_data
        return api_response , raw_sample_data

    def get_print_all_sample(self, name):
        sample_dict = {}
        sample_list = []
        if name != 'all':
            api_response, entryfound = self.get_sample_info(name)
            if entryfound is not None:
                sample_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(SAMPLE_RATE_TABLE)
            # Get data for all keys
            for k in table_data:
                api_each_flow_response, entryfound = self.get_sample_info(k)
                if entryfound is not None:
                    sample_list.append(api_each_flow_response)

        sample_dict['sample-list'] = sample_list
        show_cli_output("show_sample.j2", sample_dict)
        return

def main():

    parser = argparse.ArgumentParser(description='Handles MoD commands',
                                     version='1.0.0',
                                     formatter_class=argparse.RawTextHelpFormatter,
                                     epilog="""
Examples:
    drop-monitor -config -dropmonitor -flow flowname  --acl_table acltablename --acl_rule aclrulename --dropcollector collectorname --dropsample samplename
    drop-monitor -config -dropmonitor --aginginterval interval
    drop-monitor -config -sample samplename --rate samplingrate
""")

    parser.add_argument('-clear', '--clear', action='store_true', help='Clear mod information')
    parser.add_argument('-show', '--show', action='store_true', help='Show mod information')
    parser.add_argument('-config', '--config', action='store_true', help='Config mod information')
    #Drop Monitor params
    parser.add_argument('-dropmonitor', '--dropmonitor', action='store_true', help='Configure Drop Monitor')
    parser.add_argument('-flow', '--flowname', type=str, help='Flowname')
    parser.add_argument('-acl_table', '--acl_table', type=str, help='ACL Table Name')
    parser.add_argument('-acl_rule', '--acl_rule', type=str, help='ACL Rule Name')
    parser.add_argument('-dropcollector', '--dropcollector', type=str, help='Drop Monitor Collector Name')
    parser.add_argument('-dropsample', '--dropsample', type=str, help='Drop Monitor Sample Name')
    parser.add_argument('-aginginterval', '--aginginterval', type=int, help='Drop Monitor Aging Interval')
    parser.add_argument('-rate', '--rate', type=int, help='Sampling rate')
    #Sample Params
    parser.add_argument('-sample', '--samplename', type=str, help='Sample Name')
    parser.add_argument('-statistics', '--statistics', action='store_true', help='drop monitor statistics')
    parser.add_argument('-templ', '--template', action='store_true', help='drop monitor template to be used')
    parser.add_argument('-showflow.j2', '--showflow', action='store_true', help='dropmon flow to be used')
    parser.add_argument('-showstatisticsflow.j2', '--showstatistics', action='store_true', help='Flow statistics')
    parser.add_argument('-showsample.j2', '--showsample', action='store_true', help='Sample configuration')

    args = parser.parse_args()

    dropmon = DropMon()

    if args.config:
        if args.dropmonitor:
             if args.aginginterval:
                dropmon.config_drop_mon_aging(args)
             elif args.flowname and args.acl_table and args.acl_rule and args.dropcollector and args.dropsample:
                dropmon.config_drop_mon(args)
        elif args.samplename and args.rate:
             dropmon.config_drop_mon_sample(args)
    elif args.clear:
        if args.dropmonitor:
             if args.flowname:
                dropmon.clear_drop_mon_flow(args)
             else:
                dropmon.clear_drop_mon_aging_int(args)
        elif args.samplename:
             dropmon.clear_drop_mon_sample(args)
    elif args.show:
        if args.statistics:
            dropmon.show_statistics(args)
        elif args.flowname:
            dropmon.show_flow(args)
        elif args.aginginterval:
            dropmon.show_aging_interval(args)
        elif args.samplename:
            dropmon.show_sample(args)

    sys.exit(0)

if __name__ == "__main__":
    main()
