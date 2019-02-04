TARGET := lambdawrappertest
SHELL := /bin/bash
SAM := $(shell command -v sam 2> /dev/null)

all: main

main: clean build zip

clean:
	@go clean && rm -f $(TARGET).zip

build:
	@env GOOS=linux go build -ldflags '-d -s -w' -o $(TARGET) main.go

zip:
	@zip $(TARGET).zip ./$(TARGET)

sam-local:
ifdef SAM
	@sam local start-api --template template.yaml
else
	$(error no sam found in path, please install it first.)
endif

local: clean build zip sam-local

.PHONY: all clean main