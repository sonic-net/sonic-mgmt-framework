SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

KLISH_VERSION = 2.1.4

KLISH_SRC = $(SONIC_CLI_ROOT)/klish-$(KLISH_VERSION)

SRC_REPLACEMENTS:=$(shell find patches -type f)
all : $(SRC_REPLACEMENTS)
	tar xzvf klish-$(KLISH_VERSION).tgz -C $(SONIC_CLI_ROOT)
	cd ${KLISH_SRC} && QUILT_PATCHES=${CURDIR}/patches quilt push -a

ifeq ($(CROSS_BUILD_ENVIRON),y)
	export PATH=/python_virtualenv/env3/bin/:${PATH}
endif
	cd ${KLISH_SRC} && sh autogen.sh && ./configure --with-libxml2 --host=$(DEB_HOST_GNU_TYPE) --enable-debug=no CJSON_CFLAGS=-I/usr/include/cjson CJSON_LIBS=-lcjson && make

	mkdir -p $(SONIC_CLI_ROOT)/target/.libs
	cp $(CURDIR)/clish_start $(SONIC_CLI_ROOT)/target/.

	cp -r ${KLISH_SRC}/bin/.libs/clish   ${SONIC_CLI_ROOT}/target/.
	cp -r ${KLISH_SRC}/.libs/*.so* ${SONIC_CLI_ROOT}/target/.libs
	cp -r ${KLISH_SRC}/.libs/*.a   ${SONIC_CLI_ROOT}/target/.libs
	@echo "complete klish build"

.PHONY: clean
clean:
	rm -rf ${KLISH_SRC}
