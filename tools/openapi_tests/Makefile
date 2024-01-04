.PHONY: all test-rpc test-data-nodes

TOPDIR             := ../..
PYANG_PLUGINS_DIR  := $(TOPDIR)/tools/pyang/pyang_plugins
PYANG ?= pyang

all: test-rpc test-data-nodes test-complex-model test-complex-model-no-oneof

# This will cover RPC statements with only input, only output and with
# both input and output
test-rpc:
	$(PYANG) -f swaggerapi --plugindir $(PYANG_PLUGINS_DIR) test-rpc.yang | diff test-rpc.yang.expect -

# This will cover container, list, leaf and leaf-lists with both
# simple and nested hierarchies.
# Also this will cover data type testing such as leafref, enum and string with pattern
# simple string, integer types, leaf with default values, mandatory statements etc.
test-data-nodes:
	$(PYANG) -f swaggerapi --plugindir $(PYANG_PLUGINS_DIR) test-data-nodes.yang | diff test-data-nodes.yang.expect -

# This will cover some complex YANGs, with many nested hierarchies
# Also with choice-case statements, Union data types, range with min,max etc
# Test with one-oneof
test-complex-model:
	$(PYANG) -f swaggerapi --plugindir $(PYANG_PLUGINS_DIR) --with-oneof ietf-snmp.yang ietf-snmp-community.yang | diff ietf-snmp.yang.expect -

# Test without one-oneof
test-complex-model-no-oneof:
	$(PYANG) -f swaggerapi --plugindir $(PYANG_PLUGINS_DIR) ietf-snmp.yang ietf-snmp-community.yang | diff ietf-snmp.no-oneof.yang.expect -
