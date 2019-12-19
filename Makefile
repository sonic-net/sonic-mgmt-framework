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

.PHONY: all clean cleanall codegen rest-server rest-clean yamlGen cli clitree ham

TOPDIR := $(abspath .)
BUILD_DIR := $(TOPDIR)/build
export TOPDIR

ifeq ($(BUILD_GOPATH),)
export BUILD_GOPATH=$(TOPDIR)/build/gopkgs
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
               github.com/antchfx/xpath \
               github.com/antchfx/jsonquery \
               github.com/antchfx/xmlquery \
               github.com/facette/natsort \
               github.com/philopon/go-toposort \
               gopkg.in/godbus/dbus.v5 \
               github.com/dgrijalva/jwt-go \
               github.com/msteinert/pam


REST_BIN = $(BUILD_DIR)/rest_server/main
CERTGEN_BIN = $(BUILD_DIR)/rest_server/generate_cert

go-deps = $(BUILD_DIR)/gopkgs/.done
go-patch = $(BUILD_DIR)/gopkgs/.patch_done
go-redis-patch = $(BUILD_DIR)/gopkgs/.redis_patch_done

all: build-deps $(go-deps) $(go-redis-patch) $(go-patch) translib rest-server cli ham

build-deps:
	mkdir -p $(BUILD_DIR)/gopkgs

$(BUILD_DIR)/gopkgs/.done: $(MAKEFILE_LIST)
	$(GO) get -v $(GO_DEPS_LIST)
	touch  $@

$(go-redis-patch): $(go-deps)
	cd $(BUILD_GOPATH)/src/github.com/go-redis/redis; git checkout d19aba07b47683ef19378c4a4d43959672b7cec8 2>/dev/null ; true; \
$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/go-redis/redis
	touch  $@

cli: 
	$(MAKE) -C src/CLI

clitree:
	 TGT_DIR=$(BUILD_DIR)/cli $(MAKE) -C src/CLI/clitree

cvl: $(go-deps) $(go-patch) $(go-redis-patch)
	$(MAKE) -C src/cvl

cvl-test:
	$(MAKE) -C src/cvl gotest

rest-server: build-deps translib
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

ham:
	(cd src/ham; ./build.sh)

$(go-patch): $(go-deps)
	cd $(BUILD_GOPATH)/src/github.com/openconfig/ygot/; git reset --hard HEAD;git clean -f -d;git checkout 724a6b18a9224343ef04fe49199dfb6020ce132a 2>/dev/null ; true; \
cd ../; cp $(TOPDIR)/ygot-modified-files/ygot.patch .; \
patch -p1 < ygot.patch; rm -f ygot.patch; \
$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/ygot/ygot; \
	cd $(BUILD_GOPATH)/src/github.com/openconfig/goyang/; git reset --hard HEAD;git clean -f -d;git checkout 064f9690516f4f72db189f4690b84622c13b7296 >/dev/null ; true; \
cp $(TOPDIR)/goyang-modified-files/goyang.patch .; \
patch -p1 < goyang.patch; rm -f goyang.patch; \
$(GO) install -v -gcflags "-N -l" $(BUILD_GOPATH)/src/github.com/openconfig/goyang
	
#Apply CVL related patches
	$(apply_cvl_dep_patches)
	touch  $@


install:
	$(INSTALL) -D $(REST_BIN) $(DESTDIR)/usr/sbin/rest_server
	$(INSTALL) -D $(CERTGEN_BIN) $(DESTDIR)/usr/sbin/generate_cert
	$(INSTALL) -d $(DESTDIR)/usr/sbin/schema/
	$(INSTALL) -d $(DESTDIR)/usr/sbin/lib/
	$(INSTALL) -d $(DESTDIR)/usr/bin/
	$(INSTALL) -d $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/sonic/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/sonic/common/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/config/transformer/models_list $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/common/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/models/yang/annotations/*.yang $(DESTDIR)/usr/models/yang/
	$(INSTALL) -D $(TOPDIR)/build/yaml/api_ignore $(DESTDIR)/usr/models/yang/
	cp -rf $(TOPDIR)/build/rest_server/dist/ui/ $(DESTDIR)/rest_ui/
	cp -rf $(TOPDIR)/build/cli $(DESTDIR)/usr/sbin/
	rsync -a --exclude="test" --exclude="docs" build/swagger_client_py $(DESTDIR)/usr/sbin/lib/
	cp -rf $(TOPDIR)/src/cvl/conf/cvl_cfg.json $(DESTDIR)/usr/sbin/cvl_cfg.json

# Copy all CVL schema files
	cp -aT build/cvl/schema $(DESTDIR)/usr/sbin/schema

	# Scripts for host service
	$(INSTALL) -d $(DESTDIR)/usr/lib/sonic_host_service/host_modules
	$(INSTALL) -D $(TOPDIR)/scripts/sonic_host_server.py $(DESTDIR)/usr/lib/sonic_host_service
	$(INSTALL) -D $(TOPDIR)/scripts/host_modules/*.py $(DESTDIR)/usr/lib/sonic_host_service/host_modules
	$(INSTALL) -d $(DESTDIR)/etc/dbus-1/system.d
	$(INSTALL) -D $(TOPDIR)/scripts/org.sonic.hostservice.conf $(DESTDIR)/etc/dbus-1/system.d
	$(INSTALL) -d $(DESTDIR)/lib/systemd/system
	$(INSTALL) -D $(TOPDIR)/scripts/sonic-hostservice.service $(DESTDIR)/lib/systemd/system

	# Scripts for Host Account Management (HAM)
	$(INSTALL) -D $(TOPDIR)/src/ham/hamd/etc/dbus-1/system.d/* $(DESTDIR)/etc/dbus-1/system.d/
	$(INSTALL) -d $(DESTDIR)/etc/sonic/hamd/
	$(INSTALL) -D $(TOPDIR)/src/ham/hamd/etc/sonic/hamd/*      $(DESTDIR)/etc/sonic/hamd/
	$(INSTALL) -D $(TOPDIR)/src/ham/hamd/lib/systemd/system/*  $(DESTDIR)/lib/systemd/system/
	$(INSTALL) -D $(TOPDIR)/src/ham/hamd/usr/bin/*             $(DESTDIR)/usr/bin/
	$(INSTALL) -D $(TOPDIR)/src/ham/hamd/hamd     $(DESTDIR)/usr/sbin/.
	$(INSTALL) -D $(TOPDIR)/src/ham/hamctl/hamctl $(DESTDIR)/usr/bin/.
	$(INSTALL) -d $(DESTDIR)/lib/x86_64-linux-gnu/
	$(INSTALL) -D $(TOPDIR)/src/ham/libnss_ham/libnss_ham.so.2 $(DESTDIR)/lib/x86_64-linux-gnu/.

ifeq ($(SONIC_COVERAGE_ON),y)
	echo "" > $(DESTDIR)/usr/sbin/.test
endif

$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% :
	mv $* $(DEST)/

clean: rest-clean
	$(MAKE) -C src/translib clean
	$(MAKE) -C src/cvl clean
	rm -rf debian/.debhelper
	(cd build && find .  -maxdepth 1 -name "gopkgs" -prune -o -not -name '.' -exec rm -rf {} +) || true

cleanall:
	$(MAKE) -C src/cvl cleanall
	rm -rf build/*

#Function to apply CVL related patches
define apply_cvl_dep_patches

	cd $(BUILD_GOPATH)/src/github.com/antchfx/xpath; git reset --hard HEAD; \
	git checkout d9ad276609987dd73ce5cd7d6265fe82189b10b6; git apply $(TOPDIR)/patches/xpath.patch

	cd $(BUILD_GOPATH)/src/github.com/antchfx/jsonquery; git reset --hard HEAD; \
	git checkout 3b69d31134d889b501e166a035a4d5ecb8c6c367; git apply $(TOPDIR)/patches/jsonquery.patch

	cd $(BUILD_GOPATH)/src/github.com/antchfx/xmlquery; git reset --hard HEAD; \
	git checkout fe009d4cc63c3011f05e1dfa75a27899acccdf11; git apply $(TOPDIR)/patches/xmlquery.patch

endef
