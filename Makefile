go_version = go1.6.2.linux-amd64

go:

	@echo "\033[1mInstall Go compiler\033[0m"

	@[ -d env ] || mkdir env

	@if [ ! -d env/go ]; then \
		cd env && \
		wget https://storage.googleapis.com/golang/$(go_version).tar.gz && \
		tar -xf ./$(go_version).tar.gz && \
		mkdir gopath && \
		rm ./$(go_version).tar.gz ; \
	fi

	@GOROOT=$(shell pwd)/env/go \
	GOPATH=$(shell pwd)/env/gopath \
	env/go/bin/go get -u github.com/FiloSottile/gvt

	@GOROOT=$(shell pwd)/env/go \
	GOPATH=$(shell pwd)/env/gopath \
	env/go/bin/go get -u github.com/kshvakov/build-html

	@echo "\033[1mGo compiler installed!\033[0m"

build-hooks:

	@echo "\033[1mBuild hooks\033[0m"

	@rm -rf hooks && mkdir -p hooks/bin && \
		rm -rf env/gopath/src/github.com/postgres-ci/hooks

	@git clone -b master https://github.com/postgres-ci/hooks.git env/gopath/src/github.com/postgres-ci/hooks

	@echo "\033[1mBuild post-commit hook\033[0m"

	@GOROOT=$(shell pwd)/env/go \
	GOPATH=$(shell pwd)/env/gopath \
	env/go/bin/go build -ldflags='-s -w' -o hooks/bin/post-commit \
		env/gopath/src/github.com/postgres-ci/hooks/local/bin/post-commit.go

	@echo "\033[1mBuild post-receive hook\033[0m"

	@GOROOT=$(shell pwd)/env/go \
	GOPATH=$(shell pwd)/env/gopath \
	env/go/bin/go build -ldflags='-s -w' -o hooks/bin/post-receive \
		env/gopath/src/github.com/postgres-ci/hooks/server-side/bin/post-receive.go

	@cp env/gopath/src/github.com/postgres-ci/hooks/local/post-commit.sample hooks
	@cp env/gopath/src/github.com/postgres-ci/hooks/server-side/post-receive.sample hooks

	@echo "\033[1mBuild hooks: done!\033[0m"


all: go build-hooks
