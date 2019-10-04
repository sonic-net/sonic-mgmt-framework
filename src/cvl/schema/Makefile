TOPDIR := ../../..
CVL_YANGAPI_DIR     := $(TOPDIR)/build/yaml
CVL_YANGAPI_TARGET  := $(CVL_YANGAPI_DIR)/.cvl_done
CVL_YANGDIR         := .
CVL_YANGDIR_COMMON  := ./ietf

CVL_TOOLS_DIR        := $(TOPDIR)/tools
CVL_PYANG_DIR        := $(CVL_TOOLS_DIR)/pyang
CVL_PYANG_PLUGIN_DIR := $(CVL_PYANG_DIR)/pyang_plugins
CVL_PYANG_BIN        := pyang

src_files=$(wildcard *.yang)
out=$(patsubst %.yang, %.yin, $(src_files)) 
out_ext=$(patsubst %.yang, %.tree, $(src_files)) 

CVL_YANG_MOD_FILES		= $(src_files)
CVL_YANG_COMMON_FILES	= $(shell find $(CVL_YANGDIR_COMMON) -name '*.yang')

.PHONY: all yamlGen schema

all: yamlGen schema 

schema: $(out)

yamlGen: $(CVL_YANGAPI_TARGET)

%.yin:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f yin $$devOpt $< -o $@.tmp
	@xmllint  --noblanks  $@.tmp > $@
	@rm -rf  $@.tmp

#======================================================================
# Generate YAML files for Sonic Yang modules
#======================================================================
$(CVL_YANGAPI_TARGET): $(CVL_YANG_MOD_FILES) $(CVL_YANG_COMMON_FILES)
	@echo "+++++ Generating YAML files for Sonic Yang modules +++++"
	mkdir -p $(CVL_YANGAPI_DIR)
	$(CVL_PYANG_BIN) \
		-f swaggerapi \
		--outdir $(CVL_YANGAPI_DIR) \
		--plugindir $(CVL_PYANG_PLUGIN_DIR) \
		-p $(CVL_YANGDIR_COMMON) \
		$(CVL_YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Sonic Yang modules completed +++++"
	touch $@


%.tree:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f tree $$devOpt $< -o $@

clean:
	@echo "Removing files ..."
	rm -rf $(out)
	rm -rf $(out_ext)
	rm -f $(CVL_YANGAPI_TARGET)
	rm -f $(patsubst %.yang, $(CVL_YANGAPI_DIR)/%.yaml, $(src_files))
