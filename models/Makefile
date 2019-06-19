#######################################################################
#
# Copyright 2019 Broadcom. All rights reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
#
#######################################################################

TOPDIR := ..
ABS_TOPDIR := $(abspath $(TOPDIR))

BUILD_DIR := $(TOPDIR)/build
CODEGEN_TOOLS_DIR := $(TOPDIR)/tools/swagger_codegen

CODEGEN_VER	:= 2.4.5
CODEGEN_JAR := $(BUILD_DIR)/swagger-codegen-cli-$(CODEGEN_VER).jar

SERVER_CODEGEN_DIR	:= $(BUILD_DIR)/rest_server
SERVER_DIST_DIR		:= $(SERVER_CODEGEN_DIR)/dist
SERVER_DIST_INIT	:= $(SERVER_DIST_DIR)/.init_done
SERVER_DIST_GO		:= $(SERVER_DIST_DIR)/src/swagger
SERVER_DIST_UI		:= $(SERVER_DIST_DIR)/ui

PY_CLIENT_CODEGEN_DIR	:= $(BUILD_DIR)/swagger_client_py

YANGAPI_DIR     := $(TOPDIR)/build/yaml
YANGAPI_SPECS   := $(shell find $(YANGAPI_DIR) -name '*.yaml' | sort)
YANGAPI_NAMES   := $(basename $(notdir $(YANGAPI_SPECS)))
YANGAPI_SERVERS := $(addsuffix /.yangapi_done, $(addprefix $(SERVER_CODEGEN_DIR)/, $(YANGAPI_NAMES)))

OPENAPI_DIR   := openapi
OPENAPI_SPECS := $(shell find $(OPENAPI_DIR) -name '*.yaml' | sort)
OPENAPI_NAMES := $(basename $(notdir $(OPENAPI_SPECS)))
OPENAPI_SERVERS := $(addsuffix /.openapi_done, $(addprefix $(SERVER_CODEGEN_DIR)/, $(OPENAPI_NAMES)))

.PHONY: all clean cleanall go-server py-client

all: go-server py-client

go-server: $(YANGAPI_SERVERS) $(OPENAPI_SERVERS) $(SERVER_DIST_INIT)

py-client: $(PY_CLIENT_CODEGEN_DIR)/.done


.SECONDEXPANSION:

#======================================================================
# Common rule for directories. Use "." suffix, like "xyz/."
#======================================================================
.PRECIOUS: %/. 
%/.:
	mkdir -p $@

#======================================================================
# Download swagger codegen jar from Maven.org repo. It will be saved as
# build/swagger-codegen-cli.jar file.
#======================================================================
$(CODEGEN_JAR): | $$(@D)/.
	cd $(@D) && \
	wget http://central.maven.org/maven2/io/swagger/swagger-codegen-cli/$(CODEGEN_VER)/$(@F)

#======================================================================
# Generate swagger server in GO language for Yang generated OpenAPIs
# specs.
#======================================================================
%/.yangapi_done: $(YANGAPI_SPECS) | $$(@D)/. $(CODEGEN_JAR) $(SERVER_DIST_INIT)
	@echo "+++ Generating GO server for Yang API $$(basename $(@D)).yaml +++"
	java -jar $(CODEGEN_JAR) generate \
		--lang go-server \
		--input-spec $(YANGAPI_DIR)/$$(basename $(@D)).yaml \
		--template-dir $(CODEGEN_TOOLS_DIR)/go-server/templates-yang \
		--output $(@D)
	mv $(@D)/go/api_* $(SERVER_DIST_GO)/
	mv $(@D)/go/routers.go $(SERVER_DIST_GO)/routers_$$(basename $(@D)).go
	mv $(@D)/api/swagger.yaml $(SERVER_DIST_UI)/$$(basename $(@D)).yaml
	rm -rf $(@D)/*
	touch $@

#======================================================================
# Generate swagger server in GO language for handcoded OpenAPI specs
#======================================================================
%/.openapi_done: $(OPENAPI_SPECS) | $$(@D)/. $(CODEGEN_JAR) $(SERVER_DIST_INIT)
	@echo "+++ Generating GO server for OpenAPI $$(basename $(@D)).yaml +++"
	java -jar $(CODEGEN_JAR) generate \
		--lang go-server \
		--input-spec $(OPENAPI_DIR)/$$(basename $(@D)).yaml \
		--template-dir $(CODEGEN_TOOLS_DIR)/go-server/templates-nonyang \
		--output $(@D)
	mv $(@D)/go/api_* $(@D)/go/model_* $(SERVER_DIST_GO)/
	mv $(@D)/go/routers.go $(SERVER_DIST_GO)/routers_$$(basename $(@D)).go
	mv $(@D)/api/swagger.yaml $(SERVER_DIST_UI)/$$(basename $(@D)).yaml
	rm -rf $(@D)/*
	touch $@

#======================================================================
# Initialize dist directory for GO server code
#======================================================================
$(SERVER_DIST_INIT): | $$(@D)/.
	cp -r $(CODEGEN_TOOLS_DIR)/ui-dist $(@D)/ui
	cp -r $(CODEGEN_TOOLS_DIR)/go-server/src $(@D)/
	touch $@

#======================================================================
# Generate swagger client in Python.
# FIXME generating only for openconfig-acl.yaml for now
#======================================================================
$(PY_CLIENT_CODEGEN_DIR)/.done: | $(CODEGEN_JAR) $$(@D)/. $(YANGAPI_DIR)/openconfig-acl.yaml
	rm -rf $(@D)/*
	@echo "+++++ Generating Python client +++++"
	java -jar $(CODEGEN_JAR) generate \
		--lang python \
		--input-spec $(YANGAPI_DIR)/openconfig-acl.yaml \
		--output $(@D)
	touch $@

#======================================================================
# Cleanups
#======================================================================

clean:
	rm -rf $(SERVER_CODEGEN_DIR)
	rm -rf $(PY_CLIENT_CODEGEN_DIR)

cleanall: clean
	rm -f $(CODEGEN_JAR)

