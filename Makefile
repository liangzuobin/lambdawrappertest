TARGET := lambdawrappertest
SHELL := /bin/bash
SAM := $(shell command -v sam 2> /dev/null)
AWS := $(shell command -v aws 2> /dev/null)
STACK_NAME := "lambda-wrapper-test-stack"

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

package:
ifdef SAM
	@sam package --template-file template.yaml --output-template-file deploy.yaml --s3-bucket lambda-func --s3-prefix test
else
	$(error no sam found in path, please install it first.)
endif

cloudformation-deploy:
ifdef AWS
	@aws cloudformation deploy --template-file deploy.yaml --stack-name $(STACK_NAME)
else
	$(error no aws found in path, please install it first.)
endif

deploy: package cloudformation-deploy

.PHONY: all clean main