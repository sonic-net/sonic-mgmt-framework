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

TOPDIR     := ..
BUILD_DIR  ?= $(TOPDIR)/build

GO     ?= go
GOROOT ?= $(shell $(GO) env GOROOT)

REST_BUILD_DIR := $(BUILD_DIR)/rest_server
REST_BIN := $(REST_BUILD_DIR)/rest_server
REST_TEST_BIN := $(BUILD_DIR)/tests/rest/server.test

REST_GO_SRCS   = $(shell find . -name '*.go' -not -name '*_test.go')
REST_TEST_SRCS = $(shell find . -name '*_test.go')

# Source files affecting REST server
REST_SRCS := $(REST_GO_SRCS)
REST_SRCS += $(shell find $(REST_BUILD_DIR)/dist -name '*.go')
REST_SRCS += $(shell find $(TOPDIR)/vendor/github.com/Azure/sonic-mgmt-common -name '*.go')
REST_SRCS += $(shell find $(TOPDIR)/models/openapi -name '*.yaml')

# Certificate generator tool for generating temp certificates.
# Compiled from standard crypto/tls/generate_cert.go
CERTGEN_BIN := $(REST_BUILD_DIR)/generate_cert


DEFAULT_TARGETS = $(REST_BIN) $(CERTGEN_BIN)
ifeq ($(NO_TEST_BINS),)
DEFAULT_TARGETS += $(REST_TEST_BIN)
endif

all: $(DEFAULT_TARGETS)

$(REST_BUILD_DIR)/:
	mkdir -p $@

# REST Server binary
# Invokes yang and model make to generate swagger artifcats.
$(REST_BIN): $(REST_SRCS) | $(REST_BUILD_DIR)/
ifeq ($(SONIC_COVERAGE_ON),y)
	$(GO) test -mod=vendor -coverpkg="./..." -c -o $@ ../rest/main/main.go ../rest/main/main_test.go
else
	$(GO) build -mod=vendor -gcflags="all=-N -l" -v -o $@ ../rest/main/main.go
endif

# Gotest binary for REST Server
$(REST_TEST_BIN): $(REST_TEST_SRCS) $(REST_GO_SRCS) | $(REST_BUILD_DIR)/
	$(GO) test -mod=vendor -cover -c ../rest/server -o $@

# Compile certificate generator from standard crypto/tls/generate_cert.go.
# Source file will be available in GOROOT/src.
$(CERTGEN_BIN): | $(REST_BUILD_DIR)/
	$(GO) build -o $@ $(GOROOT)/src/crypto/tls/generate_cert.go

clean:
	$(RM) $(REST_BIN) $(CERTGEN_BIN)
	$(RM) $(REST_TEST_BIN)

