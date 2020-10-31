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

[[ -z ${TOPDIR} ]] && TOPDIR=$(git rev-parse --show-toplevel) || true
[[ -z ${GO} ]] && GO=go || true

TESTARGS=()

while [[ $# -gt 0 ]]; do
case "$1" in
    -h|-help|--help)
        echo "usage: $(basename $0) [-bin] [-json] [-run NAME] [-auth] [ARGS...]"
        echo ""
        echo " -bin       Run test binary {TOPDIR}/build/tests/rest/server.test"
        echo " -json      Prints output in json format"
        echo " -run NAME  Run specific test cases."
        echo " -auth      Run auth tests; shorthand for '-run Auth -authtest local'"
        echo " ARGS...    Arguments to test program (log level, auth test arguments etc)"
        exit 0 ;;
    -bin)
        TESTBIN=${TOPDIR}/build/tests/rest/server.test
        shift ;;
    -json)
        JSON=1
        shift ;;
    -run)
        TESTCASE="$2"
        shift 2 ;;
    -auth)
        TESTCASE="Auth"
        TESTARGS+=("-authtest" "local")
        shift ;;
    *)
        TESTARGS+=("$1")
        shift;;
esac
done

MGMT_COMMON_DIR=$(realpath ${TOPDIR}/../sonic-mgmt-common)

export CVL_SCHEMA_PATH=${MGMT_COMMON_DIR}/build/cvl/schema

export DB_CONFIG_PATH=${MGMT_COMMON_DIR}/tools/test/database_config.json

PKG="github.com/Azure/sonic-mgmt-framework/rest/server"
DIR="${PWD}"

if [[ -z ${TESTBIN} ]]; then
    # run "go test" from source 
    CMD=( ${GO} test "${PKG}" -mod=vendor -v -cover -tags test )
    [[ -z ${JSON} ]] || CMD+=( -json )
    [[ -z ${TESTCASE} ]] || CMD+=( -run "${TESTCASE}" )
    CMD+=( -args ) #keep it last
else
    # run compiled test binary
    DIR="$(dirname ${TESTBIN})"
    CMD=( ./$(basename ${TESTBIN}) -test.v )
    [[ -z ${JSON} ]] || CMD=( ${GO} tool test2json -p "${PKG}" -t "${CMD[@]}" )
    [[ -z ${TESTCASE} ]] || CMD+=( -test.run "${TESTCASE}" )
fi

[[ "${TESTARGS[@]}" =~ -(also)?logtostderr ]] || TESTARGS+=( -logtostderr )

(cd "${DIR}" && "${CMD[@]}" "${TESTARGS[@]}")

