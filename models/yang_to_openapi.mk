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

YANGAPI_DIR                 := $(BUILD_DIR)/yaml
YANGDIR                     := yangs
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

OPENAPI_GEN_PRE  := $(YANGAPI_DIR)/.

all: $(YANGAPI_DIR)/.done $(YANGAPI_DIR)/.sonic_done

.PRECIOUS: %/.
%/.:
	mkdir -p $@

#======================================================================
# Generate YAML files for Yang modules
#======================================================================
$(YANGAPI_DIR)/.done:  $(YANG_MOD_FILES) $(YANG_COMMON_FILES) | $(OPENAPI_GEN_PRE)
	@echo "+++++ Generating YAML files for Yang modules +++++"
	mkdir -p $(YANGAPI_DIR)
	$(PYANG) \
		-f swaggerapi \
		--outdir $(@D) \
		--plugindir $(PYANG_PLUGIN_DIR) \
		-p $(YANGDIR_COMMON):$(YANGDIR) \
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
		--outdir $(@D) \
		--plugindir $(PYANG_PLUGIN_DIR) \
		-p $(YANGDIR_SONIC_COMMON):$(YANGDIR_SONIC):$(YANGDIR_COMMON) \
		$(SONIC_YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Sonic Yang modules completed +++++"
	touch $@

#======================================================================
# Cleanups
#======================================================================

clean:
	$(RM) -r $(YANGAPI_DIR)

