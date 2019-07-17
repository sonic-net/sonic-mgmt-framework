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
GO = /usr/local/go/bin/go 
endif

ifeq ($(GOPATH),)
GOPATH = $(shell $(GO) env GOPATH)
endif

ifeq ($(GOROOT),)
GOROOT = $(shell $(GO) env GOROOT)
endif

TRANSLIB_PKG=$(TOP_DIR)/pkg/linux_amd64/translib.a

all:
	cd $(TOPDIR)/src/translib/ocbinds && $(GO) generate
	$(GO) build -gcflags=all="-N -l" -v translib
	$(GO) install translib


clean:
	rm -f $(TOPDIR)/src/translib/ocbinds/ocbinds.go
	rm -f $(TRANSLIB_PKG)

