#!/bin/bash

# 启动脚本 for autops 项目

# 设置工作目录为脚本所在目录
cd "$(dirname "$0")"

# 定义应用名称
APP_NAME="autops"
# 定义日志文件路径
LOG_FILE="logs/application.log"
# 定义PID文件路径
PID_FILE="logs/application.pid"

# 确保日志目录存在
mkdir -p logs

# 显示帮助信息
show_help() {
    echo "用法: $0 [命令]"
    echo "命令:"
    echo "  start    启动应用程序"
    echo "  stop     停止应用程序"
    echo "  restart  重启应用程序"
    echo "  status   查看应用程序状态"
    exit 1
}

# 检查应用是否正在运行
is_running() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 $PID 2>/dev/null; then
            return 0
        else
            # PID文件存在但进程不存在，删除PID文件
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# 启动应用程序
start_app() {
    if is_running; then
        echo "应用程序 $APP_NAME 已在运行中！"
        exit 1
    fi

    # 检查可执行文件是否存在
    if [ ! -f "./$APP_NAME" ]; then
        echo "错误: 未找到可执行文件 '$APP_NAME'。请先运行 build.sh 构建项目。"
        exit 1
    fi

    # 打印启动信息
     echo "正在启动 $APP_NAME 项目..."

    # 启动应用程序（后台运行，并将输出重定向到日志文件）
    nohup ./$APP_NAME > "$LOG_FILE" 2>&1 &

    # 记录进程ID
    PID=$!
    echo $PID > "$PID_FILE"

    echo "应用程序已在后台启动，进程ID: $PID"
    echo "日志文件: $LOG_FILE"

    # 检查启动是否成功
    if [ $? -eq 0 ]; then
        echo "应用程序启动命令已发送成功！"
    else
        echo "应用程序启动失败！"
        rm -f "$PID_FILE"
        exit 1
    fi
}

# 停止应用程序
stop_app() {
    if ! is_running; then
        echo "应用程序 $APP_NAME 未运行！"
        exit 1
    fi

    PID=$(cat "$PID_FILE")
    echo "正在停止应用程序 $APP_NAME (进程ID: $PID)..."

    # 尝试优雅终止进程
    kill $PID
    sleep 2

    # 如果进程仍在运行，强制终止
    if kill -0 $PID 2>/dev/null; then
        echo "进程未响应，强制终止..."
        kill -9 $PID
    fi

    # 删除PID文件
    rm -f "$PID_FILE"
    echo "应用程序已停止！"
}

# 查看应用程序状态
status_app() {
    if is_running; then
        PID=$(cat "$PID_FILE")
        echo "应用程序 $APP_NAME 正在运行中 (进程ID: $PID)"
        echo "日志文件: $LOG_FILE"
    else
        echo "应用程序 $APP_NAME 未运行"
    fi
}

# 解析命令行参数
if [ $# -ne 1 ]; then
    show_help
fi

case $1 in
    start)
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        stop_app
        start_app
        ;;
    status)
        status_app
        ;;
    *)
        echo "错误: 无效的命令 '$1'"
        show_help
        ;;
esac