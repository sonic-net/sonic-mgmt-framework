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


def invoke(func, enable):
    body = None
    aa = cc.ApiClient()

    feature = "nexthop_group"

    if enable == "1":

        keypath = cc.Path("/restconf/data/sonic-feature:sonic-feature")

        body = {
            "sonic-feature": {
                "FEATURE": {"FEATURE_LIST": [{"name": feature, "state": "enabled"}]}
            }
        }
        return aa.patch(keypath, body)

    else:

        keypath = cc.Path(
            "/restconf/data/sonic-feature:sonic-feature/FEATURE/FEATURE_LIST={}".format(
                feature
            )
        )

        return aa.delete(keypath)


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
