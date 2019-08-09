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
	(cd ${TGT_DIR}/command-tree;xmllint --noout --schema sonic-clish.xsd ${TGT_DIR}/command-tree/*.xml) || exit 1
	rm -rf ${TGT_DIR}/cli-macro
	rm -rf scripts/*.pyc

clean:
	rm -rf ${TGT_DIR}
	@echo "Clitree Cleaned"

