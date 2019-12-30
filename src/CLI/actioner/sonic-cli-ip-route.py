import sys
import time
import json
import ast
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output


def invoke_api(func, args):
    body = None
    prefix=""
    vrfname="default"

    api = cc.ApiClient()

    if len(args) < 4:
        print("Invalid arguments")
        return

    try:
        id = args.index("route")
    except  ValueError:
        print("Invalid arguments")
        return

    args = args[id+1:]

    if len(args) < 2:
        vrfname = "default"
        if len(args) == 1:
            prefix = args[0]
    elif len(args) < 4:
        vrfname = args[1]
        if len(args) == 3:
            prefix = args[2]

    if func == 'get_network_instance_afts_ipv4_unicast':
        if len(prefix) > 0:
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}'\
                          '/afts/ipv4-unicast/ipv4-entry={prfxKey}',name=vrfname, prfxKey=prefix)
        else:
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}'\
                          '/afts/ipv4-unicast',name=vrfname)
        return api.get(keypath)
    elif func == 'get_network_instance_afts_ipv6_unicast':
        if len(prefix) > 0:
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}'\
                          '/afts/ipv6-unicast/ipv6-entry={prfxKey}',name=vrfname, prfxKey=prefix)
        else:
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={name}'\
                          '/afts/ipv6-unicast',name=vrfname)
        return api.get(keypath)


def parse_nextHops(route, nexthops):
    nexthopsLst = []

    if "next-hop" in nexthops:
        nextHopLst = nexthops["next-hop"]
        for nextHop in nextHopLst:
            nexthop = {}
            if "index" in nextHop:
                index = nextHop["index"].encode('ascii', 'ignore')
                nexthop.update({"index":index}) 
            if "interface-ref" in nextHop:
                interfaceRef = nextHop["interface-ref"]
                if "state" in interfaceRef:
                    state = interfaceRef["state"]
                    if "interface" in state:
                        interface = state["interface"].encode('ascii', 'ignore')
                        nexthop.update({"interface":interface})
            if "state" in nextHop:
                state = nextHop["state"]
                if "ip-address" in state:
                    gateway= "via " + state["ip-address"].encode('ascii', 'ignore')
                    nexthop.update({"gateway":gateway})
                elif "directlyConnected" in state:
                    if state["directlyConnected"] == True:
                        nexthop.update({"gateway":"Direct"})
            nexthopsLst.append(nexthop)    
        route.update({"nexthops": nexthopsLst})

def getOriginProtocol(oProtocol):
    protocol = "" 
    if oProtocol == "openconfig-policy-types:DIRECTLY_CONNECTED":
        protocol = "C"
    elif oProtocol == "openconfig-policy-types:STATIC":
        protocol = "S"
    elif oProtocol == "openconfig-policy-types:OSPF" or oProtocol == "openconfig-policy-types:OSPF3":
        protocol = "O"
    elif oProtocol == "openconfig-policy-types:BGP":
        protocol = "B"
    return protocol

def parse_ipv4Entry(routeLst, response):

    ipEntriesLst = []
    if "openconfig-network-instance:ipv4-unicast" in response:
        ipEntries = response["openconfig-network-instance:ipv4-unicast"]
        if "ipv4-entry" in ipEntries:
            ipEntriesLst = ipEntries["ipv4-entry"]
    elif "openconfig-network-instance:ipv4-entry" in response:
        ipEntriesLst = response["openconfig-network-instance:ipv4-entry"]

    for ipEntry in ipEntriesLst:
        route = {}
        distance = ""
        originProtcol = ""
        if "prefix" in ipEntry:
            prefix = ipEntry["prefix"].encode('ascii', 'ignore')
            route.update({"prefix": prefix})
        if "openconfig-aft-ipv4-ext:metric" in ipEntry:
            distance = "/" + str(ipEntry["openconfig-aft-ipv4-ext:metric"])
        if "openconfig-aft-ipv4-ext:distance" in ipEntry:
            distance = str(ipEntry["openconfig-aft-ipv4-ext:distance"]) + distance
        if "openconfig-aft-ipv4-ext:origin-protocol" in ipEntry:
            originProtocol = getOriginProtocol(ipEntry["openconfig-aft-ipv4-ext:origin-protocol"])
            route.update({"originProtocol": originProtocol})
        if "openconfig-aft-ipv4-ext:uptime" in ipEntry:
            uptime = ipEntry["openconfig-aft-ipv4-ext:uptime"].encode('ascii', 'ignore')
            route.update({"uptime": uptime})
        if "openconfig-aft-ipv4-ext:next-hops" in ipEntry:
            nextHops = ipEntry["openconfig-aft-ipv4-ext:next-hops"]
            parse_nextHops(route, nextHops)
        if len(distance) > 0:
            route.update({"distance":distance})
        routeLst.append(route)

def parse_ipv6Entry(routeLst, response):

    ipEntriesLst = []
    if "openconfig-network-instance:ipv6-unicast" in response:
        ipEntries = response["openconfig-network-instance:ipv6-unicast"]
        if "ipv6-entry" in ipEntries:
            ipEntriesLst = ipEntries["ipv6-entry"]
    elif "openconfig-network-instance:ipv6-entry" in response:
        ipEntriesLst = response["openconfig-network-instance:ipv6-entry"]

    for ipEntry in ipEntriesLst:
        route = {}
        distance = ""
        originProtcol = ""
        if "prefix" in ipEntry:
            prefix = ipEntry["prefix"].encode('ascii', 'ignore')
            route.update({"prefix": prefix})
        if "openconfig-aft-ipv6-ext:metric" in ipEntry:
            distance = "/" + str(ipEntry["openconfig-aft-ipv6-ext:metric"])
        if "openconfig-aft-ipv6-ext:distance" in ipEntry:
            distance = str(ipEntry["openconfig-aft-ipv6-ext:distance"]) + distance
        if "openconfig-aft-ipv6-ext:origin-protocol" in ipEntry:
            originProtocol = getOriginProtocol(ipEntry["openconfig-aft-ipv6-ext:origin-protocol"])
            route.update({"originProtocol": originProtocol})
        if "openconfig-aft-ipv6-ext:uptime" in ipEntry:
            uptime = ipEntry["openconfig-aft-ipv6-ext:uptime"].encode('ascii', 'ignore')
            route.update({"uptime": uptime})
        if "openconfig-aft-ipv6-ext:next-hops" in ipEntry:
            nextHops = ipEntry["openconfig-aft-ipv6-ext:next-hops"]
            parse_nextHops(route, nextHops)
        if len(distance) > 0:
            route.update({"distance":distance})
        routeLst.append(route)

def run(func, args):
    try:
        api_response = invoke_api(func, args)
        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            else:
                routeLst = []
                if func == 'get_network_instance_afts_ipv4_unicast':
                    parse_ipv4Entry(routeLst, response)
                elif func == 'get_network_instance_afts_ipv6_unicast':
                    parse_ipv6Entry(routeLst, response)
                show_cli_output(args[0], routeLst)
        else:
            #error response
            print api_response.error_message()

    except:
       # system/network error
        print "%Error: Transaction Failure"

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

