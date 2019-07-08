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
GOPATH := $(shell $(GO) env GOPATH)
endif

REST_DIST_DIR := $(BUILD_DIR)/rest_server/dist
REST_BIN := $(REST_DIST_DIR)/main

# Source files affecting REST server
REST_SRCS := $(shell find $(TOPDIR)/src -name '*.go' | sort) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

REST_GOPATH := $(GOPATH):$(abspath $(TOPDIR)):$(abspath $(REST_DIST_DIR))


# Keep this as first target
# Invokes yang and model make to generate swagger artifcats.
# Top level makefile, which includes this makefile, can pass
# additional pre-requisites (like cvl) by setting REST_PREREQ
$(REST_BIN): $(REST_SRCS) $(REST_PREREQ)
	$(MAKE) -C $(TOPDIR)/models/yang
	$(MAKE) -C $(TOPDIR)/models
	GOPATH=$(REST_GOPATH) $(GO) build -gcflags=all="-N -l" -o $@ $(TOPDIR)/src/rest/main/main.go


rest-clean:
	$(MAKE) -C $(TOPDIR)/models clean
	$(MAKE) -C $(TOPDIR)/models/yang clean
	rm -f $(REST_BIN)

