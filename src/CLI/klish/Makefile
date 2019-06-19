SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

KLISH_VERSION = 2.1.4
PYTHONVER=2.7

KLISH_SRC = $(SONIC_CLI_ROOT)/klish/klish-$(KLISH_VERSION)

SRC_REPLACEMENTS:=$(shell find patches -type f)
all : $(SRC_REPLACEMENTS)
	rm -rf klish-$(KLISH_VERSION)
	git clone https://src.libcode.org/pkun/klish.git klish-$(KLISH_VERSION)
	patches/scripts/patchmake.sh -p VER=${KLISH_VERSION} TSP=${CURDIR} DSP=${CURDIR}/patches TWP=${CURDIR}

	cd klish-$(KLISH_VERSION)
	sh autogen.sh && ./configure && make

	mkdir -p $(SONIC_CLI_ROOT)/target/.libs
	cp $(KLISH_SRC)/clish_start $(SONIC_CLI_ROOT)/.

	cp -r ${KLISH_SRC}/bin/.libs/clish   ${SONIC_CLI_ROOT}/target/.
	cp -r ${KLISH_SRC}/.libs/*.so* ${SONIC_CLI_ROOT}/target/.libs
	cp -r ${KLISH_SRC}/.libs/*.a   ${SONIC_CLI_ROOT}/target/.libs
	@echo "complete klish build"
	touch $@

.PHONY: clean
clean:
	rm -rf klish-$(KLISH_VERSION)
