#!/bin/zsh

supported_modes=("single" "cluster")

mode=$1
if [[ ! " ${supported_modes[@]} " =~ " ${mode} " ]]; then
    echo "Usage: $0 [single|cluster]"
    exit 1
fi

if [[ ! -e "sentinel-$mode.yml" ]]; then
    echo "sentinel-$mode.yml not found!"
    exit 1
fi


mkdir -p ./out

# 编译go代码
go build -o out/llm_token_ratelimit_benchmark main.go
# 复制 sentinel 配置文件到输出目录
cp sentinel-$mode.yml out/sentinel.yml