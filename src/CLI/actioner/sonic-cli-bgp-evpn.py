#!/usr/bin/python
###########################################################################
#
# Copyright 2019 Broadcom, Inc.
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
import time
import json
import ast
from rpipe_utils import pipestr
import cli_client as cc
from scripts.render_cli import show_cli_output


def invoke_api(func, args=[]):
    api = cc.ApiClient()
    keypath = []
    body = None
    
    #Patch cases
    if func == 'patch_bgp_evpn_advertise_all_vni':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:advertise-all-vni',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:advertise-all-vni": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_advertise_default_gw':
        #TODO: Change to Openconfig API
        keypath = cc.Path('/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS_AF/'
            +'BGP_GLOBALS_AF_LIST={vrf},{af_name}/advertise-default-gw',
                vrf=args[0], af_name=args[1])
        body = { "sonic-bgp-global:advertise-default-gw": True if args[2] == 'True' else False }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_default_originate':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:default-originate/{originate_family}',
                vrf=args[0], af_name=args[1], originate_family=args[2])
        body = { "openconfig-bgp-evpn-ext:{}".format(args[2]): True }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_rd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:route-distinguisher',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:route-distinguisher": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_rt':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vpn-target={route_target}',
                vrf=args[0], af_name=args[1], route_target=args[2])
        body = { "openconfig-bgp-evpn-ext:vpn-target": [ 
                    { "route-target": args[2], "route-target-type": args[3] } ] 
                }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_advertise':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:advertise-list',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:advertise-list": [ args[2] ] }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_autort':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:autort',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:autort": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_flooding':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:flooding',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:flooding": args[2] }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_dad_enable':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/enabled',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:enabled": True }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_dad_params':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection',
                vrf=args[0], af_name=args[1])
        body = { 
                    "openconfig-bgp-evpn-ext:dup-addr-detection": {
                        "enabled": True,
                        "max-moves": int(args[2]),
                        "time": int(args[3])
                    }
                }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_dad_freeze':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/freeze',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:freeze": args[2] }
        return api.patch(keypath, body)

    #Patch EVPN VNI cases
    elif func == 'patch_bgp_evpn_vni':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni',
                vrf=args[0], af_name=args[1])
        body = { "openconfig-bgp-evpn-ext:vni": [ 
                    { "vni-number": int(args[2]), 
                        "config" : 
                            { "vni-number" : int(args[2]) } 
                    } ] 
                }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_vni_rt':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/vpn-target={route_target}',
                vrf=args[0], af_name=args[1], vni_number=args[2], route_target=args[3])
        body = { "openconfig-bgp-evpn-ext:vpn-target": [ 
                    { "route-target": args[3], "route-target-type": args[4] } ] 
                }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_vni_advertise_default_gw':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/advertise-default-gw',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        body = { "openconfig-bgp-evpn-ext:advertise-default-gw": True }
        return api.patch(keypath, body)
    elif func == 'patch_bgp_evpn_vni_rd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/route-distinguisher',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        body = { "openconfig-bgp-evpn-ext:route-distinguisher": args[3] }
        return api.patch(keypath, body)


    #Delete cases
    elif func == 'delete_bgp_evpn_advertise_all_vni':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:advertise-all-vni',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_advertise_default_gw':
        #TODO: Change to Openconfig API
        keypath = cc.Path('/restconf/data/sonic-bgp-global:sonic-bgp-global/BGP_GLOBALS_AF/'
            +'BGP_GLOBALS_AF_LIST={vrf},{af_name}/advertise-default-gw',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_default_originate':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:default-originate/{originate_family}',
                vrf=args[0], af_name=args[1], originate_family=args[2])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_rd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:route-distinguisher',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_rt':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vpn-target={route_target}',
                vrf=args[0], af_name=args[1], route_target=args[2])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_advertise':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:advertise-list={afi_safi_name}',
                vrf=args[0], af_name=args[1], afi_safi_name=args[2])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_autort':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:autort',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_flooding':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:flooding',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_dad_enable':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/enabled',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_dad_params':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/max-moves',
                vrf=args[0], af_name=args[1])
        api.delete(keypath)
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/time',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_dad_freeze':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:dup-addr-detection/freeze',
                vrf=args[0], af_name=args[1])
        return api.delete(keypath)

    #Delete EVPN VNI cases
    elif func == 'delete_bgp_evpn_vni':
        print('delete evpn vni '+args[0]+args[1]+args[2])
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_vni_rt':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/vpn-target={route_target}',
                vrf=args[0], af_name=args[1], vni_number=args[2], route_target=args[3])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_vni_advertise_default_gw':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/advertise-default-gw',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        return api.delete(keypath)
    elif func == 'delete_bgp_evpn_vni_rd':
        keypath = cc.Path('/restconf/data/openconfig-network-instance:network-instances/network-instance={vrf}'
            +'/protocols/protocol=BGP,bgp/bgp/global/afi-safis/afi-safi={af_name}/l2vpn-evpn'
            +'/openconfig-bgp-evpn-ext:vnis/vni={vni_number}/route-distinguisher',
                vrf=args[0], af_name=args[1], vni_number=args[2])
        return api.delete(keypath)

    else:
        body = {}

    return api.cli_not_implemented(func)

def run(func, args):
    response = invoke_api(func, args)

    if response.ok():
        if response.content is not None:
            print("Failed")
    else:
        print(response.error_message())

if __name__ == '__main__':

    pipestr().write(sys.argv)
    func = sys.argv[1]

    run(func, sys.argv[2:])

