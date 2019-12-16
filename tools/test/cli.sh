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

CLISOURCE=$TOPDIR/src/CLI
CLIBUILD=$BUILDDIR/cli

[ -z $SYSTEM_NAME ] && export SYSTEM_NAME=sonic-cli

export REST_API_ROOT=https://$HOST:$PORT

export SONIC_CLI_ROOT=$CLISOURCE/actioner

export RENDERER_TEMPLATE_PATH=$CLISOURCE/renderer/templates

#export CLISH_PATH=$CLISOURCE/clitree/cli-xml
export CLISH_PATH=$CLIBUILD/command-tree

export PYTHONVER=2.7.14

PYTHONPATH+=:$CLISOURCE/actioner
PYTHONPATH+=:$CLISOURCE/renderer
PYTHONPATH+=:$CLISOURCE/renderer/scripts
PYTHONPATH+=:$BUILDDIR/swagger_client_py
PYTHONPATH+=:$(realpath $TOPDIR/..)/sonic-py-swsssdk/src
export PYTHONPATH

# KLISH_BIN can be set to use klish exe and libs from other directory
[ ! -d "$KLISH_BIN" ] && KLISH_BIN=$CLIBUILD

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$KLISH_BIN/.libs

export PATH=$PATH:$KLISH_BIN

(cd $BUILDDIR && clish ${ARGS[@]})

