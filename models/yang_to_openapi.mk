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

TOPDIR := ..
BUILD_DIR := $(TOPDIR)/build

MGMT_COMMON_DIR             ?= $(TOPDIR)/../sonic-mgmt-common
YANGAPI_DIR                 := $(BUILD_DIR)/yaml
MD_DIR                      := $(TOPDIR)/build/restconf_md
SERVER_DIST_DIR             := $(BUILD_DIR)/rest_server/dist/openapi
YANGDIR                     := $(or $(wildcard $(MGMT_COMMON_DIR)/build/yang), $(MGMT_COMMON_DIR)/models/yang)
YANGDIR_COMMON              := $(YANGDIR)/common
YANGDIR_EXTENSIONS          := $(YANGDIR)/extensions
YANG_MOD_FILES              := $(wildcard $(YANGDIR)/*.yang)
YANG_MOD_FILES              += $(wildcard $(YANGDIR_EXTENSIONS)/*.yang)
YANG_COMMON_FILES           := $(wildcard $(YANGDIR_COMMON)/*.yang)

YANGDIR_SONIC               := $(YANGDIR)/sonic
YANGDIR_SONIC_COMMON        := $(YANGDIR_SONIC)/common
SONIC_YANG_MOD_FILES        := $(wildcard $(YANGDIR_SONIC)/*.yang)
SONIC_YANG_COMMON_FILES     := $(wildcard $(YANGDIR_SONIC_COMMON)/*.yang)

TOOLS_DIR        := $(TOPDIR)/tools
PYANG_PLUGIN_DIR := $(TOOLS_DIR)/pyang/pyang_plugins
PYANG ?= pyang

OPENAPI_GEN_PRE  := $(YANGAPI_DIR)/. $(MD_DIR)/. $(YANGAPI_DIR)/.openapi_gen_ut

all: $(YANGAPI_DIR)/.openapi_gen_ut $(YANGAPI_DIR)/.done $(YANGAPI_DIR)/.sonic_done

.PRECIOUS: %/.
%/.:
	mkdir -p $@

#======================================================================
# Unit tests for OpenAPI generator 
#======================================================================
$(YANGAPI_DIR)/.openapi_gen_ut: $(PYANG_PLUGIN_DIR)/openapi.py | $(YANGAPI_DIR)/.
	$(MAKE) -C $(TOOLS_DIR)/openapi_tests
	touch $@

#======================================================================
# Generate YAML files for Yang modules
#======================================================================
$(YANGAPI_DIR)/.done: $(YANG_MOD_FILES) $(YANG_COMMON_FILES) | $(OPENAPI_GEN_PRE)
	@echo "+++++ Generating YAML files for Yang modules +++++"
	$(PYANG) \
		-f swaggerapi \
		--outdir $(YANGAPI_DIR) \
		--plugindir $(PYANG_PLUGIN_DIR) \
		--with-md-doc \
		--md-outdir $(MD_DIR) \
		--with-serverstub \
		--stub-outdir $(SERVER_DIST_DIR) \
		-p $(YANGDIR_COMMON):$(YANGDIR):$(YANGDIR_EXTENSIONS) \
		$(YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Yang modules completed +++++"
	touch $@

#======================================================================
# Generate YAML files for SONiC YANG modules
#======================================================================
$(YANGAPI_DIR)/.sonic_done: $(SONIC_YANG_MOD_FILES) $(SONIC_YANG_COMMON_FILES) | $(OPENAPI_GEN_PRE)
	@echo "+++++ Generating YAML files for Sonic Yang modules +++++"
	$(PYANG) \
		-f swaggerapi \
		--with-md-doc \
		--outdir $(YANGAPI_DIR) \
		--md-outdir $(MD_DIR) \
		--plugindir $(PYANG_PLUGIN_DIR) \
		--with-serverstub \
		--stub-outdir $(SERVER_DIST_DIR) \
		-p $(YANGDIR_COMMON):$(YANGDIR_SONIC_COMMON):$(YANGDIR_SONIC) \
		$(SONIC_YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Sonic Yang modules completed +++++"
	touch $@

#======================================================================
# Cleanups
#======================================================================

clean:
	$(RM) -r $(YANGAPI_DIR)
	$(RM) -r $(MD_DIR)

cleanall: clean

