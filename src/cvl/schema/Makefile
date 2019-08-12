TOPDIR := ../../..
CVL_YANGAPI_DIR     := $(TOPDIR)/build/yaml
CVL_YANGDIR         := $(TOPDIR)/src/cvl/schema
CVL_YANGDIR_COMMON  := $(TOPDIR)/src/cvl/schema/ietf
CVL_YANG_MOD_FILES  := $(shell find $(CVL_YANGDIR) -maxdepth 1 -name '*.yang' | sort)
CVL_YANG_COMMON_FILES  := $(shell find $(CVL_YANGDIR_COMMON) -name '*.yang' | sort)

CVL_TOOLS_DIR        := $(TOPDIR)/tools
CVL_PYANG_DIR        := $(CVL_TOOLS_DIR)/pyang
CVL_PYANG_PLUGIN_DIR := $(CVL_PYANG_DIR)/pyang_plugins
CVL_PYANG_BIN        := pyang

src_files=$(wildcard *.yang)
out=$(patsubst %.yang, %.yin, $(src_files)) 
out_ext=$(patsubst %.yang, %.tree, $(src_files)) 

all: yamlGen allyangs.tree allyangs_tree.html schema 

schema: $(out)

#yamlGen: $(CVL_YANGAPI_DIR)/.done

%.yin:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f yin $$devOpt $< -o $@.tmp
	@xmllint  --noblanks  $@.tmp > $@
	@rm -rf  $@.tmp

allyangs.tree:  $(YANG_MOD_FILES) $(YANG_COMMON_FILES)
	$(CVL_PYANG_BIN) \
		-f tree \
		-o $(CVL_YANGDIR)/$@ \
		-p $(CVL_YANGDIR_COMMON):$(CVL_YANGDIR) \
		$(CVL_YANG_MOD_FILES)
	@echo "+++++ Generation of  YANG tree for Sonic Yang modules completed +++++"

allyangs_tree.html:  $(CVL_YANG_MOD_FILES) $(CVL_YANG_COMMON_FILES)
	$(CVL_PYANG_BIN) \
		-f jstree \
		-o $(CVL_YANGDIR)/$@ \
		-p $(CVL_YANGDIR_COMMON):$(CVL_YANGDIR) \
		$(CVL_YANG_MOD_FILES)
	@echo "+++++ Generation of  HTML tree for Sonic Yang modules completed +++++"

#======================================================================
# Generate YAML files for Sonic Yang modules
#======================================================================
yamlGen:
	@echo "+++++ Generating YAML files for Sonic Yang modules +++++"
	mkdir -p $(CVL_YANGAPI_DIR)
	$(CVL_PYANG_BIN) \
		-f swaggerapi \
		--outdir $(CVL_YANGAPI_DIR) \
		--plugindir $(CVL_PYANG_PLUGIN_DIR) \
		-p $(CVL_YANGDIR_COMMON):$(CVL_YANGDIR) \
		$(CVL_YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Sonic Yang modules completed +++++"
	# touch $@


%.tree:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f tree $$devOpt $< -o $@

clean:
	@echo "Removing files ..."
	rm -rf $(out)
	rm -rf $(out_ext)
	rm -rf $(CVL_YANGAPI_DIR)
	rm -rf allyangs.tree allyangs_tree.html
