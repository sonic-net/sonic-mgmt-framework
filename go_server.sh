#!/bin/bash

set -e

TOPDIR=$PWD
SERVER_DIR=$TOPDIR/build/rest_server/dist
CVLDIR=$TOPDIR/src/cvl

cd $SERVER_DIR

# LD_LIBRARY_PATH for CVL
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$CVLDIR/build/pcre-8.43/install/lib:$CVLDIR/build/libyang/build/lib64

# Setup CVL schema directory
if [ -z $CVL_SCHEMA_PATH ]; then
    export CVL_SCHEMA_PATH=$CVLDIR/schema
fi

if [ $(find $CVL_SCHEMA_PATH -name *.yin | wc -l) == 0 ]; then
    echo "WARNING: no yin files at $CVL_SCHEMA_PATH"
fi

##
# Start server
./main -ui $SERVER_DIR/ui -logtostderr $*

