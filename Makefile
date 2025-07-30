# Copyright 2025 Yaacov Zamir <kobi.zamir@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

VERSION_GIT := $(shell git describe --tags 2>/dev/null || echo "0.0.0-dev")
VERSION ?= ${VERSION_GIT}

all: build


.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: build
build:
	go build -o ec2-to-ova ./cmd/main.go


.PHONY: golangci-lint
golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: golangci-lint
	go vet ./pkg/... ./cmd/...
	$(shell go env GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

.PHONY: dist
dist: build
	tar -zcvf ec2-to-ova.tar.gz LICENSE ec2-to-ova
	sha256sum ec2-to-ova.tar.gz > ec2-to-ova.tar.gz.sha256sum