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

TOPDIR=$PWD

TEST_ARGS=()
REST_ARGS=()
PIPE=tee

while [[ $# -gt 0 ]]; do
case "$1" in
    -h|-help|--help)
        echo "usage: $0 [-run NAME] [-auth] [-json] [-tparse] [ARGS...]"
        echo ""
        echo " -run NAME  Run specific test cases."
        echo "            Value starting with 'Auth' implicitly enables -auth"
        echo " -auth      Enable local password authentication test cases"
        echo "            Extra arguments may be needed as described in pamAuth_test.go"
        echo " -json      Prints output in json format"
        echo " -tparse    Render output through tparse; implicitly enables -json"
        echo " ARGS...    Arguments to test program (log level, auth test arguments etc)"
        exit 0 ;;
    -auth)
        REST_ARGS+=("-authtest" "local")
        shift ;;
    -run)
        TEST_ARGS+=("-run" "$2")
        [[ $2 == Auth* ]] && REST_ARGS+=("-authtest" "local")
        shift 2 ;;
    -json)
        TEST_ARGS+=("-json")
        shift ;;
    -tparse)
        TEST_ARGS+=("-json")
        PIPE=tparse
        shift ;;
    *)
        REST_ARGS+=("$1")
        shift;;
esac
done

export GOPATH=$TOPDIR:$TOPDIR/build/gopkgs:$TOPDIR/build/rest_server/dist

export CVL_SCHEMA_PATH=$TOPDIR/build/cvl/schema

export YANG_MODELS_PATH=$TOPDIR/build/all_yangs

go test rest/server -v -cover "${TEST_ARGS[@]}" -args -logtostderr "${REST_ARGS[@]}" | $PIPE

