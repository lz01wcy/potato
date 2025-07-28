#!/bin/bash

# 获取当前脚本所在的目录
SCRIPT_DIR=$(dirname "$0")

# 将当前工作目录切换到脚本所在的目录
cd "$SCRIPT_DIR" || exit

# 生成代码
protoc -I. --go_out=. --go-grain_out=. rpc.proto
protoc -I. --go_out=. --autoregister_out=. nice.proto