#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

.PHONY: all clean cleanall codegen rest-server yamlGen

ifeq ($(GOPATH),)
export GOPATH=/tmp/go
endif

GOROOT := /usr/local/go1.12
GO := $(GOROOT)/bin/go
export GO

INSTALL := /usr/bin/install

MAIN_TARGET = sonic-mgmt-framework_1.0-01_amd64.deb

GO_DEPS_LIST = github.com/gorilla/mux \
               github.com/Workiva/go-datastructures/queue \
               github.com/openconfig/goyang \
               github.com/openconfig/ygot/ygot \
               github.com/go-redis/redis \
               github.com/golang/glog \
               gopkg.in/go-playground/validator.v9


APT_DEPS_LIST = default-jre-headless \
                libxml2-dev \
                libxslt-dev

PIP_DEPS_LIST = pyang pyyaml

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
CVL_GOPATH=$(TOPDIR)/src/cvl/build


all: golang go-deps go-patch apt-deps pip-deps rest-server

golang:
	wget https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz
	tar -zxvf go1.12.6.linux-amd64.tar.gz
	sudo mv go /usr/local/go1.12

go-deps: $(GO_DEPS_LIST)
apt-deps: $(APT_DEPS_LIST)
pip-deps: $(PIP_DEPS_LIST)

$(GO_DEPS_LIST):
	/usr/local/go/bin/go get -v $@
	/usr/local/go/bin/go get -v $@

$(APT_DEPS_LIST):
	sudo apt-get install -y $@


$(PIP_DEPS_LIST):
	sudo pip3 install $@

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
	/usr/local/go1.12/bin/go install -v -gcflags "-N -l" /tmp/go/src/github.com/openconfig/ygot/ygot


install:
	$(INSTALL) -D $(TOPDIR)/build/rest_server/dist/main $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -T $(TOPDIR)/src/cvl/build/pcre-8.43/install/lib/libpcre.so.1.2.11 $(DESTDIR)/usr/sbin/lib/libpcre.so.1
	$(INSTALL) -T $(TOPDIR)/src/cvl/build/libyang/build/libyang.so.1.1.25 $(DESTDIR)/usr/sbin/lib/libyang.so.1
	$(INSTALL) -D $(TOPDIR)/src/cvl/build/libyang/build/extensions/*.so $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -D $(TOPDIR)/src/cvl/build/libyang/build/user_types/*.so $(DESTDIR)/usr/sbin/lib/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/

$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean: rest-clean
	$(MAKE) -C src/cvl clean
	$(MAKE) -C src/cvl/schema clean

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build

