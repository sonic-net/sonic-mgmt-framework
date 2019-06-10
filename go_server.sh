#!/bin/bash

set -e

TOPDIR=$PWD
SERVER_DIR=$TOPDIR/build/rest_server/dist

##
# Start server
$SERVER_DIR/main -ui $SERVER_DIR/ui -logtostderr $*
