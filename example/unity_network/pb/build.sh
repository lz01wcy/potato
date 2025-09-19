#!/bin/bash

# 获取当前脚本所在的目录
SCRIPT_DIR=$(dirname "$0")

# 将当前工作目录切换到脚本所在的目录
cd "$SCRIPT_DIR" || exit

# 生成描述文件
protoc -I. --descriptor_set_out=msg.desc game.proto

# 生成c#代码
protoc -I. --csharp_out=./cs *.proto

# 生成客户端id注册
go run gen_reg_cs.go

# 复制到客户端
cp -r ./cs/* ../Network/Pb