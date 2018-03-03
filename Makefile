#!/bin/bash

build:
	go build -o service main.go cache.go file.go fixture.go response.go validate.go

clean:
	rm -rf ./service