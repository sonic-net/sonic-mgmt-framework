#!/usr/bin/env bash

################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

set -e

ARGS=()
HOST=localhost
PORT=443

while [[ $# -gt 0 ]]; do
case "$1" in
    -host) HOST=$2; shift 2 ;;
    -port) PORT=$2; shift 2 ;;
    *) ARGS+=("$1"); shift ;;
esac
done

TOPDIR=$PWD
BUILDDIR=$TOPDIR/build

[ -z $SYSTEM_NAME ] && export SYSTEM_NAME=$HOSTNAME

export REST_API_ROOT=https://$HOST:$PORT

export SONIC_CLI_ROOT=$BUILDDIR/cli

export CLISH_PATH=$SONIC_CLI_ROOT/command-tree

export PYTHONVER=2.7.14
export PYTHONPATH=$PYTHONPATH:$BUILDDIR/swagger_client_py:$SONIC_CLI_ROOT:$SONIC_CLI_ROOT/scripts

# KLISH_BIN can be set to use klish exe and libs from other directory
[ ! -d "$KLISH_BIN" ] && KLISH_BIN=$SONIC_CLI_ROOT

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$KLISH_BIN/.libs

export PATH=$PATH:$KLISH_BIN

(cd $SONIC_CLI_ROOT && clish ${ARGS[@]})

