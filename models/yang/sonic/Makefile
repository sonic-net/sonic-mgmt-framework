TOPDIR := ../../../
SONIC_YANGAPI_DIR     := $(TOPDIR)/build/yaml
SONIC_YANGDIR         := $(TOPDIR)/models/yang/sonic
SONIC_YANGDIR_DEVIATION := $(TOPDIR)/models/yang/sonic/deviation
SONIC_YANGDIR_COMMON  := $(TOPDIR)/models/yang/sonic/common
SONIC_YANGDIR_COMMON_IETF := $(TOPDIR)/models/yang/sonic/common/ietf
SONIC_YANG_MOD_FILES  := $(shell find $(SONIC_YANGDIR) -maxdepth 1 -name '*.yang' | sort)
SONIC_YANG_COMMON_FILES  := $(shell find $(SONIC_YANGDIR_COMMON) -name '*.yang' | sort)
SONIC_YANG_COMMON_FILES  += $(shell find $(SONIC_YANGDIR_COMMON_IETF) -name '*.yang' | sort)

SONIC_TOOLS_DIR        := $(TOPDIR)/tools
SONIC_PYANG_DIR        := $(SONIC_TOOLS_DIR)/pyang
SONIC_PYANG_PLUGIN_DIR := $(SONIC_PYANG_DIR)/pyang_plugins
SONIC_PYANG_BIN        := pyang

all: yamlGen allyangs.tree allyangs_tree.html

#yamlGen: $(SONIC_YANGAPI_DIR)/.done

allyangs.tree:  $(SONIC_YANG_MOD_FILES) $(SONIC_YANG_COMMON_FILES)
	$(SONIC_PYANG_BIN) \
		-f tree \
		-o $(SONIC_YANGDIR)/$@ \
		-p $(SONIC_YANGDIR_COMMON):$(SONIC_YANGDIR) \
		$(SONIC_YANG_MOD_FILES)
	@echo "+++++ Generation of  YANG tree for Sonic Yang modules completed +++++"

allyangs_tree.html:  $(SONIC_YANG_MOD_FILES) $(SONIC_YANG_COMMON_FILES)
	$(SONIC_PYANG_BIN) \
		-f jstree \
		-o $(SONIC_YANGDIR)/$@ \
		-p $(SONIC_YANGDIR_COMMON):$(SONIC_YANGDIR) \
		$(SONIC_YANG_MOD_FILES)
	@echo "+++++ Generation of  HTML tree for Sonic Yang modules completed +++++"

#======================================================================
# Generate YAML files for SONiC YANG modules
#======================================================================
yamlGen:
	@echo "+++++ Generating YAML files for Sonic Yang modules +++++"
	mkdir -p $(SONIC_YANGAPI_DIR)
	$(SONIC_PYANG_BIN) \
		-f swaggerapi \
		--outdir $(SONIC_YANGAPI_DIR) \
		--plugindir $(SONIC_PYANG_PLUGIN_DIR) \
		-p $(SONIC_YANGDIR_COMMON):$(SONIC_YANGDIR) \
		$(SONIC_YANG_MOD_FILES)
	@echo "+++++ Generation of  YAML files for Sonic Yang modules completed +++++"

clean:
	@echo "Removing files ..."
	rm -rf $(SONIC_YANGAPI_DIR)
	rm -rf allyangs.tree allyangs_tree.html
