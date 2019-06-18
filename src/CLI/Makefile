SHELL = /bin/bash
.ONESHELL:
.SHELLFLAGS += -e

SUBDIRS := clitree renderer klish 
export SONIC_CLI_ROOT=$(shell pwd)

all: packages $(SUBDIRS)
$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: clean

all:
	for dir in $(SUBDIRS); do \
    	$(MAKE) -C $$dir -f Makefile $@; \
	done

clean:
	make --directory=klish clean
	rm -f clish_start
	rm -rf target

packages:
	if ! dpkg -l | grep autoconf -c >>/dev/null; then sudo apt-get install autoconf; fi
	if ! dpkg -l | grep m4 -c >>/dev/null; then sudo apt-get install m4; fi

