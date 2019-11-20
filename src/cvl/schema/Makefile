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

sonic_yang=../../../models/yang/sonic
sonic_yang_common=../../../models/yang/sonic/common
sonic_yang_platform=../../../models/yang/sonic/platform
pyang_plugin_dir=../../../tools/pyang/pyang_plugins

out_dir=$(TOPDIR)/build/cvl/schema
src_files=$(wildcard $(sonic_yang)/*.yang)
src_files += $(wildcard $(sonic_yang_common)/*.yang)
out=$(patsubst %.yang, $(out_dir)/%.yin, $(shell ls -1 $(sonic_yang)/*.yang | cut -d'/' -f7)) 
out_common=$(patsubst %.yang, $(out_dir)/%.yin, $(shell ls -1 $(sonic_yang_common)/*.yang | cut -d'/' -f8)) 
out_platform=$(patsubst %.yang, $(out_dir)/%.yin, $(shell find $(sonic_yang_platform) -name '*.yang' | cut -d'/' -f7-9))
out_platform_dep=$(shell find $(sonic_yang_platform) -name '*.yang')
out_tree=$(patsubst %.yang, $(out_dir)/%.tree, $(src_files)) 

search_path=$(sonic_yang):$(sonic_yang_common):$(sonic_yang_common)/ietf


all: schema

schema: $(out) $(out_common) $(out_platform)
	@$(call install_cvl_schema)

schema-tree: $(out_tree)

#Build YANG models
$(out_dir)/%.yin:$(sonic_yang)/%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $@ | cut -d . -f1`-deviation.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $$devOpt $< -o $@


#Build common YANG models
$(out_dir)/%.yin:$(sonic_yang_common)/%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $@ | cut -d . -f1`-deviation.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $$devOpt $< -o $@

#Build platform specific YANG models
$(out_platform):$(out_platform_dep)
	@mkdir -p `dirname $@`
	@echo "Generating $@ ..."
	@pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $< -o $@

$(out_dir)/%.tree:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path) -f tree $$devOpt $< -o `basename $@`

clean:
	@echo "Removing files ..."
	rm -rf $(out_dir)/*
	rm -rf platform/

