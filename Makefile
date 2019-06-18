#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

.PHONY: all clean cleanall codegen rest-server yamlGen

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
REST_DIST_DIR := $(BUILD_DIR)/rest_server/dist

# Source files affecting REST server
REST_SRCS := $(shell find $(TOPDIR)/src -name '*.go' | sort) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

CVL_GOPATH=$(TOPDIR):$(TOPDIR)/src/cvl/build
REST_BIN := $(REST_DIST_DIR)/main
REST_GOPATH = $(CVL_GOPATH):$(shell go env GOPATH):$(TOPDIR):$(REST_DIST_DIR)

#$(info REST_SRCS = $(REST_SRCS) )

all: rest-server

rest-server: $(REST_BIN)

yamlGen:
	$(MAKE) -C models/yang

$(REST_BIN): $(REST_SRCS)
	$(MAKE) -C src/cvl
	$(MAKE) -C models/yang
	$(MAKE) -C models
	GOPATH=$(REST_GOPATH) go build -o $@ $(TOPDIR)/src/rest/main/main.go

codegen:
	$(MAKE) -C models

clean:
	$(MAKE) -C src/cvl clean
	$(MAKE) -C models clean
	$(MAKE) -C models/yang clean

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build
