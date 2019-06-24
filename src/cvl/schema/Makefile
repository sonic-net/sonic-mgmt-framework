src_files=$(wildcard *.yang)
out=$(patsubst %.yang, %.yin, $(src_files)) 
out_ext=$(patsubst %.yang, %.tree, $(src_files)) 

all:schema

schema: $(out) $(out_ext)

%.yin:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f yin $$devOpt $< -o $@.tmp
	@xmllint  --noblanks  $@.tmp > $@
	@rm -rf  $@.tmp

%.tree:%.yang
	@echo "Generating $@ ..."
	@devFile="`echo $< | cut -d . -f1`-dev.yang"; \
	if [ -f $$devFile ] ; then devOpt="--deviation-module $$devFile"; fi; \
	pyang -p ./ietf/ -f tree $$devOpt $< -o $@

clean:
	@echo "Removing files ..."
	rm -rf $(out)
	rm -rf $(out_ext)
