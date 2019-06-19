SRC_FILES=$(wildcard *.go)
OUT=$(patsubst %.go, %, $(SRC_FILES)) 
TOPDIR := $(abspath ../../..)
GOPATH = $(TOPDIR):$(shell go env GOPATH)

all:tests

tests: $(OUT)

%:%.go
	@echo "Building $@ ..."
	GOPATH=$(GOPATH) go build -gcflags="all=-N -l" $<

clean:
	@echo "Removing files ..."
	rm -rf $(OUT)
