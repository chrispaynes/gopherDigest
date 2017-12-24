src = ./src
main = $(src)/main.go
pkgDir = $(src)/$(pkg)

.PHONY: clean build install start test

build:
	@go build $(main)

dockbuild:
	docker build .

dockUp: dockDown
	docker-compose up

dockDown:
	docker-compose down	

clean:
	@rm -f ./cmd/main

fmt: 
	@go fmt ./...

install:
	go install $(main)

src-package:
	@mkdir -p $(pkgDir)
	@echo package $(pkg) | tee $(pkgDir)/$(pkg).go $(pkgDir)/$(pkg)_test.go

start: clean
	@go run $(main)

test:
	@go test **/*_test.go