#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

.PHONY: all clean cleanall codegen rest-server yamlGen cli

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
export TOPDIR

ifeq ($(BUILD_GOPATH),)
export BUILD_GOPATH=$(TOPDIR)/gopkgs
endif

ifeq ($(GOPATH),)
export GOPATH=$(BUILD_GOPATH)
endif

ifeq ($(GO),)
GO := /usr/local/go/bin/go 
export GO
endif

INSTALL := /usr/bin/install

MAIN_TARGET = sonic-mgmt-framework_1.0-01_amd64.deb

GO_DEPS_LIST = github.com/gorilla/mux \
               github.com/Workiva/go-datastructures/queue \
               github.com/openconfig/goyang \
               github.com/openconfig/ygot/ygot \
               github.com/go-redis/redis \
               github.com/golang/glog \
               github.com/pkg/profile \
               gopkg.in/go-playground/validator.v9 \
               github.com/msteinert/pam \
               golang.org/x/crypto/ssh \
	       github.com/antchfx/jsonquery \
	       github.com/antchfx/xmlquery


PIP2_DEPS_LIST = connexion python_dateutil certifi
REST_BIN = $(BUILD_DIR)/rest_server/dist/main
CERTGEN_BIN = $(BUILD_DIR)/rest_server/dist/generate_cert

CVL_GOPATH=$(TOPDIR):$(TOPDIR)/src/cvl/build
GOPATH := $(GOPATH):$(CVL_GOPATH)


all: build-deps pip2-deps go-deps go-patch translib rest-server cli

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

rest-server:
	$(MAKE) -C src/rest

translib: cvl
	$(MAKE) -C src/translib


codegen:
	$(MAKE) -C models

yamlGen:
	$(MAKE) -C models/yang

go-patch:
	cp $(TOPDIR)/ygot-modified-files/* $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/
	$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ygot


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
	$(MAKE) -C src/translib clean
	$(MAKE) -C models clean
	$(MAKE) -C src/cvl/schema clean
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
	rm -rf debian/.debhelper

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
