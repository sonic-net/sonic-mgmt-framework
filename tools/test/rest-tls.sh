#!/bin/bash

set -e

TEST_DIR=$(dirname "$(readlink -f "$0")")
TEST_BIN=$(mktemp /tmp/rest-tls-test-XXXXXX)
trap 'rm -f "$TEST_BIN"' EXIT

"${CXX:-g++}" -std=c++11 -Wall -Wextra -Werror "$TEST_DIR/rest-tls.cpp" -o "$TEST_BIN"
"$TEST_BIN"
