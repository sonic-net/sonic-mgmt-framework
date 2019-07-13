#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

.PHONY: all clean cleanall codegen rest-server yamlGen cli

ifeq ($(GOPATH),)
export GOPATH=/tmp/go
endif

GO := /usr/local/go/bin/go 
export GO

INSTALL := /usr/bin/install

MAIN_TARGET = sonic-mgmt-framework_1.0-01_amd64.deb

GO_DEPS_LIST = github.com/gorilla/mux \
               github.com/Workiva/go-datastructures/queue \
               github.com/openconfig/goyang \
               github.com/openconfig/ygot/ygot \
               github.com/go-redis/redis \
               github.com/golang/glog \
               github.com/pkg/profile \
               gopkg.in/go-playground/validator.v9


PIP2_DEPS_LIST = connexion python_dateutil certifi

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build

export TOPDIR

# Source files affecting REST server
REST_SRCS := $(shell find $(TOPDIR)/src -name '*.go' | sort) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

CVL_GOPATH=$(TOPDIR):$(TOPDIR)/src/cvl/build
REST_BIN := $(REST_DIST_DIR)/main
REST_GOPATH = $(GOPATH):$(CVL_GOPATH):$(TOPDIR):$(REST_DIST_DIR)

#$(info REST_SRCS = $(REST_SRCS) )

all: build-deps pip2-deps cli go-deps go-patch rest-server

build-deps:
	mkdir -p $(BUILD_DIR)

go-deps: $(GO_DEPS_LIST)
pip2-deps: $(PIP2_DEPS_LIST)

$(GO_DEPS_LIST):
	$(GO) get -v $@

$(PIP2_DEPS_LIST):
	sudo pip install $@

cli:
	$(MAKE) -C src/CLI

cvl:
	$(MAKE) -C src/cvl
	$(MAKE) -C src/cvl/schema

REST_PREREQ := cvl
GOPATH := $(GOPATH):$(CVL_GOPATH)
include src/rest/Makefile

rest-server: $(REST_BIN)

yamlGen:
	$(MAKE) -C models/yang


codegen:
	$(MAKE) -C models

go-patch:
	cp $(TOPDIR)/ygot-modified-files/* /tmp/go/src/github.com/openconfig/ygot/ytypes/
	$(GO) install -v -gcflags "-N -l" /tmp/go/src/github.com/openconfig/ygot/ygot


install:
	$(INSTALL) -D $(REST_BIN) $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -D $(CERTGEN_BIN) $(DESTDIR)/usr/sbin/generate_cert
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/
	cp -rf $(TOPDIR)/build/cli $(DESTDIR)/usr/sbin/
	cp -rf $(TOPDIR)/build/swagger_client_py/ $(DESTDIR)/usr/sbin/lib/


$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean:
	$(MAKE) -C src/cvl clean
	$(MAKE) -C src/cvl/schema clean
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
	rm -rf debian/.debhelper

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
