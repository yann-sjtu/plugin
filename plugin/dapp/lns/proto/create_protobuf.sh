#!/bin/sh
chain33_path=$(go list -f '{{.Dir}}' "github.com/33cn/chain33")
protoc --go_out=plugins=grpc:../../../../../../../ ./*.proto --proto_path=. --proto_path="${chain33_path}/types/proto/"
