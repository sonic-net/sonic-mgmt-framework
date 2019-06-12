all: deps schema tests install
.PHONY: deps schema tests install
GO=go
SRC_FILES=$(wildcard *.go)
TOPDIR := $(abspath ../..)
GOFLAGS:=
GOPATH = $(shell go env GOPATH):$(TOPDIR)

ifdef DEBUG
	GOFLAGS += -gcflags="all=-N -l"
endif

deps: 
	$(GO) get -u github.com/antchfx/jsonquery
	$(GO) get -u github.com/antchfx/xmlquery
	$(GO) get -u github.com/golang/glog
	$(GO) get -u github.com/go-redis/redis
install:
	GOPATH=$(GOPATH) $(GO) install
build:
	GOPATH=$(GOPATH) $(GO) build $(GOFLAGS) $(SRC_FILES)

schema:
	make -C schema

tests:
	make -C tests

clean:
	$(GO) clean
	make -C schema clean
	make -C tests clean
