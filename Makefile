.PHONY: all clean build

GO ?= go

all: build

build: clean
	${GO} build -o bin/ar-logger main.go

clean:
	@rm -rf bin/ar-logger
