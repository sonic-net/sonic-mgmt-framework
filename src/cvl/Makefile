all: precheck deps schema tests
GO?=/usr/local/go/bin/go
SRC_FILES=$(shell find . -name '*.go' | grep -v '_test.go' | grep -v '/tests/')
TEST_FILES=$(wildcard *_test.go)
TOP_DIR := $(abspath ../..)
GOFLAGS:=
BUILD_DIR:=build
GO_DOWNLOAD_PATH:=$(BUILD_GOPATH)
CVL_PKG=$(TOP_DIR)/pkg/linux_amd64/cvl.a

CVL_TEST_DIR = $(TOP_DIR)/build/tests/cvl
CVL_TEST_BIN = $(CVL_TEST_DIR)/cvl.test

ifdef DEBUG
	GOFLAGS += -gcflags="all=-N -l"
endif

precheck:
	$(shell mkdir -p $(BUILD_DIR))

deps:  $(BUILD_DIR)/.deps $(CVL_PKG) $(CVL_TEST_BIN)

	
$(BUILD_DIR)/.deps:
# Patch code
	@grep ParseJsonMap  $(GO_DOWNLOAD_PATH)/src/github.com/antchfx/jsonquery/node.go || \
	printf "\nfunc ParseJsonMap(jsonMap *map[string]interface{}) (*Node, error) {\n \
		doc := &Node{Type: DocumentNode}\n \
		parseValue(*jsonMap, doc, 1)\n \
		return doc, nil\n \
	}\n" >> $(GO_DOWNLOAD_PATH)/src/github.com/antchfx/jsonquery/node.go
	touch $@

$(CVL_PKG):
	@echo "Building $@"
	GOPATH=$(GOPATH) $(GO) build -v $(GOFLAGS) cvl
	GOPATH=$(GOPATH) $(GO) install cvl

$(CVL_TEST_BIN): $(TEST_FILES) $(SRC_FILES)
	GOPATH=$(GOPATH) $(GO) test -c -cover -coverpkg=cvl,cvl/internal/util,cvl/internal/yparser cvl -o $@
	cp -r testdata $(@D)/

install:
	GOPATH=$(GO_DEP_PATH) $(GO) install

schema:
	make -C schema

tests:
	make -C tests

gotest:
	 CVL_CFG_FILE=$(abspath .)/conf/cvl_cfg.json CVL_SCHEMA_PATH=$(abspath .)/schema GOPATH=$(GOPATH) tests/run_test.sh

clean:
	make -C tests clean
	rm -rf $(CVL_PKG)
	rm -rf $(CVL_TEST_DIR)

cleanall:
	rm -rf $(BUILD_DIR)
	rm -rf $(CVL_PKG)
	rm -rf $(CVL_TEST_DIR)

