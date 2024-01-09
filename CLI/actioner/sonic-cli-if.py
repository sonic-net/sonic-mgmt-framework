#!/usr/bin/python3
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
from natsort import natsorted

from cli_client import ApiClient, Path
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


def config_intf(name, sub_path, value):
    path = Path("/restconf/data/openconfig-interfaces:interfaces/interface={name}", name=name).join(sub_path)
    if value is None:
        resp = ApiClient().delete(path, ignore404=True)
    else:
        resp = ApiClient().patch(path, value)
    if not resp.ok():
        print(resp.error_message())
        return 1
    return 0


def ipaddr_payload(ip):
    addr, mask = ip.split("/")
    return {"address": [{
        "ip": addr,
        "config": {
            "ip": addr,
            "prefix-length": int(mask),
        },
    }]}


class Handlers:
    @staticmethod
    def patch_openconfig_interfaces_interfaces_interface_config_description(name, value, *args):
        return config_intf(name, "config/description", {"description": value})

    @staticmethod
    def patch_openconfig_interfaces_interfaces_interface_config_enabled(name, value, *args):
        return config_intf(name, "config/enabled", {"enabled": (value == "True")})

    @staticmethod
    def patch_openconfig_interfaces_interfaces_interface_config_mtu(name, mtu, *args):
        return config_intf(name, "config/mtu", {"mtu": int(mtu)})

    @staticmethod
    def patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config(name, ip4, *args):
        ip4_path = "subinterfaces/subinterface=0/openconfig-if-ip:ipv4/addresses/address"
        return config_intf(name, ip4_path, ipaddr_payload(ip4))

    @staticmethod
    def delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_addresses_address_config_prefix_length(name, ip4):
        ip4_path = Path("subinterfaces/subinterface=0/openconfig-if-ip:ipv4/addresses/address={ip}", ip=ip4)
        return config_intf(name, ip4_path, None)

    @staticmethod
    def patch_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config(name, ip6, *args):
        ip6_path = "subinterfaces/subinterface=0/openconfig-if-ip:ipv6/addresses/address"
        return config_intf(name, ip6_path, ipaddr_payload(ip6))

    @staticmethod
    def delete_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_addresses_address_config_prefix_length(name, ip6):
        ip6_path = Path("subinterfaces/subinterface=0/openconfig-if-ip:ipv6/addresses/address={ip}", ip=ip6)
        return config_intf(name, ip6_path, None)

    @staticmethod
    def get_openconfig_interfaces_interfaces(template, *args):
        resp = ApiClient().get("/restconf/data/openconfig-interfaces:interfaces")
        if not resp.ok():
            print(resp.error_message())
            return 1
        if not resp.content:
            return 0
        # Sort interface records by name
        intf_list = resp.content.get("openconfig-interfaces:interfaces", {}).get("interface", [])
        if len(intf_list) > 1:
            sorted_list = natsorted(intf_list, key=lambda x: x["name"])
            resp.content["openconfig-interfaces:interfaces"]["interface"] = sorted_list
        show_cli_output(template, resp.content)
        return 0

    @staticmethod
    def get_openconfig_interfaces_interfaces_interface(name, template, *args):
        path = Path("/restconf/data/openconfig-interfaces:interfaces/interface={name}", name=name)
        resp = ApiClient().get(path, ignore404=False)
        if not resp.ok():
            print(resp.error_message())
            return 1
        if resp.content:
            show_cli_output(template, resp.content)
        return 0


def run(func, args):
    return getattr(Handlers, func)(*args)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])
