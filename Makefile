SHELL := /bin/bash
CURRENT_PATH = $(shell pwd)
DISTRO = $(shell uname)
CURRENT_TAG =$(shell git describe --abbrev=0 --tags)

GO  = GO111MODULE=on go
## make fabric1.4: build fabric(1.4) client plugin
fabric1.4:
	mkdir -p build
	$(GO) build -o build/fabric-client-1.4 ./*.go