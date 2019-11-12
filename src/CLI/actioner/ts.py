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

TAM_INT_IFA_FLOW_TS_TABLE_PREFIX = "TAM_INT_IFA_TS_FLOW"
TAM_INT_IFA_TS_FEATURE_TABLE_PREFIX = "TAM_INT_IFA_TS_FEATURE_TABLE"

class Ts(object):

    def __init__(self):
        # connect CONFIG DB
        self.config_db = ConfigDBConnector()
        self.config_db.connect()
		
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

    sys.exit(0)

if __name__ == "__main__":
    main()
