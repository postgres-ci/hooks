go_version = go1.7.linux-amd64

ifeq ($(shell uname), Darwin)
	go_version = go1.7.darwin-amd64
endif

go:

	@echo "\033[1mInstall Go compiler\033[0m"

	@[ -d env ] || mkdir env

	@if [ ! -d env/go ]; then \
		cd env && \
		curl -O https://storage.googleapis.com/golang/$(go_version).tar.gz && \
		tar -xf ./$(go_version).tar.gz && \
		rm ./$(go_version).tar.gz ; \
	fi

	@echo "\033[1mGo compiler installed!\033[0m"

build:

	@echo "\033[1mBuild hook\033[0m"

	@rm -rf bin && mkdir -p bin && rm -rf env/gopath/src/github.com/postgres-ci/hooks && mkdir -p env/gopath/src/github.com/postgres-ci/hooks

	@cp -r ./git env/gopath/src/github.com/postgres-ci/hooks/

	@GOROOT=$(shell pwd)/env/go \
	GOPATH=$(shell pwd)/env/gopath \
	CGO_ENABLED=0 \
	env/go/bin/go build -ldflags='-s -w' -o bin/postgres-ci-git-hook postgres-ci-git-hook.go

	@echo "\033[1mdone!\033[0m"

	@file bin/postgres-ci-git-hook

install: build

	sudo cp bin/postgres-ci-git-hook /usr/bin/postgres-ci-git-hook

all: go build install
