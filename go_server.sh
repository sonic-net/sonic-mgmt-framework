#!/bin/bash

set -e

TOPDIR=$PWD
SERVER_DIR=$TOPDIR/build/rest_server/dist
CVLDIR=$TOPDIR/src/cvl

cd $SERVER_DIR

# LD_LIBRARY_PATH for CVL
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$CVLDIR/build/pcre-8.43/install/lib:$CVLDIR/build/libyang/build/lib64

# Setup CVL schema directory
if [ ! -d schema ]; then
    ln -s $CVLDIR/schema schema
fi

##
# Start server
./main -ui $SERVER_DIR/ui -logtostderr $*

