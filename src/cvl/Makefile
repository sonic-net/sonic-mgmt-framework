all: precheck deps schema tests
GO?=/usr/local/go/bin/go
SRC_FILES=$(wildcard *.go)
TOP_DIR := $(abspath ../..)
GOFLAGS:=
BUILD_DIR:=build
GO_DOWNLOAD_PATH:=$(BUILD_GOPATH)

ifdef DEBUG
	GOFLAGS += -gcflags="all=-N -l"
endif

precheck:
	$(shell mkdir -p $(BUILD_DIR))

deps:  $(BUILD_DIR)/.deps cvl.a

	
$(BUILD_DIR)/.deps:
# Patch code
	@grep ParseJsonMap  $(GO_DOWNLOAD_PATH)/src/github.com/antchfx/jsonquery/node.go || \
	printf "\nfunc ParseJsonMap(jsonMap *map[string]interface{}) (*Node, error) {\n \
		doc := &Node{Type: DocumentNode}\n \
		parseValue(*jsonMap, doc, 1)\n \
		return doc, nil\n \
	}\n" >> $(GO_DOWNLOAD_PATH)/src/github.com/antchfx/jsonquery/node.go
	touch $@

cvl.a:$(BUILD_DIR)/cvl.a

$(BUILD_DIR)/cvl.a:
	@echo "Building $@"
	GOPATH=$(GOPATH) $(GO) build -v $(GOFLAGS) cvl
	GOPATH=$(GOPATH) $(GO) install cvl

install:
	GOPATH=$(GO_DEP_PATH) $(GO) install

schema:
	make -C schema

tests:
	make -C tests

gotest:
	CVL_SCHEMA=$(abspath .)/schema GOPATH=$(GOPATH) tests/run_test.sh

clean:
	make -C tests clean
	rm -rf $(BUILD_DIR)/cvl.a

cleanall:
	rm -rf $(BUILD_DIR)

