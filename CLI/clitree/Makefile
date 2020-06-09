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

.PHONY: install

TGT_DIR := $(SONIC_CLI_ROOT)/target

all:
	mkdir -p ${TGT_DIR}/command-tree ${TGT_DIR}/cli-macro ${TGT_DIR}/render-templates ${TGT_DIR}/scripts
	cp -r cli-xml/include ${TGT_DIR}/command-tree
	cp cli-xml/*.xml ${TGT_DIR}/command-tree
	cp macro/*.xml ${TGT_DIR}/cli-macro
	(cd scripts;./klish_platform_features_process.sh ../../clicfg ${TGT_DIR})
	python scripts/klish_preproc_cmdtree.py ${TGT_DIR}/command-tree ${TGT_DIR}/cli-macro 3
	cp ./../actioner/*.py ${TGT_DIR}/.
	cp ../renderer/scripts/*.py ${TGT_DIR}/scripts
	cp ../renderer/templates/* ${TGT_DIR}/render-templates
	cp scripts/sonic-clish.xsd ${TGT_DIR}/command-tree
	(cd ${TGT_DIR}/command-tree ; xmllint --noout --schema sonic-clish.xsd ${TGT_DIR}/command-tree/*.xml && \
            xmllint --noout --schema sonic-clish.xsd ${TGT_DIR}/command-tree/include/*.xml) || exit 1
	rm -rf ${TGT_DIR}/cli-macro
	rm -rf scripts/*.pyc

clean:
	rm -rf ${TGT_DIR}
	@echo "Clitree Cleaned"

