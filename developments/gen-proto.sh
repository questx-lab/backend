#!/bin/sh

#* variables
PROTO_PATH=/api/proto
PROTO_OUT=/idl/pb
IDL_PATH=/idl
DOC_OUT=/docs

#! create nat and cobra folder
mkdir -p ${DOC_OUT}/swagger

#* gen normal proto
protoc \
	${PROTO_PATH}/*.proto \
	-I=/usr/local/include \
	--proto_path=${PROTO_PATH} \
	--go_out=:${IDL_PATH} \
	--validate_out=lang=go:${IDL_PATH} \
	--go-grpc_out=:${IDL_PATH} \
	--grpc-gateway_out=:${IDL_PATH} \
	--openapiv2_out=:${DOC_OUT}/swagger
