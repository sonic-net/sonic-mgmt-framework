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


APT_DEPS_LIST = default-jre-headless \
                libxml2-dev \
                libxslt-dev

PIP_DEPS_LIST = pyang pyyaml

PIP2_DEPS_LIST = connexion python_dateutil certifi six urllib3

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

all: build-deps apt-deps pip-deps pip2-deps cli go-deps go-patch rest-server

build-deps:
	mkdir -p $(BUILD_DIR)

go-deps: $(GO_DEPS_LIST)
apt-deps: $(APT_DEPS_LIST)
pip-deps: $(PIP_DEPS_LIST)
pip2-deps: $(PIP2_DEPS_LIST)

$(GO_DEPS_LIST):
	$(GO) get -v $@

$(APT_DEPS_LIST):
	sudo apt-get install -y $@


$(PIP_DEPS_LIST):
	sudo pip3 install $@

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
	$(INSTALL) -D $(TOPDIR)/build/rest_server/dist/main $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -T $(TOPDIR)/src/cvl/build/pcre-8.43/install/lib/libpcre.so.1.2.11 $(DESTDIR)/usr/sbin/lib/libpcre.so.1
	$(INSTALL) -T $(TOPDIR)/src/cvl/build/libyang/build/libyang.so.1.1.* $(DESTDIR)/usr/sbin/lib/libyang.so.1
	$(INSTALL) -D $(TOPDIR)/src/cvl/build/libyang/build/extensions/*.so $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -D $(TOPDIR)/src/cvl/build/libyang/build/user_types/*.so $(DESTDIR)/usr/sbin/lib/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/
	cp -rf $(TOPDIR)/build/cli $(DESTDIR)/usr/sbin/
	cp -rf $(TOPDIR)/build/swagger_client_py/ $(DESTDIR)/usr/sbin/lib/


$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean:
	$(MAKE) -C src/cvl clean
	$(MAKE) -C src/cvl/schema clean
	$(MAKE) -C src/CLI clean
	$(MAKE) -C src/cvl cleanall
	rm -rf build
	rm -rf debian/.debhelper

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build
