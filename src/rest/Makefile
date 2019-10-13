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

TOPDIR     ?= ../..
ABS_TOPDIR ?= $(abspath $(TOPDIR))
BUILD_DIR  ?= $(TOPDIR)/build

GO     ?= go
GOPATH ?= $(ABS_TOPDIR)/gopkgs:$(ABS_TOPDIR):$(shell $(GO) env GOPATH)
GOROOT ?= $(shell $(GO) env GOROOT)


REST_BUILD_DIR := $(BUILD_DIR)/rest_server
REST_BIN := $(REST_BUILD_DIR)/main
REST_TEST_BIN := $(BUILD_DIR)/tests/rest/server.test

ALL_GO_SRCS = $(shell find $(TOPDIR)/src -name '*.go' | grep -v '_test.go' | grep -v '/tests\?/')
REST_TEST_SRCS = $(shell find . -name '*_test.go')

# Source files affecting REST server
REST_SRCS := $(ALL_GO_SRCS) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/src/cvl/schema '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

REST_GOPATH := $(GOPATH):$(abspath $(REST_BUILD_DIR)/dist)

# Certificate generator tool for generating temp certificates.
# Compiled from standard crypto/tls/generate_cert.go
CERTGEN_BIN := $(REST_BUILD_DIR)/generate_cert


# Default target
all: $(REST_BIN) $(CERTGEN_BIN) $(REST_TEST_BIN)

$(REST_BUILD_DIR)/:
	mkdir -p $@

# REST Server binary
# Invokes yang and model make to generate swagger artifcats.
$(REST_BIN): $(REST_SRCS) | $(REST_BUILD_DIR)/
	$(MAKE) -C $(TOPDIR)/models/yang
	$(MAKE) -C $(TOPDIR)/models
ifeq ($(SONIC_COVERAGE_ON),y)
	GOPATH=$(REST_GOPATH) $(GO) test -coverpkg=".././..." -c -o $@ main/main.go main/main_test.go
else
	GOPATH=$(REST_GOPATH) $(GO) build -gcflags="all=-N -l" -v -o $@ main/main.go
endif

# Gotest binary for REST Server
$(REST_TEST_BIN): $(REST_TEST_SRCS) $(ALL_GO_SRCS) | $(REST_BUILD_DIR)/
	GOPATH=$(REST_GOPATH) $(GO) test -cover -c rest/server -o $@

# Compile a certificate generator temporary certificates from
# standard crypto/tls/generate_cert.go file. Source file will be
# available in GOROOT/src.
$(CERTGEN_BIN): | $(REST_BUILD_DIR)/
	$(GO) build -o $@ $(GOROOT)/src/crypto/tls/generate_cert.go

test:
	GOPATH=$(REST_GOPATH) $(GO) test -cover -v rest/server 

clean:
	$(MAKE) -C $(TOPDIR)/models clean
	$(MAKE) -C $(TOPDIR)/models/yang clean
	rm -f $(REST_BIN) $(CERTGEN_BIN)
	rm -f $(REST_TEST_BIN)

