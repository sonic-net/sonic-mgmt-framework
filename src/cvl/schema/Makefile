sonic_yang=../../../models/yang/sonic
sonic_yang_common=../../../models/yang/sonic/common
pyang_plugin_dir=../../../tools/pyang/pyang_plugins

src_files=$(wildcard $(sonic_yang)/*.yang)
src_files += $(wildcard $(sonic_yang_common)/*.yang)
out=$(patsubst %.yang, %.yin, $(shell ls -1 $(sonic_yang)/*.yang | cut -d'/' -f7)) 
out_common=$(patsubst %.yang, %.yin, $(shell ls -1 $(sonic_yang_common)/*.yang | cut -d'/' -f8)) 
out_tree=$(patsubst %.yang, %.tree, $(src_files)) 
search_path=$(sonic_yang):$(sonic_yang_common):$(sonic_yang_common)/ietf


all:schema

schema: $(out) $(out_common)

schema-tree: $(out_tree)

%.yin:$(sonic_yang)/%.yang
	@echo "Generating `basename $@` ..."
	@devFile="`echo $@ | cut -d . -f1`-deviation.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p $(search_path)	--plugindir $(pyang_plugin_dir) \
		-f yin-cvl $$devOpt $< -o `basename $@`

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
