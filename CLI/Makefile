SHELL = /bin/bash
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################

.ONESHELL:
.SHELLFLAGS += -e

TOPDIR  ?= $(abspath ..)
SUBDIRS := clitree renderer klish 
export SONIC_CLI_ROOT=$(TOPDIR)/build
TGT_DIR := $(SONIC_CLI_ROOT)/target

CLI_DIR         := $(TOPDIR)/CLI
PY_FILES        := $(shell find $(CLI_DIR)/actioner -name '*.py')
PY_FILES        += $(shell find $(CLI_DIR)/renderer/scripts -name '*.py')
PY_FILES        += $(shell find $(CLI_DIR)/clitree/scripts -name '*.py')
PYLINT_DONE     := $(SONIC_CLI_ROOT)/pylint/.pylint_done
TOOLS_DIR       := $(TOPDIR)/tools
LINTER          := $(TOOLS_DIR)/test/pylint.sh

.PHONY: clean

all: $(PYLINT_DONE)
	for dir in $(SUBDIRS); do \
    	$(MAKE) -C $$dir -f Makefile $@; \
	done

	rm -rf $(TOPDIR)/build/cli
	mv -f $(TGT_DIR) $(TOPDIR)/build/cli

.SECONDEXPANSION:

#======================================================================
# Common rule for directories. Use "." suffix, like "xyz/."
#======================================================================
.PRECIOUS: %/. 
%/.:
	mkdir -p $@

pylint: $(PYLINT_DONE)

$(PYLINT_DONE): $(PY_FILES) | $$(@D)/.
	$(LINTER) $?
	touch $@

pylint-clean:
	$(RM) -r $(PYLINT_DONE)
	$(RM) -r $(SONIC_CLI_ROOT)/cli/pylint.log*

clean: pylint-clean
	make --directory=klish clean
	rm -rf $(TOPDIR)/build/cli
	rm -rf $(TGT_DIR)


