SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

SUBDIRS := clitree renderer klish 
export SONIC_CLI_ROOT=$(TOPDIR)/build
TGT_DIR := $(SONIC_CLI_ROOT)/target

all: packages $(SUBDIRS)
$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: clean

all:
	for dir in $(SUBDIRS); do \
    	$(MAKE) -C $$dir -f Makefile $@; \
	done

	rm -rf $(TOPDIR)/build/cli/*
	mv -f $(TGT_DIR) $(TOPDIR)/build/cli

clean:
	make --directory=klish clean
	rm -rf $(TOPDIR)/build/cli
	rm -rf $(TGT_DIR)

packages:
	if ! dpkg -l | grep autoconf -c >>/dev/null; then sudo apt-get install autoconf; fi
	if ! dpkg -l | grep m4 -c >>/dev/null; then sudo apt-get install m4; fi
	if ! dpkg -l | grep libxml2-utils -c >>/dev/null; then sudo apt-get install libxml2-utils; fi
	if ! dpkg -l | grep xsltproc -c >>/dev/null; then sudo apt-get install xsltproc; fi
	if ! dpkg -l | grep python-lxml -c >>/dev/null; then sudo apt-get install python-lxml; fi
	if ! dpkg -l | grep libexpat1-dev -c >>/dev/null; then sudo apt-get install libexpat1-dev; fi

