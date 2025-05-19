# 定义目录    
PROJECT_DIR			= $(shell pwd)
CURRENT_DIR			= $(shell pwd)
API_PROTO_DIR		= $(PROJECT_DIR)/api

# 定义变量
GOPATH				= $(shell go env GOPATH)
VERSION				= $(shell git describe --tags --always)

SERVICE=gil_teacher
ARGS=$(args)
DS := /

# 搜索proto文件
API_PROTO_FILES		= $(shell find $(API_PROTO_DIR) -name *.proto)
INTERNAL_PROTO_FILES= $(shell find internal -name *.proto)

ENV=$(shell if [ ${PROJECT_ENV} ]; then echo ${PROJECT_ENV}; else echo "local"; fi)

export GO111MODULE=on
#export GOSUMDB=off

.PHONY: info1
info1:
	@echo PROJECT_DIR		: $(PROJECT_DIR)
	@echo CURRENT_DIR		: $(CURRENT_DIR)
	@echo API_PROTO_DIR		: $(API_PROTO_DIR)
	@echo API_PROTO_FILES	: $(API_PROTO_FILES)

info2:
	@echo INTERNAL_PROTO_FILES: $(INTERNAL_PROTO_FILES)
	@echo GOPATH		: $(GOPATH)
	@echo VERSION		: $(VERSION)


.PHONY: init
# init env
init:
	@echo ------------------------------- init -------------------------------
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install golang.org/x/tools/cmd/stringer@latest

clean:
	rm -rf ./bin/*
	rm -rf $(PROJECT_DIR)/third_party/proto-repo/gil_teacher/
	git rm -f --cached third_party/aio_proto-repo

.PHONY: config
# generate app proto
config:
	protoc --proto_path=. \
		   --proto_path=$(PROJECT_DIR)/third_party \
 	       --go_out=paths=source_relative:. \
	       $(INTERNAL_PROTO_FILES)


.PHONY: generate
# generate
generate:
	go mod tidy
	go mod vendor
	go generate ./...

.PHONY: gen
# generate
gen:
	go generate ./...

.PHONY: proto
# generate
proto:
	./proto/generate.sh

.PHONY: wire
# generate
wire:
	go install github.com/google/wire/cmd/wire@latest
	cd main/gil_teacher/ && wire
	cd main/gil_teacher_consumer/ && wire

.PHONY: all
# generate all
all:
	-make generate;
	-make proto;
.PHONY: build
# build x86_64
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/teacher-api ./main/gil_teacher
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/teacher-consumer ./main/gil_teacher_consumer
#	mkdir -p bin/ && go build -o ./bin/ ./script
# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

#################################################### test工具 ##############################################################
.PHONY: runP
# runTest
runP:
	@make build
	@bin$(DS)$(p)

# grpcui
grpcui:
	@go install github.com/fullstorydev/grpcui/cmd/grpcui@v1.1.0
	@grpcui -plaintext localhost:9000
#make run s=task args="test hi -r=a"
#make run s=ucenter
#make run s=daemon
run:
	go run $(CURRENT_DIR)/main/$(SERVICE) --conf $(CURRENT_DIR)/configs/$(ENV)/api/ $(ARGS)

.PHONY: r
r:
	-make wire;
	-make build;

.PHONY: http
http:
	go run $(CURRENT_DIR)/main/$(SERVICE) --conf $(CURRENT_DIR)/configs/$(ENV)/api/ $(ARGS) --mode 1

.PHONY: grpc
grpc:
	go run $(CURRENT_DIR)/main/$(SERVICE) --conf $(CURRENT_DIR)/configs/$(ENV)/api/ $(ARGS) --mode 2