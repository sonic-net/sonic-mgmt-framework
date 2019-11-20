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

sonic_yang=../../../../models/yang/sonic
pyang_plugin_dir=../../../../tools/pyang/pyang_plugins
src_files=$(wildcard *.yang)
out_dir=$(TOPDIR)/build/cvl
out=$(patsubst %.yang, $(out_dir)/%.yin, $(src_files)) 
out_ext=$(patsubst %.yang, $(out_dir)/%.tree, $(src_files)) 

all:schema

schema: $(out)

schema-tree: $(out_ext)

$(out_dir)/%.yin:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(sonic_yang):$(sonic_yang)/common:$(sonic_yang)/common/ietf \
		--plugindir $(pyang_plugin_dir) -f yin-cvl $$devOpt $< -o $@

$(out_dir)/%.tree:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(sonic_yang):$(sonic_yang)/common:$(sonic_yang)/common/ietf \
			-f tree $$devOpt $< -o $@

clean:
	@echo "Removing files ..."
	rm -rf platform
