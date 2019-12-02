#!/usr/bin/python

############################################################################
#
# tam is a tool for handling TAM commands.
#
############################################################################

import argparse
import getopt
import json
import os
import re
import sys
import swsssdk
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector
from os import path

TAM_DEVICE_TABLE_PREFIX    = "TAM_DEVICE_TABLE"
TAM_COLLECTOR_TABLE_PREFIX = "TAM_COLLECTOR_TABLE"

collectorheader = ['NAME', 'IP TYPE', 'IP', 'PORT']

class Tam(object):

    def __init__(self):
        # connect CONFIG DB
        self.config_db = ConfigDBConnector()
        self.config_db.connect()

        # connect APPL DB
        self.app_db = ConfigDBConnector()
        self.app_db.db_connect('APPL_DB')

    def get_tam_collector_info(self, k):
        api_response = {}
        key = TAM_COLLECTOR_TABLE_PREFIX + '|' + k
        raw_coll_data = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        api_response['coll-key'] = k 
        api_response['each-coll-data'] = raw_coll_data
        return api_response , raw_coll_data

    def get_print_all_tam_collectors(self, name):
        coll_dict = {}
        coll_list = []
        if name != 'all':
            api_response, entryfound = self.get_tam_collector_info(name)
            if entryfound is not None:
                coll_list.append(api_response)
        else:
            table_data = self.config_db.get_keys(TAM_COLLECTOR_TABLE_PREFIX)
            # Get data for all keys
            for k in table_data:
                api_each_flow_response, entryfound = self.get_tam_collector_info(k)
                if entryfound is not None:
                    coll_list.append(api_each_flow_response)

        coll_dict['flow-list'] = coll_list
        show_cli_output("show_collector.j2", coll_dict)
        return

    def config_device_id(self, args):
        key = 'device'
        entry = self.config_db.get_entry(TAM_DEVICE_TABLE_PREFIX, key)
        if entry is None:
            if args.deviceid: 
                self.config_db.set_entry(TAM_DEVICE_TABLE_PREFIX, key, {'deviceid' : args.deviceid})
        else:
            if args.deviceid:
                entry_value = entry.get('deviceid', [])

                if entry_value != args.deviceid:
                    self.config_db.mod_entry(TAM_DEVICE_TABLE_PREFIX, key, {'deviceid' : args.deviceid})
        return

    def config_collector(self, args):
        if args.iptype == 'ipv4':
           if args.ipaddr == "0.0.0.0":
              print "Collector IP should be non-zero ip address"
              return False

        if args.iptype == 'ipv6':
             print "IPv6 Collector type not supported"
             return False

        self.config_db.mod_entry(TAM_COLLECTOR_TABLE_PREFIX, args.collectorname, {'ipaddress-type' : args.iptype, 'ipaddress' : args.ipaddr, 'port' : args.port})
                             
        return

    def clear_device_id(self):
        key = 'device'
        entry = self.config_db.get_entry(TAM_DEVICE_TABLE_PREFIX, key)
        if entry:
            self.config_db.set_entry(TAM_DEVICE_TABLE_PREFIX, key, None)
        return

    def clear_collector(self, args):
        key = args.collectorname
        entry = self.config_db.get_entry(TAM_COLLECTOR_TABLE_PREFIX, key)
        if entry:
            self.config_db.set_entry(TAM_COLLECTOR_TABLE_PREFIX, key, None)
        else:
            print "Entry Not Found"
            return False
        return

    def show_device_id(self):
        key = TAM_DEVICE_TABLE_PREFIX + '|' + 'device'
        data = self.config_db.get_all(self.config_db.CONFIG_DB, key)
        print "TAM Device identifier"
        print "-------------------------------"
        if data:
            if 'deviceid' in data:
                print "Device Identifier    - ", data['deviceid']
        return

    def show_collector(self, args):
        self.get_print_all_tam_collectors(args.collectorname)
        return

def main():

    parser = argparse.ArgumentParser(description='Handles TAM commands',
                                     version='1.0.0',
                                     formatter_class=argparse.RawTextHelpFormatter,
                                     epilog="""
Examples:
    tam -config -deviceid value
    tam -config -collector collectorname -iptype ipv4/ipv6 -ip ipaddr -port value
    tam -clear -device_id
    tam -clear -collector collectorname
    tam -show -device_id
    tam -show -collector collectorname
""")

    parser.add_argument('-clear', '--clear', action='store_true', help='Clear tam information')
    parser.add_argument('-show', '--show', action='store_true', help='Show tam information')
    parser.add_argument('-config', '--config', action='store_true', help='Config tam information')
    parser.add_argument('-device', '--device', action='store_true', help='tam device identifier')
    parser.add_argument('-deviceid', '--deviceid', type=int, help='tam device identifier')
    parser.add_argument('-collector', '--collectorname', type=str, help='tam collector name')
    parser.add_argument('-iptype', '--iptype', type=str, choices=['ipv4', 'ipv6'], help='tam collector IP type')
    parser.add_argument('-ipaddr', '--ipaddr', type=str, help='tam collector ip')
    parser.add_argument('-port', '--port', type=str, help='tam collector port')
    parser.add_argument('-templ', '--template', action='store_true', help='ifa template to be used')
    parser.add_argument('-showcollector.j2', '--showcollector', action='store_true', help='ifa template for collector to be used')

    args = parser.parse_args()

    tam = Tam()

    if args.config:
        if args.device:
            tam.config_device_id(args)
        elif args.collectorname and args.iptype and args.ipaddr and args.port:
            tam.config_collector(args)
    elif args.clear:
        if args.device:
            tam.clear_device_id()
        elif args.collectorname:
            tam.clear_collector(args)
    elif args.show:
        if args.device:
            tam.show_device_id()
        elif args.collectorname:
            tam.show_collector(args)

    sys.exit(0)

if __name__ == "__main__":
    main()
