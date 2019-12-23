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

TOPDIR=../../../..
YANGDIR=$(TOPDIR)/models/yang
sonic_yang=$(YANGDIR)/sonic
std_yang_common=$(YANGDIR)/common/
sonic_yang_common=$(sonic_yang)/common
pyang_plugin_dir=$(TOPDIR)/tools/pyang/pyang_plugins
src_files=$(wildcard *.yang)
out=$(patsubst %.yang, %.yin, $(src_files)) 
out_ext=$(patsubst %.yang, %.tree, $(src_files)) 
search_path=$(std_yang_common):$(sonic_yang):$(sonic_yang_common)

all:schema

schema: $(out)

schema-tree: $(out_ext)

%.yin:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path) \
		--plugindir $(pyang_plugin_dir) -f yin-cvl $$devOpt $< -o `basename $@`

%.tree:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path) \
       		-f tree $$devOpt $< -o `basename $@`

clean:
	@echo "Removing files ..."
	rm -rf *.yin *.tree
