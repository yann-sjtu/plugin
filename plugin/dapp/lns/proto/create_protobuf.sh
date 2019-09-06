#!/bin/sh
protoc --go_out=plugins=grpc:../../../../../../../ ./*.proto --proto_path=. --proto_path="../../../../../../../../vendor/github.com/33cn/chain33/types/proto/"
