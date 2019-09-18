SRC_FILES=$(wildcard *.go)
OUT=$(patsubst %.go, %, $(SRC_FILES)) 
TOPDIR := $(abspath ../../..)
GO=/usr/local/go/bin/go
GOPATH = $(TOPDIR):$(shell go env GOPATH)

all:tests

tests: $(OUT)

%:%.go
	make -C ../testdata/schema
	@echo "Building $@ ..."
	GOPATH=$(GOPATH) $(GO) build -gcflags="all=-N -l" $<

clean:
	@echo "Removing files ..."
	rm -rf $(OUT)
