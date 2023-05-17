################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

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

GO_DEPS_LIST = golang.org/x/crypto@https\://go.googlesource.com/crypto@0ec3e9974c59449edd84298612e9f16fa13368e8 \
               golang.org/x/sys@https\://go.googlesource.com/sys@85ca7c5b95cdf1e557abb38a283d1e61a5959c31 \
               golang.org/x/text@https\://go.googlesource.com/text@06d492aade888ab8698aad35476286b7b555c961 \
               golang.org/x/net@https\://go.googlesource.com/net@d3edc9973b7eb1fb302b0ff2c62357091cea9a30 \
               github.com/go-playground/validator.v9@https\://github.com/go-playground/validator.v9@21c910fc6d9c3556c28252b04beb17de0c2d40ec \
               google.golang.org/grpc@https\://github.com/grpc/grpc-go@192c8a2a3506bb69336f5c135733a3435a04fb30 \
               google.golang.org/genproto@https\://github.com/google/go-genproto@5b2d0af7952bff3ac94fc13b18af3efa8927cf94 \
               google.golang.org/protobuf@https\://go.googlesource.com/protobuf@d037755d51bc456237589295f966da868ee6dd35 \
               github.com/golang/protobuf@https\://github.com/golang/protobuf@84668698ea25b64748563aa20726db66a6b8d299 \
               github.com/golang/groupcache@https\://github.com/golang/groupcache@8c9f03a8e57eb486e42badaed3fb287da51807ba \
               github.com/golang/glog@https\://github.com/golang/glog@23def4e6c14b4da8ac2ed8007337bc5eb5007998 \
               github.com/go-redis/redis@https\://github.com/go-redis/redis@1c4dd844c436c4f293fd9f17c522027a5dc96e56 \
               github.com/openconfig/gnmi@https\://github.com/openconfig/gnmi@e7106f7f5493a9fa152d28ab314f2cc734244ed8 \
               github.com/openconfig/ygot@https\://github.com/openconfig/ygot@724a6b18a9224343ef04fe49199dfb6020ce132a \
               github.com/openconfig/goyang@https\://github.com/openconfig/goyang@a00bece872fc729c37e32bc697c8f3e7eb019172 \
               github.com/pkg/profile@https\://github.com/pkg/profile@acd64d450fd45fb2afa41f833f3788c8a7797219 \
               github.com/go-playground/locales@https\://github.com/go-playground/locales@9f105231d3a5f6877a2bf8321dfa15ea6f844b1b \
               github.com/go-playground/universal-translator@https\://github.com/go-playground/universal-translator@f87b1403479a348651dbf5f07f5cc6e5fcf07008 \
               github.com/leodido/go-urn@https\://github.com/leodido/go-urn@a0f5013415294bb94553821ace21a1a74c0298cc \
               github.com/Workiva/go-datastructures@https\://github.com/Workiva/go-datastructures@0819bcaf26091e7c33585441f8961854c2400faa \
               github.com/facette/natsort@https\://github.com/facette/natsort@2cd4dd1e2dcba4d85d6d3ead4adf4cfd2b70caf2 \
               github.com/kylelemons/godebug@https\://github.com/kylelemons/godebug@fa7b53cdfc9105c70f134574002f406232921437 \
               github.com/google/go-cmp@https\://github.com/google/go-cmp@f6dc95b586bc4e5c03cc308129693d9df2819e1c \
               github.com/antchfx/xpath@https\://github.com/antchfx/xpath@496661144dd35339be6985b945ae86a1b17d7064 \
               github.com/antchfx/jsonquery@https\://github.com/antchfx/jsonquery@29c5ac780efc692efb427b5f8863607c86133b38 \
               github.com/antchfx/xmlquery@https\://github.com/antchfx/xmlquery@96baba5f1e7e8d1e44efcbf1bb0827f6c8181232 \
               github.com/pborman/getopt@https\://github.com/pborman/getopt@ee0cd42419d3adee9239dbd1c375717fe482dac7 \
               github.com/gorilla/mux@https\://github.com/gorilla/mux@75dcda0896e109a2a22c9315bca3bb21b87b2ba5 \
               github.com/philopon/go-toposort@https\://github.com/philopon/go-toposort@9be86dbd762f98b5b9a4eca110a3f40ef31d0375

REST_BIN = $(BUILD_DIR)/rest_server/main
CERTGEN_BIN = $(BUILD_DIR)/rest_server/generate_cert

build-deps := $(BUILD_DIR)/.
go-deps    := $(BUILD_GOPATH)/.done
go-patch   := $(BUILD_GOPATH)/.patch_done

all: build-deps $(go-deps) $(go-patch) translib rest-server cli

build-deps:
	mkdir -p $(BUILD_DIR)

go-deps: $(GO_DEPS_LIST)

$(GO_DEPS_LIST):
	mkdir -p $(BUILD_GOPATH)/src/$(word 1,$(subst @, , $@))
	git clone $(word 2,$(subst @, , $@)) $(BUILD_GOPATH)/src/$(word 1,$(subst @, , $@)) && git -C $(BUILD_GOPATH)/src/$(word 1,$(subst @, , $@)) reset --hard $(word 3,$(subst @, , $@))

cli: rest-server
	$(MAKE) -C src/CLI

cvl: $(go-deps) $(go-patch)
	$(MAKE) -C src/cvl
	$(MAKE) -C src/cvl/schema
	$(MAKE) -C src/cvl/testdata/schema

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
	$(MAKE) -C models/yang/sonic

$(go-deps): $(MAKEFILE_LIST)
	$(RM) -r $(BUILD_GOPATH)
	$(MAKE) go-deps
	touch $@

$(go-patch): $(go-deps)
	cd $(BUILD_GOPATH)/src/github.com/openconfig/; \
	patch -p1 < $(TOPDIR)/ygot-modified-files/ygot.patch; rm -f ygot.patch; \
	$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ygot;
	cd $(BUILD_GOPATH)/src/github.com/openconfig/goyang; \
	patch -p1 < $(TOPDIR)/goyang-modified-files/goyang.patch; rm -f goyang.patch; \
	$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/goyang;
	cd $(BUILD_GOPATH)/src/github.com/antchfx/jsonquery; \
	git apply $(TOPDIR)/patches/jsonquery.patch; \
	$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/antchfx/jsonquery
	touch $@

install:
	$(INSTALL) -D $(REST_BIN) $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -D $(CERTGEN_BIN) $(DESTDIR)/usr/sbin/generate_cert
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -d $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/sonic/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/sonic/common/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/src/cvl/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -D $(TOPDIR)/src/cvl/testdata/schema/*.yin $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -D $(TOPDIR)/models/yang/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/config/transformer/models_list $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/common/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/annotations/*.yang $(DESTDIR)/usr/models/yang/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/
	cp -rf $(TOPDIR)/build/cli $(DESTDIR)/usr/sbin/
	cp -rf $(TOPDIR)/build/swagger_client_py/ $(DESTDIR)/usr/sbin/lib/
	cp -rf $(TOPDIR)/src/cvl/conf/cvl_cfg.json $(DESTDIR)/usr/sbin/cvl_cfg.json

ifeq ($(SONIC_COVERAGE_ON),y)
	echo "" > $(DESTDIR)/usr/sbin/.test
endif

$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean: rest-clean
	$(MAKE) -C src/cvl clean
	$(MAKE) -C src/translib clean
	$(MAKE) -C src/cvl/schema clean
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
	rm -rf debian/.debhelper
	rm -rf $(BUILD_GOPATH)/src/github.com/openconfig/goyang/annotate.go

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*
	rm -rf gopkgs
