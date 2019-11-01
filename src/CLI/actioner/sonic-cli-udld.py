#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################

import sys
import json
import collections
import re
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


def invoke(func, args):
    body = None
    aa = cc.ApiClient()

    if func == 'udldGlobalShowHandler':
        return udldShowHandler("global_show", args)

    if func == 'udldNeighborShowHandler':
        return udldShowHandler("neighbor_show", args)

    if func == 'udldInterfaceShowHandler':
        return udldShowHandler("interface_show", args)

    if func == 'udldInterfaceCountersShowHandler':
        return udldShowHandler("counters_show", args)

    # Configure UDLD global
    if func == 'udldGlobalEnableHandler' :
        return udldConfigHandler("global_enable", args)

    # Configure UDLD normal
    if func == 'udldGlobalNormalEnableHandler' :
        return udldConfigHandler("global_normal", args)

    # Configure UDLD message-time
    if func == 'udldMsgTimeHandler' :
        return udldConfigHandler("global_msg_time", args)

    # Configure UDLD multiplier
    if func == 'udldMultiplierHandler' :
        return udldConfigHandler("global_multiplier", args)

    # Configure UDLD enable/disable at Interface
    if func == 'udldInterfaceEnableHandler' :
        return udldConfigHandler("interface_enable", args)

    # Configure UDLD normal at Interface
    if func == 'udldInterfaceNormalEnableHandler' :
        return udldConfigHandler("interface_normal", args)


def udldConfigHandler(option, args):
    if option == "global_enable":
        if args[0] == '1':
            print("Enabled UDLD globally")
        else:    
            print("Disabled UDLD globally")
    elif option == "global_normal":
        if args[0] == '1':
            print("Enabled UDLD globally in Normal mode")
        else:    
            print("Disabled UDLD globally in Normal mode")
    elif option == "global_msg_time":
        print("Set UDLD Message Time to: " + args[0])
    elif option == "global_multiplier":
        print("Set UDLD Multiplier to: " + args[0])
    elif option == "interface_enable":
        if args[1] == '1':
            print("Enabled UDLD on interface: " + args[0])
        else:    
            print("Disabled UDLD on interface: " + args[0])
    elif option == "interface_normal":
        if args[1] == '1':
            print("Enabled UDLD in Normal mode on interface: " + args[0])
        else:    
            print("Disabled UDLD in Normal mode on interface: " + args[0])

    return ""


def udldShowHandler(option, args):
    if option == "global_show":
        print("UDLD GLobal Information")
        print("Admin State:         UDLD Enabled")
        print("Mode:                Aggresive")
        print("UDLD Message time:   1 secs")
        print("UDLD Multiplier:     3")
    elif option == "neighbor_show":
        print("Port           Device Name     Device ID         Port ID         Neighbor State")
        print("---------------------------------------------------------------------------------")
        print("Ethernet0      Sonic           3c2c.992d.8201    Ethernet8       Bidirectional")
        print("Ethernet4      Sonic           3c2c.992d.8205    Ethernet12      Bidirectional")
    elif option == "interface_show":
        print("UDLD information for " + args[0])
        print("  UDLD Admin State:                  Enabled")
        print("  Mode:                              Aggressive")
        print("  Status:                            Bidirectional")
        print("  Local device id:                   3c2c.992d.8201")
        print("  Local port id :                    " + args[0])
        print("  Local device name:                 Sonic")
        print("  Message Time:                      1 secs")
        print("  Timeout Interval:                  3")
        print("     Neighbor Entry 1")
        print("     ----------------------------------------------------------------------------------------")
        print("     Neighbor device id:         3c2c.992d.8235")
        print("     Neighbor port id:           Ethernet8") 
        print("     Neighbor device name:       Sonic") 
        print("     Neighbor message time:      1")
        print("     Neighbor timeout interval:  3")
    elif option == "counters_show":
        if len(args) > 0:
            print("UDLD Interface statistics for " + args[0])
            print("Frames transmitted:         10")
            print("Frames received:            9")
            print("Frames with error:          0")
        else:    
            print("UDLD Interface statistics for Ethernet0")
            print("Frames transmitted:         120")
            print("Frames received:            39")
            print("Frames with error:          0")
            print("UDLD Interface statistics for Ethernet4")
            print("Frames transmitted:         20")
            print("Frames received:            23")
            print("Frames with error:          0")
            print("UDLD Interface statistics for Ethernet8")
            print("Frames transmitted:         68")
            print("Frames received:            53")
            print("Frames with error:          3")

    return ""


def run(func, args):
        api_response = invoke(func, args)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

