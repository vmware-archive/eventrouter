# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

TARGET = eventrouter
REGISTRY ?= gcr.io/heptio-images
VERSION ?= v0.1
IMAGE = $(REGISTRY)/$(BIN)
BUILD_IMAGE ?= golang:1.7-alpine
DOCKER ?= docker
DIR := ${CURDIR}

all: local container
	
local: 
	$(DOCKER) run --rm -v $(DIR):/go/src/github.com/heptio/eventrouter -w /go/src/github.com/heptio/eventrouter $(BUILD_IMAGE) go build -v

container:
	$(DOCKER) build -t $(REGISTRY)/$(TARGET):latest -t $(REGISTRY)/$(TARGET):$(VERSION) .

# push: 
#	docker -- push $(REGISTRY)/$(TARGET)
#
# TODO: Determine tagging mechanics

.PHONY: all local container

clean: 
	rm -f eventrouter
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):$(VERSION)
