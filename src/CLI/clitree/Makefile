.PHONY: install

TGT_DIR := $(SONIC_CLI_ROOT)/target

all:
	mkdir -p ${TGT_DIR}/command-tree ${TGT_DIR}/cli-macro ${TGT_DIR}/render-templates ${TGT_DIR}/scripts ${TGT_DIR}/sonic-cli-xml 
	cp cli-xml/*.xml ${TGT_DIR}/command-tree
	cp macro/*.xml ${TGT_DIR}/cli-macro
	(cd scripts;./klish_platform_features_process.sh ../../clicfg ${TGT_DIR})
	python scripts/klish_preproc_cmdtree.py ${TGT_DIR}/command-tree ${TGT_DIR}/cli-macro 3
	cp ./../actioner/*.py ${TGT_DIR}/.
	########## remove following line when docker with correctly packaged Klish parser engine available
	cp -r sonic-cli-xml ${TGT_DIR}/
	rm -rf ${TGT_DIR}/cli-macro
	rm -rf scripts/*.pyc

clean:
	@echo "nothing to clean for now"
	rm -rf TGT_DIR
	
