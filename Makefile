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

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
MGMT_COMMON_DIR := $(abspath ../sonic-mgmt-common)

GO      ?= /usr/local/go/bin/go
GOPATH  ?= /tmp/go

GO_MOD   = go.mod
GO_DEPS  = vendor/.done

export TOPDIR MGMT_COMMON_DIR GO GOPATH

.PHONY: all
all: rest cli

$(GO_MOD):
	$(GO) mod init github.com/Azure/sonic-mgmt-framework

$(GO_DEPS): $(GO_MOD)
	$(MAKE) -C models -f openapi_codegen.mk go-server-init
	$(GO) mod vendor
	$(MGMT_COMMON_DIR)/patches/apply.sh vendor
	touch  $@

go-deps: $(GO_DEPS)

go-deps-clean:
	$(RM) -r vendor

cli:
	$(MAKE) -C CLI

clitree:
	TGT_DIR=$(BUILD_DIR)/cli $(MAKE) -C CLI/clitree

clish:
	SONIC_CLI_ROOT=$(BUILD_DIR) $(MAKE) -C CLI/klish

.PHONY: rest
rest: $(GO_DEPS) models
	$(MAKE) -C rest

# Special target for local compilation of REST server binary.
# Compiles models, translib and cvl schema from sonic-mgmt-common
rest-server: go-deps-clean
	NO_TEST_BINS=1 $(MAKE) -C $(MGMT_COMMON_DIR)
	NO_TEST_BINS=1 $(MAKE) rest

rest-clean: go-deps-clean models-clean
	$(MAKE) -C rest clean

.PHONY: models
models:
	$(MAKE) -C models

models-clean:
	$(MAKE) -C models clean

clean: rest-clean models-clean
	git check-ignore debian/* | xargs -r $(RM) -r
	$(RM) -r debian/.debhelper
	$(RM) -r $(BUILD_DIR)

cleanall: clean

