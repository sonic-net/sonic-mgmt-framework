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

TOPDIR=$(git rev-parse --show-toplevel || echo ${PWD})
BUILD_DIR=$TOPDIR/build
SERVER_DIR=$TOPDIR/build/rest_server
MGMT_COMMON_DIR=$(realpath $TOPDIR/../sonic-mgmt-common)

if [[ ! -f $SERVER_DIR/rest_server ]]; then
    echo "error: REST server not compiled"
    echo "Please run 'make rest-server' and try again"
    exit 1
fi

source ${MGMT_COMMON_DIR}/tools/test/env.sh --dest=${SERVER_DIR}

EXTRA_ARGS="-rest_ui $SERVER_DIR/dist/ui -logtostderr"

for V in $@; do
    case $V in
    -cert|--cert|-cert=*|--cert=*) HAS_CRTFILE=1;;
    -key|--key|-key=*|--key=*) HAS_KEYFILE=1;;
    -v|--v|-v=*|--v=*) HAS_V=1;;
    -client_auth|--client_auth) HAS_AUTH=1;;
    -client_auth=*|--client_auth=*) HAS_AUTH=1;;
    esac
done

[[ -z ${HAS_V} ]] && EXTRA_ARGS+=" -v 1"
[[ -z ${HAS_AUTH} ]] && EXTRA_ARGS+=" -client_auth none"

cd $SERVER_DIR

##
# Setup TLS server cert/key pair
if [ -z $HAS_CRTFILE ] && [ -z $HAS_KEYFILE ]; then
    if [ -f cert.pem ] && [ -f key.pem ]; then
        echo "Reusing existing cert.pem and key.pem ..."
    else 
        echo "Generating temporary server certificate ..."
        ./generate_cert --host localhost
    fi

    EXTRA_ARGS+=" -cert cert.pem -key key.pem"
fi

##
# Start server
./rest_server $EXTRA_ARGS $* 

