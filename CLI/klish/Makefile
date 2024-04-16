SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

KLISH_VERSION = 2.1.4

KLISH_SRC = $(SONIC_CLI_ROOT)/klish-$(KLISH_VERSION)

# Python changed how to link against Python for embedding purposes in 3.8
# which means that in order to support both Debian 10 and Debian >=11
# we need to try two alternatives.
CFLAGS_PYTHON := $(shell pkg-config --silence-errors --cflags python3-embed || pkg-config --cflags python3)
LDFLAGS_PYTHON := $(shell pkg-config --silence-errors --libs python3-embed || pkg-config --libs python3)
LDFLAGS_CURL := $(shell pkg-config --libs libcurl)

ifeq ($(CROSS_BUILD_ENVIRON),y)
	CROSS_CONFIGURE_OPTS = --host=$(CROSS_HOST_TYPE)
	LIB_PATH = $(CROSS_PKGS_LIB_PATH)
else
	LIB_PATH = /usr/lib/x86_64-linux-gnu
endif

SRC_REPLACEMENTS:=$(shell find patches -type f)
all : $(SRC_REPLACEMENTS)
	tar xzvf klish-$(KLISH_VERSION).tgz -C $(SONIC_CLI_ROOT)
	./patches/scripts/patchmake.sh -p VER=${KLISH_VERSION} TSP=${SONIC_CLI_ROOT} DSP=${CURDIR}/patches TWP=${SONIC_CLI_ROOT}

	cd ${KLISH_SRC} && \
		sh autogen.sh && \
		./configure \
			--with-libxml2=/usr \
			$(CROSS_CONFIGURE_OPTS) \
			--enable-debug=no \
			LIBS='$(LDFLAGS_CURL) -L$(LIB_PATH) $(LDFLAGS_PYTHON) -Wl,-rpath=$(LIB_PATH) -lcjson' \
			CFLAGS='${CFLAGS} -g $(CFLAGS_PYTHON)' \
			CPPFLAGS='${CPPFLAGS} -I/usr/include/cjson' && \
		make

	mkdir -p $(SONIC_CLI_ROOT)/target/.libs
	cp $(CURDIR)/clish_start $(SONIC_CLI_ROOT)/target/.

	cp -r ${KLISH_SRC}/bin/.libs/clish   ${SONIC_CLI_ROOT}/target/.
	cp -r ${KLISH_SRC}/.libs/*.so* ${SONIC_CLI_ROOT}/target/.libs
	cp -r ${KLISH_SRC}/.libs/*.a   ${SONIC_CLI_ROOT}/target/.libs
	@echo "complete klish build"

.PHONY: clean
clean:
	rm -rf ${KLISH_SRC}
