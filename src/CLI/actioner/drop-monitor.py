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

TAM_DEVICE_TABLE_PREFIX    = "TAM_DEVICE_TABLE"
TAM_COLLECTOR_TABLE_PREFIX = "TAM_COLLECTOR_TABLE"
SAMPLE_RATE_TABLE = "SAMPLE_RATE_TABLE"
TAM_DROP_MONITOR_AGING_INTERVAL_TABLE = "TAM_DROP_MONITOR_AGING_INTERVAL_TABLE" 
TAM_DROP_MONITOR_FLOW_TABLE = "TAM_DROP_MONITOR_FLOW_TABLE" 

collectorheader = ['NAME', 'IP TYPE', 'IP', 'PORT']

class DropMon(object):

    def __init__(self):
        # connect CONFIG DB
        self.config_db = ConfigDBConnector()
        self.config_db.connect()

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

    def clear_drop_mon_flow(self, args):
        key = args.flowname
        entry = self.config_db.get_entry(TAM_DROP_MONITOR_FLOW_TABLE, key)
        if entry:
            self.config_db.set_entry(TAM_DROP_MONITOR_FLOW_TABLE, key, None)
        else:
            print "Entry Not Found"
            return False
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
            print "Entry Not Found"
            return False
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
    #Sample Params
    parser.add_argument('-sample', '--samplename', type=str, help='Sample Name')
    parser.add_argument('-rate', '--rate', type=str, help='Sample Rate')

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

    sys.exit(0)

if __name__ == "__main__":
    main()
