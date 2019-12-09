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
BUILD_DIR=$TOPDIR/build
SERVER_DIR=$BUILD_DIR/rest_server

# LD_LIBRARY_PATH for CVL
[ -z $LD_LIBRARY_PATH ] && export LD_LIBRARY_PATH=/usr/local/lib

# Setup CVL schema directory
if [ -z $CVL_SCHEMA_PATH ]; then
    export CVL_SCHEMA_PATH=$BUILD_DIR/cvl/schema
fi

echo "CVL schema directory is $CVL_SCHEMA_PATH"
if [ $(find $CVL_SCHEMA_PATH -name *.yin | wc -l) == 0 ]; then
    echo "WARNING: no yin files at $CVL_SCHEMA_PATH"
    echo ""
fi

# Prepare yang files directiry for transformer
if [ -z $YANG_MODELS_PATH ]; then
    export YANG_MODELS_PATH=$BUILD_DIR/all_yangs
    mkdir -p $YANG_MODELS_PATH
    pushd $YANG_MODELS_PATH > /dev/null
    rm -f *
    find $TOPDIR/models/yang -name "*.yang" -not -path "*/testdata/*" -exec ln -sf {} \;
    ln -sf $TOPDIR/config/transformer/models_list
    ln -sf $BUILD_DIR/yaml/api_ignore
    popd > /dev/null
fi

EXTRA_ARGS="-ui $SERVER_DIR/dist/ui -logtostderr"
HAS_CRTFILE=
HAS_KEYFILE=

for V in $@; do
    case $V in
    -cert|--cert|-cert=*|--cert=*) HAS_CRTFILE=1;;
    -key|--key|-key=*|--key=*) HAS_KEYFILE=1;;
    esac
done

cd $SERVER_DIR

##
# Setup TLS server cert/key pair
if [ -z $HAS_CRTFILE ] && [ -z $HAS_KEYFILE ]; then
    if [ -f cert.pem ] && [ -f key.pem ]; then
        echo "Reusing existing cert.pem and key.pem ..."
    else 
        echo "Generating temporary server certificate ..."
        ./generate_cert --host=localhost --ecdsa-curve=P256
    fi

    EXTRA_ARGS+=" -cert cert.pem -key key.pem"
fi

##
# Start server
./main $EXTRA_ARGS $* 

