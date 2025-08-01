#!/bin/bash

# 构建脚本 for autops 项目

# 默认架构
DEFAULT_ARCH="amd64"
# 支持的架构列表
SUPPORTED_ARCHES=("amd64" "arm64" "386")

# 显示帮助信息
show_help() {
    echo "用法: $0 [架构]"
    echo "支持的架构: ${SUPPORTED_ARCHES[*]}"
    echo "默认架构: $DEFAULT_ARCH"
    echo "示例: $0 amd64  # 构建amd64架构"
    exit 1
}

# 解析命令行参数
if [ $# -gt 1 ]; then
    show_help
elif [ $# -eq 1 ]; then
    ARCH=$1
    # 检查架构是否受支持
    if [[ ! "${SUPPORTED_ARCHES[*]}" =~ "$ARCH" ]]; then
        echo "错误: 不支持的架构 '$ARCH'"
        show_help
    fi
else
    ARCH=$DEFAULT_ARCH
fi

# 设置工作目录为脚本所在目录
cd "$(dirname "$0")"

# 打印构建信息
 echo "开始构建 autops 项目 (架构: $ARCH)..."

# 安装依赖
 echo "正在安装依赖..."
 go mod tidy

# 设置GOARCH环境变量
export GOARCH=$ARCH

# 构建项目，根据架构生成不同的可执行文件
OUTPUT_FILE="autops_$ARCH"
 echo "正在构建项目..."
 go build -o "$OUTPUT_FILE" main.go

# 检查构建是否成功
if [ $? -eq 0 ]; then
    # 创建软链接，方便使用
    ln -sf "$OUTPUT_FILE" autops
    echo "构建成功！可执行文件: $OUTPUT_FILE"
    echo "已创建软链接: autops -> $OUTPUT_FILE"
else
    echo "构建失败！"
    exit 1
fi