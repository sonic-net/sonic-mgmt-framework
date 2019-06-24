SRC_FILES=$(wildcard *.go)
OUT=$(patsubst %.go, %, $(SRC_FILES)) 
TOPDIR := $(abspath ../../..)
GO=/usr/local/go1.12/bin/go
GOPATH = $(TOPDIR):$(shell go env GOPATH)

all:tests

tests: $(OUT)

%:%.go
	@echo "Building $@ ..."
	GOPATH=$(GOPATH) $(GO) build -gcflags="all=-N -l" $<

clean:
	@echo "Removing files ..."
	rm -rf $(OUT)
