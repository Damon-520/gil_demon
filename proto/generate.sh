#!/bin/bash

set -e

OUTPUT_BASE=proto/gen
OUTPUT_OPENAPI=proto/openapi
OUTPUT_GOLANG=$OUTPUT_BASE/go

EXTRA_DEPS=(
# 其他的依赖继续加,一个行一个.
)


# remove all generated files
rm -rf $OUTPUT_BASE
rm -rf $OUTPUT_OPENAPI
mkdir -p $OUTPUT_GOLANG
mkdir -p $OUTPUT_OPENAPI

# For golang
# make sure tools are installed
# go mod tidy
go install -mod=mod \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

EXTRA_DEPS_PATH=""
for v in "${EXTRA_DEPS[@]}"
do
    dep_path=`go list -f '{{ .Dir }}' -m ${v}`
	echo " -> dep: ${v} add dep path: ${dep_path}"
    EXTRA_DEPS_PATH="$EXTRA_DEPS_PATH:${dep_path}"
done

echo "-----------------------------"
echo "ROOT_PATH:         $ROOT_PATH"
echo "OUTPUT_OPENAPI:    $OUTPUT_OPENAPI"
echo "OUTPUT_GOLANG:     $OUTPUT_GOLANG"
echo "using protoc     : $(protoc --version)"
echo "-----------------------------"

IMPORT_BASE_PATH=`pwd`

# generate golang codes
GOLANG_OUTPUT=gen/go/pb
proto_files=$(find ./proto -iname "*.proto")
for proto_file in $proto_files
do
    echo 'proto_file:' $proto_file
    protoc --proto_path=. \
		   --proto_path=./third_party \
 	       --go_out=./$OUTPUT_GOLANG --go_opt=paths=source_relative \
           --go-grpc_out=./$OUTPUT_GOLANG --go-grpc_opt=paths=source_relative \
           --grpc-gateway_out $OUTPUT_GOLANG \
           --grpc-gateway_opt logtostderr=true \
           --grpc-gateway_opt paths=source_relative \
           --openapiv2_out=./$OUTPUT_OPENAPI --openapiv2_opt logtostderr=true,json_names_for_fields=false \
	       $proto_file
    # protoc \
    #     --proto_path=.
    #     -I./third_party/google/api \
    #     --go_out $OUTPUT_GOLANG --go_opt paths=source_relative \
    #     --go-grpc_out $OUTPUT_GOLANG --go-grpc_opt paths=source_relative \
    #     --grpc-gateway_out $OUTPUT_GOLANG \
    #     --grpc-gateway_opt logtostderr=true \
    #     --grpc-gateway_opt paths=source_relative \
    #     --openapiv2_out $OUTPUT_OPENAPI --openapiv2_opt logtostderr=true,json_names_for_fields=false \
    #     $proto_file
done

# go mod tidy
echo "done"