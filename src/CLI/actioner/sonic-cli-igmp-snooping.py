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
    
    # Get the rules of all ACL table entries.
    if func == 'get_igmp_snooping_interfaces_interface_state':
        if len(args) == 4 and args[2].lower() == 'vlan':
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface=Vlan{vlanid}',
                vlanid=args[3])
            return aa.get(keypath)
        elif len(args) == 3 and args[2].lower() == 'groups':
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces')            
            return aa.get(keypath)
        elif len(args) == 5 and args[2].lower() == 'groups':
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface=Vlan{vlanid}',
                vlanid=args[4])
            return aa.get(keypath)
        else: 
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces')
            return aa.get(keypath)            

    if func == 'patch_igmp_snooping_interfaces_interface_config' :            
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}',
                vlanid=args[0])
        
        body=collections.defaultdict(dict)
        
        if len(args) == 2 :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'querier' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "querier": True,
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'fast-leave' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "fast-leave": True,
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'version' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "version": int(args[3]),
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'query-interval' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "query-interval": int(args[3]),
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'last-member-query-interval' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "last-member-query-interval": int(args[3]),
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'query-max-response-time' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "query-max-response-time": int(args[3]),
                        "enabled": True
                        }
                      }
                    ]
                   }
            
        elif args[2] == 'mrouter' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "mrouter-interface": [(args[4])],
                        "enabled": True
                        }
                      }
                    ]
                   }
                        
        elif args[2] == 'static-group' :
            body = { "interface": [                    
                      { 
                        "name": args[0],   
                        "config" : {
                        "static-multicast-group": [ { "group": args[3], "outgoing-interface": [args[5]] } ],
                        "enabled": True
                        }
                      }
                    ]
                   }
        else:    
            print("%Error: Invalid command")
            exit(1)
            
        return aa.patch(keypath, body)
    elif func == 'delete_igmp_snooping_interfaces_interface_config' :
        keypath = None
        
        if len(args) == 2 :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/enabled',
                vlanid=args[0])
        elif args[2] == 'querier' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/querier',
                vlanid=args[0])            
        elif args[2] == 'fast-leave' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/fast-leave',
                vlanid=args[0])
        elif args[2] == 'version' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/version',
                vlanid=args[0])
        elif args[2] == 'query-interval' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/query-interval',
                vlanid=args[0])
        elif args[2] == 'last-member-query-interval' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/last-member-query-interval',
                vlanid=args[0])
        elif args[2] == 'query-max-response-time' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/query-max-response-time',
                vlanid=args[0])
        elif args[2] == 'mrouter' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/mrouter-interface={ifname}',
                vlanid=args[0], ifname=args[4])
        elif args[2] == 'static-group' :
            keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/static-multicast-group={grpAddr}/outgoing-interface={ifname}',
                vlanid=args[0], grpAddr=args[3], ifname=args[5])
            api_response = aa.delete (keypath)
            if api_response.ok():
                keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/static-multicast-group={grpAddr}/outgoing-interface',
                                  vlanid=args[0], grpAddr=args[3])
                get_response = aa.get(keypath)
                if get_response.ok() and len(get_response.content) == 0:
                    keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance=default/protocols/protocol=IGMP_SNOOPING,IGMP-SNOOPING/openconfig-network-instance-deviation:igmp-snooping/interfaces/interface={vlanid}/config/static-multicast-group={grpAddr}',
                        vlanid=args[0], grpAddr=args[3])
                    return aa.delete (keypath)
                else:
                    return api_response
            else:
                return api_response
        else:    
            print("%Error: Invalid command")
            exit(1)
            
        return aa.delete (keypath)
    else:
        print("%Error: not implemented")
        exit(1)

def run(func, args):
    try:
        api_response = invoke(func, args)
        if api_response.ok():
            response = api_response.content
            if response is None:
                print "Success"
            elif (len(args) == 2 and args[1].lower() == 'snooping') or (len(args) == 4 and args[2].lower() == 'vlan'):   
                if 'openconfig-network-instance-deviation:interfaces' in response.keys():
                    value = response['openconfig-network-instance-deviation:interfaces']
                    if value is None:
                        return
                    show_cli_output(args[0], value)                
                elif 'openconfig-network-instance-deviation:interface' in response.keys():
                    show_cli_output(args[0], response)
            elif len(args) >= 3 and args[2].lower() == 'groups':
                if 'openconfig-network-instance-deviation:interfaces' in response.keys():
                    value = response['openconfig-network-instance-deviation:interfaces']
                    if value is None:
                        return
                    show_cli_output('show_igmp_snooping-groups.j2', value)                
                elif 'openconfig-network-instance-deviation:interface' in response.keys():
                    show_cli_output('show_igmp_snooping-groups.j2', response)
            else:
                print "Error"                
        else:
            #error response
            print api_response.error_message()

    except:
            # system/network error
            print "%Error: Transaction Failure"


if __name__ == '__main__':
    pipestr().write(sys.argv)
    #pdb.set_trace()
    run(sys.argv[1], sys.argv[2:])

