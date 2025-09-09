#!/bin/bash

# 获取当前脚本所在的目录
SCRIPT_DIR=$(dirname "$0")

# go path
GOBIN=$(go env GOPATH)/bin

# 将当前工作目录切换到脚本所在的目录
cd "$SCRIPT_DIR" || exit

# 生成代码
protoc -I. --go_out=. \
  --go-vtproto_out=. \
  --go-vtproto_opt=features=marshal+unmarshal+size \
  --autoregisterpair_out=. pair.proto