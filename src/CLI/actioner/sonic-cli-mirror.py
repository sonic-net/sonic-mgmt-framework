#!/usr/bin/python

############################################################################
#
# mirror is a tool for handling MIRROR commands.
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
from scripts.render_cli import show_cli_output
from os import path

CFG_MIRROR_SESSION_TABLE = "MIRROR_SESSION"
STATE_MIRROR_SESSION_TABLE = "MIRROR_SESSION_TABLE"

def show_session(session_name):
        """
        Show mirror session configuration. Temporary implementation for now. will be modified to Jinja files in next commit.
        :param session_name: Optional. Mirror session name. Filter sessions by specified name.
        :return:
        """
        configdb = ConfigDBConnector()
        configdb.connect()
        statedb = swsssdk.SonicV2Connector(host='127.0.0.1')
        statedb.connect(statedb.STATE_DB)
        sessions_db_info = configdb.get_table(CFG_MIRROR_SESSION_TABLE)
        for key in sessions_db_info.keys():
            state_db_info = statedb.get_all(statedb.STATE_DB, "{}|{}".format(STATE_MIRROR_SESSION_TABLE, key))
            if state_db_info:
                status = state_db_info.get("status", "inactive")
            else:
                status = "error"
            sessions_db_info[key]["status"] = status
        erspan_header = ("Name", "Status", "SRC IP", "DST IP", "GRE", "DSCP", "TTL", "Queue",
                            "Policer", "SRC Port", "Direction")
        span_header = ("Name", "Status", "DST Port", "SRC Port", "Direction")

        erspan_data = []
        span_data = []
        if session_name is None:
            print("\nERSPAN Sessions")
            print("---------------------------------------------------------------------------------------------------------")
            print("%10s %6s %16s %16s %6s %6s %6s %6s %6s %12s %6s" %("Name", "Status", "SRC IP", "DST IP", "GRE", "DSCP", "TTL", "Queue", "Policer", "SRC Port", "Direction"))

        for key, val in sessions_db_info.iteritems():
            if session_name and key != session_name:
                continue

            if "src_ip" in val:
                if session_name and key == session_name:
                    print("\nERSPAN Sessions")
                    print("---------------------------------------------------------------------------------------------------------")
                    print("%10s %6s %16s %16s %6s %6s %6s %6s %6s %12s %6s" %("Name", "Status", "SRC IP", "DST IP", "GRE", "DSCP", "TTL", "Queue", "Policer", "SRC Port", "Direction"))
                print("%10s %6s %16s %16s %6s %6s %6s %6s %6s %12s %6s" %(key, val.get("status", ""), val.get("src_ip", ""), val.get("dst_ip", ""), val.get("gre_type", ""), val.get("dscp", ""),
                                         val.get("ttl", ""), val.get("queue", ""), val.get("policer", ""),
                                                                  val.get("src_port", ""), val.get("direction", "")))
        if session_name is None:
            print("\nSPAN Sessions")
            print("---------------------------------------------------------------------------------------------------------")
            print("%10s %6s %16s %16s %6s" %("Name", "Status", "DST Port", "SRC Port", "Direction"))
        for key, val in sessions_db_info.iteritems():
            if session_name and key != session_name:
                continue
            if "dst_port" in val:
                if session_name and key == session_name:
                    print("\nSPAN Sessions")
                    print("---------------------------------------------------------------------------------------------------------")
                    print("%10s %6s %16s %16s %6s" %("Name", "Status", "DST Port", "SRC Port", "Direction"))
                print("%10s %6s %16s %16s %6s" %(key, val.get("status", ""), val.get("dst_port", ""), val.get("src_port", ""), val.get("direction", "")))


def session(session_name):
    """
    Show mirror session configuration.
    :return:
    """
    show_session(session_name)

def show_mirror(args):
    """
    Add port mirror session
    """
    session(args.session)

def config_span(args):
    """
    Add port mirror session
    """
    config_db = ConfigDBConnector()
    config_db.connect()

    session_info = {
            }

    if args.destination is not None:
        session_info['dst_port'] = args.destination

    if args.source is not None:
        session_info['src_port'] = args.source

    if args.direction is not None:
        session_info['direction'] = args.direction

    if args.dst_ip is not None:
        session_info['dst_ip'] = args.dst_ip

    if args.src_ip is not None:
        session_info['src_ip'] = args.src_ip

    if args.dscp is not None:
        session_info['dscp'] = args.dscp

    if args.ttl is not None:
        session_info['ttl'] = args.ttl

    if args.gre is not None:
        session_info['gre_type'] = args.gre

    if args.source is not None:
        print("sucess. create mirror session " + args.session + " destination " + args.destination + " source " + args.source + " direction " + args.direction)

    if args.dst_ip is not None:
        print("sucess. create mirror session " + args.session + " dst_ip " + args.dst_ip + " src_ip " + args.src_ip + " dscp " + args.dscp + " ttl " + args.ttl)

    config_db.set_entry("MIRROR_SESSION", args.session, session_info)

def remove_span(args):
    """
    Delete mirror session
    """
    config_db = ConfigDBConnector()
    config_db.connect()

    print("sucess. remove mirror session " + args.session)
    config_db.set_entry("MIRROR_SESSION", args.session, None)

def main():

    parser = argparse.ArgumentParser(description='Handles MIRROR commands',
                                     version='1.0.0',
                                     formatter_class=argparse.RawTextHelpFormatter,
                                     epilog="""
Examples:
    mirror -config -deviceid value
    mirror -config -collector collectorname -iptype ipv4/ipv6 -ip ipaddr -port value
    mirror -clear -device_id
    mirror -clear -collector collectorname
    mirror -show -device_id
    mirror -show -collector collectorname
""")

    parser.add_argument('-clear', '--clear', action='store_true', help='Clear mirror information')
    parser.add_argument('-show', '--show', action='store_true', help='Show mirror information')
    parser.add_argument('-config', '--config', action='store_true', help='Config mirror information')
    parser.add_argument('-session', '--session', type=str, help='mirror session name')
    parser.add_argument('-destination', '--destination', help='destination port')
    parser.add_argument('-source', '--source', type=str, help='mirror source port')
    parser.add_argument('-direction', '--direction', type=str, help='mirror direction')
    parser.add_argument('-dst_ip', '--dst_ip', help='ERSPAN destination ip address')
    parser.add_argument('-src_ip', '--src_ip', help='ERSPAN source ip address')
    parser.add_argument('-dscp', '--dscp', help='ERSPAN dscp')
    parser.add_argument('-gre', '--gre', help='ERSPAN gre')
    parser.add_argument('-ttl', '--ttl', help='ERSPAN ttl')

    args = parser.parse_args()

    if args.config:
            config_span(args)
    elif args.clear:
            remove_span(args)
    elif args.show:
            show_mirror(args)

    sys.exit(0)

if __name__ == "__main__":
    main()
