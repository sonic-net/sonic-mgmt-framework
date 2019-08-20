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

TRANSLIB_PKG = $(TOPDIR)/pkg/linux_amd64/translib.a

TRANSLIB_MAIN_SRCS = $(shell find . -name '*.go' | grep -v '_test.go' | grep -v '/test/')
TRANSLIB_TEST_SRCS = $(shell find . -maxdepth 1 -name '*_test.go')
TRANSL_DB_ALL_SRCS = $(shell find db/ -name '*.go' | grep -v '/test/')

TRANSLIB_TEST_DIR  = $(BUILD_DIR)/tests/translib
TRANSLIB_TEST_BIN  = $(TRANSLIB_TEST_DIR)/translib.test
TRANSL_DB_TEST_BIN = $(TRANSLIB_TEST_DIR)/db.test

YANG_FILES = $(shell find $(TOPDIR)/models/yang -name '*.yang')
YGOT_BINDS = ocbinds/ocbinds.go

all: $(TRANSLIB_PKG) $(TRANSLIB_TEST_BIN) $(TRANSL_DB_TEST_BIN)

$(TRANSLIB_PKG): $(TRANSLIB_MAIN_SRCS) $(YGOT_BINDS)
	$(GO) build -gcflags="all=-N -l" -v translib
	$(GO) install translib

$(TRANSLIB_TEST_BIN): $(TRANSLIB_MAIN_SRCS) $(TRANSLIB_TEST_SRCS) $(YGOT_BINDS)
	$(GO) test -cover -coverpkg=translib,translib/tlerr -c translib -o $@

$(TRANSL_DB_TEST_BIN) : $(TRANSL_DB_ALL_SRCS)
	$(GO) test -cover -c translib/db -o $@

$(YGOT_BINDS): $(YANG_FILES)
	cd ocbinds && $(GO) generate

clean:
	rm -f $(YGOT_BINDS)
	rm -f $(TRANSLIB_PKG)
	rm -rf $(TRANSLIB_TEST_DIR)

