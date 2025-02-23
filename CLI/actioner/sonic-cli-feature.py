#!/usr/bin/python
###########################################################################
#
# Copyright (C) 2023 NIPPON TELEGRAPH AND TELEPHONE CORPORATION.
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
import cli_client as cc
import re


def config_nhg_feature(enable):
    body = None
    aa = cc.ApiClient()

    if enable:
        keypath = cc.Path("/restconf/data/sonic-device_metadata:sonic-device_metadata")
        body = {
            "sonic-device_metadata": {
                "DEVICE_METADATA": {"localhost": {"nexthop_group": "enabled"}}
            }
        }
        return aa.patch(keypath, body)
    else:
        keypath = cc.Path(
            "/restconf/data/sonic-device_metadata:sonic-device_metadata/DEVICE_METADATA/localhost/nexthop_group"
        )
        return aa.delete(keypath)


def invoke(func, enable):
    if func == "configure_sonic_nexthop_groups":
        return config_nhg_feature(enable == "1")


def run(func, enable):
    if func != "configure_sonic_nexthop_groups":
        print("%Error: Invalid function")
        return

    try:
        api_response = invoke(func, enable)
    except ValueError as err_msg:
        print(
            "%Error: An exception occurred while attempting to execute "
            "the requested RPC call: {}".format(err_msg)
        )

    if not api_response.ok():
        # Print the message for a failing return code
        print("CLI transformer error: ")
        print("    status code: {}".format(api_response.status_code))
        if (
            "error" in api_response.errors()
            and len(api_response.errors()["error"]) >= 1
            and "error-message" in api_response.errors()["error"][0]
        ):
            print(
                "    error: {}".format(
                    api_response.errors()["error"][0]["error-message"]
                )
            )
        return

    print("Success")


if __name__ == "__main__":
    run(sys.argv[1], sys.argv[2])
