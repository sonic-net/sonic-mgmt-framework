################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY   #
#  KIND, either express or implied. See the License for the specific language #
#  governing permissions and limitations under the License.                    #
#                                                                              #
################################################################################

TOPDIR     := ..
BUILD_DIR  ?= $(TOPDIR)/build

GO     ?= go
GOROOT ?= $(shell $(GO) env GOROOT)

OBSERVABILITYD_BUILD_DIR := $(BUILD_DIR)/observabilityd
OBSERVABILITYD_BIN := $(OBSERVABILITYD_BUILD_DIR)/observabilityd

OBSERVABILITYD_GO_SRCS = $(shell find . -name '*.go' -not -name '*_test.go')

all: $(OBSERVABILITYD_BIN)

$(OBSERVABILITYD_BUILD_DIR)/:
	mkdir -p $@

# Observabilityd binary
$(OBSERVABILITYD_BIN): $(OBSERVABILITYD_GO_SRCS) | $(OBSERVABILITYD_BUILD_DIR)/
	@echo 'Building observabilityd...'
	$(GO) mod tidy
	$(GO) build -buildvcs=false -o $@ .
	@echo 'Finished building observabilityd'

clean:
	$(RM) $(OBSERVABILITYD_BIN)

.PHONY: all clean
