SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

KLISH_VERSION = 2.1.4
PYTHONVER=2.7

#MAIN_TARGET = klish-$(KLISH_VERSION)

SRC_REPLACEMENTS:=$(shell find patches -type f)
#$(addprefix $(DEST)/, $(MAIN_TARGET)): $(DEST)/% : $(SRC_REPLACEMENTS)
all : $(SRC_REPLACEMENTS)
	rm -rf klish-$(KLISH_VERSION)
	git clone https://src.libcode.org/pkun/klish.git klish-$(KLISH_VERSION)
	patches/scripts/patchmake.sh -p VER=${KLISH_VERSION} TSP=${CURDIR} DSP=${CURDIR}/patches TWP=${CURDIR}

	cd klish-$(KLISH_VERSION)
	sh autogen.sh && ./configure && make

	mkdir -p $(SONIC_CLI_ROOT)/target
	cp $(SONIC_CLI_ROOT)/klish/klish-$(KLISH_VERSION)/clish_start $(SONIC_CLI_ROOT)/.
	cp $(SONIC_CLI_ROOT)/klish/klish-$(KLISH_VERSION)/bin/clish $(SONIC_CLI_ROOT)/target/.
	@echo "complete klish build"

.PHONY: clean
clean:
	rm -rf klish-$(KLISH_VERSION)
