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
import re
from cli_client import ApiClient, Path
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output


_proto_number_map = {
    "icmp": "IP_ICMP",
    "tcp": "IP_TCP",
    "udp": "IP_UDP",
    "6": "IP_TCP",
    "17": "IP_UDP",
    "1": "IP_ICMP",
    "2": "IP_IGMP",
    "103": "IP_PIM",
    "46": "IP_RSVP",
    "47": "IP_GRE",
    "51": "IP_AUTH",
    "115": "IP_L2TP",
}

_action_map = {
    "permit": "ACCEPT",
    "deny": "DROP",
}


def acl_path(a_name=None, a_type=None):
    p = Path("/restconf/data/openconfig-acl:acl/acl-sets")
    if a_name is None or a_type is None:
        return p.join("acl-set")
    return p.join("acl-set={name},{type}", name=a_name, type=a_type)


def acl_if_path(if_name=None):
    p = Path("/restconf/data/openconfig-acl:acl/interfaces")
    if if_name is None:
        return p.join("interface")
    return p.join("interface={id}", id=if_name)


def check_ok(resp):
    if not resp.ok():
        print(resp.error_message())
        return 1
    return 0


def render(path, template):
    resp = ApiClient().get(path, ignore404=True)
    if not resp.ok():
        print(resp.error_message())
        return 1
    if resp.content:
        show_cli_output(template, resp.content)
    return 0


class Handlers:
    @staticmethod
    def patch_list_openconfig_acl_acl_acl_sets_acl_set(a_name, a_type):
        keys = {"name": a_name, "type": a_type}
        body = {"acl-set": [{**keys, "config": {**keys}}]}
        resp = ApiClient().patch(acl_path(), body)
        return check_ok(resp)

    @staticmethod
    def delete_openconfig_acl_acl_acl_sets_acl_set(a_name, a_type):
        resp = ApiClient().delete(acl_path(a_name, a_type))
        return check_ok(resp)

    @staticmethod
    def patch_list_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry(a_name, a_type, seq, action, proto, *args):
        if proto not in _proto_number_map:
            print("%Error: Invalid protocol number")
            return 1

        action_config = {"forwarding-action": _action_map[action]}
        ipv4_config = {"protocol": _proto_number_map[proto]}
        xport_config = {}
        xport_flags = []

        re_ip = re.compile("^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}")
        addr_key = iter(["source-address", "destination-address"])
        arg_list = list(args)

        while arg_list:
            arg = arg_list.pop(0)
            if re_ip.match(arg):
                ipv4_config[next(addr_key)] = arg
            elif arg == "any":
                ipv4_config[next(addr_key)] = "0.0.0.0/0"
            elif arg == "src-eq":
                xport_config["source-port"] = int(arg_list.pop(0))
            elif arg == "dst-eq":
                xport_config["destination-port"] = int(arg_list.pop(0))
            elif arg == "dscp":
                ipv4_config["dscp"] = int(arg_list.pop(0))
            elif arg in ['fin', 'syn', 'ack', 'urg', 'rst', 'psh']:
                xport_flags.append("TCP_"+arg.upper())

        if xport_flags:
            xport_config["tcp-flags"] = xport_flags

        path = acl_path(a_name, a_type).join("acl-entries/acl-entry")
        body = {"acl-entry": [{
            "sequence-id": int(seq),
            "config":      {"sequence-id": int(seq)},
            "ipv4":        {"config": ipv4_config},
            "transport":   {"config": xport_config},
            "actions":     {"config": action_config},
        }]}

        resp = ApiClient().patch(path, body)
        return check_ok(resp)

    @staticmethod
    def delete_openconfig_acl_acl_acl_sets_acl_set_acl_entries_acl_entry(a_name, a_type, seq):
        path = acl_path(a_name, a_type).join("acl-entries/acl-entry={seq}", seq=seq)
        resp = ApiClient().delete(path)
        return check_ok(resp)

    @staticmethod
    def patch_list_openconfig_acl_acl_interfaces_interface(a_name, a_type, if_name, direction):
        a_ref = {"set-name": a_name, "type": a_type}
        a_ref = {**a_ref, "config": {**a_ref}}
        binding = {
            "id": if_name,
            "config": {"id": if_name},
        }
        if direction == "ingress":
            binding["ingress-acl-sets"] = {"ingress-acl-set": [a_ref]}
        else:
            binding["egress-acl-sets"] = {"egress-acl-set": [a_ref]}

        resp = ApiClient().patch(acl_if_path(), {"interface": [binding]})
        return check_ok(resp)

    @staticmethod
    def delete_openconfig_acl_acl_interfaces_interface_ingress_acl_sets_ingress_acl_set(if_name, a_name, a_type):
        path = acl_if_path(if_name).join("ingress-acl-sets/ingress-acl-set={name},{type}", name=a_name, type=a_type)
        resp = ApiClient().delete(path)
        return check_ok(resp)

    @staticmethod
    def delete_openconfig_acl_acl_interfaces_interface_egress_acl_sets_egress_acl_set(if_name, a_name, a_type):
        path = acl_if_path(if_name).join("egress-acl-sets/egress-acl-set={name},{type}", name=a_name, type=a_type)
        resp = ApiClient().delete(path)
        return check_ok(resp)

    @staticmethod
    def get_openconfig_acl_acl_acl_sets(template, *args):
        return render(acl_path(), template)

    @staticmethod
    def get_openconfig_acl_acl_acl_sets_acl_set_acl_entries(a_name, a_type, template, *args):
        return render(acl_path(a_name, a_type), template)

    @staticmethod
    def get_openconfig_acl_acl_interfaces(template, *args):
        return render(acl_if_path(), template)


def run(func, args):
    return getattr(Handlers, func)(*args)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])
