#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

ifeq ($(TOPDIR),)
TOPDIR := ../..
endif

ifeq ($(BUILD_DIR),)
BUILD_DIR := $(TOPDIR)/build
endif

ifeq ($(GO),)
GO := go
endif

ifeq ($(GOPATH),)
ABS_TOPDIR := $(abspath $(TOPDIR))
GOPATH := $(ABS_TOPDIR)/gopkgs:$(ABS_TOPDIR):$(shell $(GO) env GOPATH)
endif

ifeq ($(GOROOT),)
GOROOT := $(shell $(GO) env GOROOT)
endif

REST_BUILD_DIR := $(BUILD_DIR)/rest_server
REST_BIN := $(REST_BUILD_DIR)/main

# Source files affecting REST server
REST_SRCS := $(shell find $(TOPDIR)/src -name '*.go' | sort) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

REST_GOPATH := $(GOPATH):$(abspath $(REST_BUILD_DIR)/dist)

# Certificate generator tool for generating temp certificates.
# Compiled from standard crypto/tls/generate_cert.go
CERTGEN_BIN := $(REST_BUILD_DIR)/generate_cert


# Default target
all: $(REST_BIN) $(CERTGEN_BIN)

# Invokes yang and model make to generate swagger artifcats.
$(REST_BIN): $(REST_SRCS)
	$(MAKE) -C $(TOPDIR)/models/yang
	$(MAKE) -C $(TOPDIR)/models
	GOPATH=$(REST_GOPATH) $(GO) build -gcflags=all="-N -l" -v -o $@ $(TOPDIR)/src/rest/main/main.go


# Compile a certificate generator temporary certificates from
# standard crypto/tls/generate_cert.go file. Source file will be
# available in GOROOT/src.
$(CERTGEN_BIN):
	mkdir -p $(@D)
	$(GO) build -o $@ $(GOROOT)/src/crypto/tls/generate_cert.go


clean:
	$(MAKE) -C $(TOPDIR)/models clean
	$(MAKE) -C $(TOPDIR)/models/yang clean
	rm -f $(REST_BIN) $(CERTGEN_BIN)

