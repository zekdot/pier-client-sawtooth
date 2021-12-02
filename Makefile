SHELL := /bin/bash
CURRENT_PATH = $(shell pwd)
DISTRO = $(shell uname)
CURRENT_TAG =$(shell git describe --abbrev=0 --tags)

GO  = GO111MODULE=on go
## make fabric1.4: build fabric(1.4) client plugin
sawtooth:
	mkdir -p build
	$(GO) build --buildmode=plugin -o build/sawtooth.so ./*.go