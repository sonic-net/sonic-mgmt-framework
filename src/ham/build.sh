#!/usr/bin/env bash

SUBDIRS=$( find * -maxdepth 0 -type d | grep -v 'shared' | sort )

TARGET=$1
for dir in ${SUBDIRS}; do
	make -C ${dir} ${TARGET}
done

