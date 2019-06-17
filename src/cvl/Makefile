all: precheck deps schema tests
GO=go
SRC_FILES=$(wildcard *.go)
TOP_DIR := $(abspath ../..)
GOFLAGS:=
BUILD_DIR:=build
GO_DEP_PATH=$(abspath .)/$(BUILD_DIR)
GO_TRANSLIB_PATH=$(TOP_DIR)/sonic-mgmt-framework
GO_SONIC_TELEMETRY_PATH=$(TOP_DIR)
GOPATH = $(GO_DEP_PATH):$(GO_TRANSLIB_PATH):$(GO_SONIC_TELEMETRY_PATH)

LIBPCRE:=pcre-8.43
LIBPCRE_PATH:=$(abspath .)/$(BUILD_DIR)/$(LIBPCRE)/install
LIBYANG_PATH:=$(abspath .)/$(BUILD_DIR)/libyang/build

ifdef DEBUG
	GOFLAGS += -gcflags="all=-N -l"
endif

precheck:
	$(shell mkdir -p $(BUILD_DIR))

deps: $(BUILD_DIR)/.lib_deps $(BUILD_DIR)/.deps cvl.a

$(BUILD_DIR)/.lib_deps:
	touch $@
	cd build && wget -O $(LIBPCRE).tar.gz https://ftp.pcre.org/pub/pcre/$(LIBPCRE).tar.gz && tar -xvzf $(LIBPCRE).tar.gz && cd -
	cd $(BUILD_DIR)/$(LIBPCRE) && ./configure --prefix=$(LIBPCRE_PATH) --enable-utf --enable-unicode-properties && make && make install && cd -
	cd build && git clone https://github.com/CESNET/libyang  && cd -
	@mkdir -p $(BUILD_DIR)/libyang/build
	cd $(LIBYANG_PATH) && cmake -D CMAKE_PREFIX_PATH=$(LIBPCRE_PATH)/include -D CMAKE_LIBRARY_PATH=$(LIBPCRE_PATH)/lib -D CMAKE_INSTALL_PREFIX:PATH=$(LIBYANG_PATH) -DPLUGINS_DIR:PATH=$(LIBYANG_PATH) .. && make && make install && cd -

	
$(BUILD_DIR)/.deps:
	touch $@
	GOPATH=$(GO_DEP_PATH) $(GO) get -u github.com/antchfx/jsonquery
	GOPATH=$(GO_DEP_PATH) $(GO) get -u github.com/antchfx/xmlquery
	GOPATH=$(GO_DEP_PATH) $(GO) get -u github.com/golang/glog
	GOPATH=$(GO_DEP_PATH) $(GO) get -u github.com/go-redis/redis

cvl.a:$(BUILD_DIR)/cvl.a

$(BUILD_DIR)/cvl.a:
	@echo "Building $@"
	GOPATH=$(GOPATH) $(GO) build $(GOFLAGS) -o $@ $(SRC_FILES)

install:
	GOPATH=$(GO_DEP_PATH) $(GO) install

schema:
	make -C schema

tests:
	make -C tests

clean:
	$(GO) clean
	make -C schema clean
	make -C tests clean
	rm -rf $(BUILD_DIR)/cvl.a

cleanall:
	rm -rf $(BUILD_DIR)

