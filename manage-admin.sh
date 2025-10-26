#!/bin/bash
# manage-admin.sh - 用于在开发环境管理 pulseadmin 服务。

BIN_DIR="./bin"
PROG_NAME="pulseadmin"
LOGS_DIR="./bin/admin/logs"
PID_FILE="./bin/admin/admin.pid"

mkdir -p "$LOGS_DIR"

start() {
  echo "==> 正在启动服务: $PROG_NAME..."
  if [ -f "$PID_FILE" ] && ps -p $(cat "$PID_FILE") > /dev/null; then
    echo "服务已经在运行 (PID: $(cat "$PID_FILE"))."
    return
  fi
  nohup "$BIN_DIR/$PROG_NAME" > "$LOGS_DIR/run.log" 2>&1 &
  local pid=$!
  echo $pid > "$PID_FILE"
  echo "服务启动成功! [PID: $pid]"
}

stop() {
  echo "==> 正在停止服务: $PROG_NAME..."
  if [ ! -f "$PID_FILE" ]; then
    echo "服务未运行 (找不到PID文件)."
    return
  fi
  local pid=$(cat "$PID_FILE")
  if [ -z "$pid" ]; then
    rm -f "$PID_FILE"
    return
  fi
  if ps -p "$pid" > /dev/null; then
    kill "$pid" &>/dev/null
    sleep 1
    if ps -p "$pid" > /dev/null; then kill -9 "$pid" &>/dev/null; fi
    echo "服务已停止。"
  else
    echo "服务之前未在运行。"
  fi
  rm -f "$PID_FILE"
}

case "$1" in
  start|stop|restart)
    if [ "$1" == "restart" ]; then stop; sleep 1; fi
    if [ "$1" != "stop" ]; then start; fi
    ;;
  *)
    echo "用法: $0 {start|stop|restart}"
    exit 1
esac