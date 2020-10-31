#!/usr/bin/env bash
################################################################################
#                                                                              #
#  Copyright 2020 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
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
    -h|-help|--help)
        echo "usage: $(basename $0) [-host HOST] [-port PORT] [KLISH_OPTIONS]"
        ARGS+=("-h")
        shift ;;
    -host) HOST=$2; shift 2 ;;
    -port) PORT=$2; shift 2 ;;
    -debug) DEBUG=1; shift ;;
    *) ARGS+=("$1"); shift ;;
esac
done

TOPDIR=$(git rev-parse --show-toplevel)
BUILDDIR=${TOPDIR}/build

CLISOURCE=${TOPDIR}/CLI
CLIBUILD=${BUILDDIR}/cli

[[ -z ${AUTH} ]] && export CLISH_NOAUTH=1
[[ -z ${DEBUG} ]] || export LOGTOSCREEN=1

# Prompt
[[ -z ${SYSTEM_NAME} ]] && export SYSTEM_NAME=sonic-cli

export REST_API_ROOT=https://${HOST}:${PORT}

export SONIC_CLI_ROOT=${CLISOURCE}/actioner

export RENDERER_TEMPLATE_PATH=${CLISOURCE}/renderer/templates

export SHOW_CONFIG_TOOLS=${CLIBUILD}/render-templates

export CLISH_PATH=${CLIBUILD}/command-tree

# KLISH_BIN can be set to use klish exe and libs from other directory.
if [[ -z ${KLISH_BIN} ]]; then
    if [[ -f ${CLIBUILD}/clish ]]; then
        KLISH_BIN=${CLIBUILD}
    elif [[ -f ${BUILDDIR}/target/clish ]]; then
        KLISH_BIN=${BUILDDIR}/target
    else
        echo "Error: could not locate clish."
        exit 1
    fi
fi

PYTHONPATH+=:${CLISOURCE}/actioner
PYTHONPATH+=:${CLISOURCE}/renderer
PYTHONPATH+=:${CLISOURCE}/renderer/scripts
PYTHONPATH+=:${BUILDDIR}/swagger_client_py
PYTHONPATH+=:$(realpath ${TOPDIR}/..)/sonic-py-swsssdk/src
export PYTHONPATH

export LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:${KLISH_BIN}/.libs


(cd ${BUILDDIR} && ${KLISH_BIN}/clish "${ARGS[@]}")
