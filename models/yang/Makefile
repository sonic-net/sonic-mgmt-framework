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

TOPDIR := ../..
BUILD_DIR := $(TOPDIR)/build

YANGAPI_DIR     := $(TOPDIR)/build/yaml
YANGDIR         := $(TOPDIR)/models/yang
YANGDIR_COMMON  := $(TOPDIR)/models/yang/common
YANG_MOD_FILES  := $(shell find $(YANGDIR) -maxdepth 1 -name '*.yang' | sort)
YANG_COMMON_FILES  := $(shell find $(YANGDIR_COMMON) -name '*.yang' | sort)

TOOLS_DIR        := $(TOPDIR)/tools
PYANG_DIR        := $(TOOLS_DIR)/pyang
PYANG_PLUGIN_DIR := $(PYANG_DIR)/pyang_plugins
PYANG_BIN        := pyang

.PHONY: all yamlGen

all: yamlGen allyangs.tree allyangs_tree.html 

yamlGen: $(YANGAPI_DIR)/.done

allyangs.tree:  $(YANG_MOD_FILES) $(YANG_COMMON_FILES)
	$(PYANG_BIN) \
		-f tree \
		-o $(YANGDIR)/$@ \
		-p $(YANGDIR_COMMON):$(YANGDIR) \
		$(YANG_MOD_FILES)
	@echo "+++++ Generation of  YANG tree for Yang modules completed +++++"

allyangs_tree.html:  $(YANG_MOD_FILES) $(YANG_COMMON_FILES)
	$(PYANG_BIN) \
		-f jstree \
		-o $(YANGDIR)/$@ \
		-p $(YANGDIR_COMMON):$(YANGDIR) \
		$(YANG_MOD_FILES)
	@echo "+++++ Generation of  HTML tree for Yang modules completed +++++"

#======================================================================
# Generate YAML files for Yang modules
#======================================================================
$(YANGAPI_DIR)/.done:  $(YANG_MOD_FILES) $(YANG_COMMON_FILES)
	@echo "+++++ Generating YAML files for Yang modules +++++"
	mkdir -p $(YANGAPI_DIR)
	$(PYANG_BIN) \
		-f swaggerapi \
		--outdir $(YANGAPI_DIR) \
		--plugindir $(PYANG_PLUGIN_DIR) \
		-p $(YANGDIR_COMMON):$(YANGDIR) \
		$(YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Yang modules completed +++++"
	touch $@

#======================================================================
# Cleanups
#======================================================================

clean:
	rm -rf $(YANGAPI_DIR)
	rm -rf allyangs.tree allyangs_tree.html
