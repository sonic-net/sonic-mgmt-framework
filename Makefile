#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

.PHONY: all clean cleanall codegen rest-server rest-clean yamlGen cli

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
export TOPDIR

ifeq ($(BUILD_GOPATH),)
export BUILD_GOPATH=$(TOPDIR)/gopkgs
endif

export GOPATH=$(BUILD_GOPATH):$(TOPDIR)

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
               golang.org/x/crypto/ssh \
	       github.com/antchfx/jsonquery \
	       github.com/antchfx/xmlquery


REST_BIN = $(BUILD_DIR)/rest_server/main
CERTGEN_BIN = $(BUILD_DIR)/rest_server/generate_cert


all: build-deps go-deps go-patch translib rest-server cli

build-deps:
	mkdir -p $(BUILD_DIR)

go-deps: $(GO_DEPS_LIST)

$(GO_DEPS_LIST):
	$(GO) get -v $@

cli: rest-server
	$(MAKE) -C src/CLI

cvl: go-deps
	$(MAKE) -C src/cvl
	$(MAKE) -C src/cvl/schema
cvl-test:
	$(MAKE) -C src/cvl gotest

rest-server: translib
	$(MAKE) -C src/rest

rest-clean:
	$(MAKE) -C src/rest clean

translib: cvl
	$(MAKE) -C src/translib

codegen:
	$(MAKE) -C models

yamlGen:
	$(MAKE) -C models/yang

go-patch: go-deps
	cd $(BUILD_GOPATH)/src/github.com/openconfig/ygot/; git checkout 724a6b18a9224343ef04fe49199dfb6020ce132a 2>/dev/null ; true; \
cp $(TOPDIR)/ygot-modified-files/debug.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/../util/debug.go; \
cp $(TOPDIR)/ygot-modified-files/node.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/node.go; \
cp $(TOPDIR)/ygot-modified-files/container.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/container.go; \
cp $(TOPDIR)/ygot-modified-files/list.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/list.go; \
cp $(TOPDIR)/ygot-modified-files/leaf.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/leaf.go; \
cp $(TOPDIR)/ygot-modified-files/util_schema.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/util_schema.go; \
cp $(TOPDIR)/ygot-modified-files/schema.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/../util/schema.go; \
cp $(TOPDIR)/ygot-modified-files/unmarshal.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/unmarshal.go; \
cp $(TOPDIR)/ygot-modified-files/validate.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/validate.go; \
cp $(TOPDIR)/ygot-modified-files/reflect.go $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ytypes/../util/reflect.go; \
$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ygot

install:
	$(INSTALL) -D $(REST_BIN) $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -D $(CERTGEN_BIN) $(DESTDIR)/usr/sbin/generate_cert
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -d $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/common/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/annotations/*.yang $(DESTDIR)/usr/models/yang/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/
	cp -rf $(TOPDIR)/build/cli $(DESTDIR)/usr/sbin/
	cp -rf $(TOPDIR)/build/swagger_client_py/ $(DESTDIR)/usr/sbin/lib/


$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean: rest-clean
	$(MAKE) -C src/cvl clean
	$(MAKE) -C src/translib clean
	$(MAKE) -C src/cvl/schema clean
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
	rm -rf debian/.debhelper

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
