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
REST_DIST_DIR := $(BUILD_DIR)/rest_server/dist

# Source files affecting REST server
REST_SRCS := $(shell find $(TOPDIR)/src -name '*.go' | sort) \
			 $(shell find $(TOPDIR)/models/yang -name '*.yang' | sort) \
			 $(shell find $(TOPDIR)/models/openapi -name '*.yaml' | sort)

CVL_GOPATH=$(TOPDIR):$(TOPDIR)/src/cvl/build
REST_BIN := $(REST_DIST_DIR)/main
REST_GOPATH = $(GOPATH):$(CVL_GOPATH):$(TOPDIR):$(REST_DIST_DIR)

#$(info REST_SRCS = $(REST_SRCS) )

all: golang go-deps apt-deps pip-deps rest-server

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

rest-server: $(REST_BIN)

yamlGen:
	$(MAKE) -C models/yang

$(REST_BIN): $(REST_SRCS)
	$(MAKE) -C src/cvl
	$(MAKE) -C models/yang
	$(MAKE) -C models
	GOPATH=$(REST_GOPATH) /usr/local/go1.12/bin/go build -o $@ $(TOPDIR)/src/rest/main/main.go

codegen:
	$(MAKE) -C models


install:
	$(INSTALL) -D $(TOPDIR)/build/rest_server/dist/main $(DESTDIR)/usr/sbin/rest_server
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/

$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean:
	$(MAKE) -C src/cvl clean
	$(MAKE) -C models clean
	$(MAKE) -C models/yang clean

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build
