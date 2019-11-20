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

all: precheck deps schema tests
GO?=/usr/local/go/bin/go
SRC_FILES=$(shell find . -name '*.go' | grep -v '_test.go' | grep -v '/tests/')
TEST_FILES=$(wildcard *_test.go)
TOP_DIR := $(abspath ../..)
GOFLAGS:=
BUILD_DIR:=$(TOP_DIR)/build/cvl
GO_DOWNLOAD_PATH:=$(BUILD_GOPATH)
CVL_PKG=$(TOP_DIR)/pkg/linux_amd64/cvl.a

CVL_TEST_DIR = $(TOP_DIR)/build/tests/cvl
CVL_TEST_BIN = $(CVL_TEST_DIR)/cvl.test

ifdef DEBUG
	GOFLAGS += -gcflags="all=-N -l"
endif

precheck:
	$(shell mkdir -p $(BUILD_DIR))

deps:  $(BUILD_DIR)/.deps $(CVL_PKG) $(CVL_TEST_BIN)

	
$(BUILD_DIR)/.deps:
	touch $@

$(CVL_PKG):
	@echo "Building $@"
	GOPATH=$(GOPATH) $(GO) build -v $(GOFLAGS) cvl
	GOPATH=$(GOPATH) $(GO) install cvl

$(CVL_TEST_BIN): $(TEST_FILES) $(SRC_FILES)
	GOPATH=$(GOPATH) $(GO) test -c -cover -coverpkg=cvl,cvl/internal/util,cvl/internal/yparser cvl -o $@
	cp -r testdata $(@D)/

install:
	GOPATH=$(GO_DEP_PATH) $(GO) install

schema:
	make -C schema

tests:
	make -C tests

gotest:
	make -C schema
	make -C testdata/schema
	cp schema/*.yin testdata/schema
	rm -rf testdata/schema/platform
	cp -rf schema/platform testdata/schema/
	CVL_CFG_FILE=$(abspath .)/conf/cvl_cfg.json CVL_SCHEMA_PATH=$(abspath .)/testdata/schema GOPATH=$(GOPATH) tests/run_test.sh

clean:
	make -C tests clean
	make -C schema clean
	make -C testdata/schema clean
	rm -rf $(CVL_PKG)
	rm -rf $(CVL_TEST_DIR)

cleanall:clean
	rm -rf $(BUILD_DIR)
	rm -rf $(CVL_PKG)
	rm -rf $(CVL_TEST_DIR)

