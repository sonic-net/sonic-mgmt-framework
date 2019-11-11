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
pyang_plugin_dir=../../../tools/pyang/pyang_plugins

src_files=$(wildcard $(sonic_yang)/*.yang)
src_files += $(wildcard $(sonic_yang_common)/*.yang)
out=$(patsubst %.yang, %.yin, $(shell ls -1 $(sonic_yang)/*.yang | cut -d'/' -f7)) 
out_common=$(patsubst %.yang, %.yin, $(shell ls -1 $(sonic_yang_common)/*.yang | cut -d'/' -f8)) 
out_tree=$(patsubst %.yang, %.tree, $(src_files)) 
search_path=$(sonic_yang):$(sonic_yang_common):$(sonic_yang_common)/ietf

all: schema

schema: $(out) $(out_common)


schema-tree: $(out_tree)

#Build YANG models
%.yin:$(sonic_yang)/%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $@ | cut -d . -f1`-deviation.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $$devOpt $< -o `basename $@`


#Build common YANG models
%.yin:$(sonic_yang_common)/%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $@ | cut -d . -f1`-deviation.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $$devOpt $< -o `basename $@`

%.tree:%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path) -f tree $$devOpt $< -o `basename $@`

clean:
	@echo "Removing files ..."
	rm -rf *.yin *.tree
