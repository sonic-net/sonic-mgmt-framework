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

# Setup database config file path
if [[ -z ${DB_CONFIG_PATH} ]]; then
    export DB_CONFIG_PATH=${MGMT_COMMON_DIR}/tools/test/database_config.json
fi

# LD_LIBRARY_PATH for CVL
[ -z $LD_LIBRARY_PATH ] && export LD_LIBRARY_PATH=/usr/local/lib

# Setup CVL schema directory
if [ -z $CVL_SCHEMA_PATH ]; then
    export CVL_SCHEMA_PATH=$MGMT_COMMON_DIR/build/cvl/schema
fi

echo "CVL schema directory is $CVL_SCHEMA_PATH"
if [ $(find $CVL_SCHEMA_PATH -name *.yin | wc -l) == 0 ]; then
    echo "WARNING: no yin files at $CVL_SCHEMA_PATH"
    echo ""
fi

# Prepare CVL config file with all traces enabled
if [[ -z $CVL_CFG_FILE ]]; then
    export CVL_CFG_FILE=$SERVER_DIR/cvl_cfg.json
    if [[ ! -e $CVL_CFG_FILE ]]; then
        F=$MGMT_COMMON_DIR/cvl/conf/cvl_cfg.json
        sed -E 's/((TRACE|LOG).*)\"false\"/\1\"true\"/' $F > $CVL_CFG_FILE
    fi
fi

# Prepare yang files directiry for transformer
if [[ -z ${YANG_MODELS_PATH} ]]; then
    export YANG_MODELS_PATH=${BUILD_DIR}/all_yangs
    mkdir -p ${YANG_MODELS_PATH}
    pushd ${YANG_MODELS_PATH} > /dev/null
    MGMT_COMN=$(realpath --relative-to=${PWD} ${MGMT_COMMON_DIR})
    rm -f *
    find ${MGMT_COMN}/models/yang -name "*.yang" -not -path "*/testdata/*" -exec ln -sf {} \;
    ln -sf ${MGMT_COMN}/models/yang/version.xml
    ln -sf ${MGMT_COMN}/config/transformer/models_list
    popd > /dev/null
fi

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

