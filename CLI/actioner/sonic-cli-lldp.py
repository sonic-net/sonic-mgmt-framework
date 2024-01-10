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


def show_lldp_interface(path, template):
    resp = ApiClient().get(path, ignore404=False)
    if not resp.ok():
        print(resp.error_message())
        return 1
    if "openconfig-lldp:interfaces" in resp.content:
        data = resp.content["openconfig-lldp:interfaces"].get("interface", [])
    else:
        data = resp.content.get("openconfig-lldp:interface", [])
    # Server returns junk value for not found case!! Identify valid resp here
    neigh = data[0].get("neighbors", {}).get("neighbor", []) if data else None
    if neigh and "state" in neigh[0]:
        sorted_data = natsorted(data, key=lambda x: x["name"])
        show_cli_output(template, sorted_data)
    return 0


class Handlers:
    @staticmethod
    def get_openconfig_lldp_lldp_interfaces(template, *args):
        allif_path = Path("/restconf/data/openconfig-lldp:lldp/interfaces")
        return show_lldp_interface(allif_path, template)

    @staticmethod
    def get_openconfig_lldp_lldp_interfaces_interface(template, ifname, *args):
        oneif_path = Path("/restconf/data/openconfig-lldp:lldp/interfaces/interface={name}", name=ifname)
        return show_lldp_interface(oneif_path, template)


def run(func, args):
    return getattr(Handlers, func)(*args)


if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = sys.argv[1]
    run(func, sys.argv[2:])
