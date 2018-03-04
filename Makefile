#!/bin/bash

build:
	go build -o service main.go cache.go downloader.go fixture.go imager.go registry.go response.go validate.go

test:
	go test ./... -cover

clean:
	rm -rf ./service
