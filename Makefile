HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-monitor
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION}"
BINARY := "helm-monitor"

.PHONY: dist release build install test lint vet dep

install: dep build
	cp $(BINARY) $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

build:
	go build -o $(BINARY) -ldflags $(LDFLAGS) ./cmd/...

dist:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) -ldflags $(LDFLAGS) ./cmd/...
	tar -zcvf $(DIST)/helm-monitor_linux_$(VERSION).tgz $(BINARY) README.md LICENSE plugin.yaml
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY) -ldflags $(LDFLAGS) ./cmd/...
	tar -zcvf $(DIST)/helm-monitor_darwin_$(VERSION).tgz $(BINARY) README.md LICENSE plugin.yaml
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe -ldflags $(LDFLAGS) ./cmd/...
	tar -zcvf $(DIST)/helm-monitor_windows_$(VERSION).tgz $(BINARY).exe README.md LICENSE plugin.yaml

test-all: vet lint test

test:
	go test -v -parallel=4 ./cmd/...

lint:
	@go get github.com/golang/lint/golint
	go list ./cmd/... | grep -v vendor | xargs -n1 golint

vet:
	go vet ./cmd/...

dep:
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
	dep ensure
