#!/bin/bash
# manage-nodes.sh - (已更新为最终目录结构)

# --- 配置 ---
BIN_DIR="./bin"
PROG_NAME="pulsenode"

# 【路径更新】: 指向新的本地配置文件和运行时目录
MOCK_CONF_DIR="./_mock_node/conf"
TARGET_CONF_DIR="./bin/node/conf/testing"
ENVIRONMENT="testing"
RUNTIME_DIR="./_mock_node/runtime"
LOGS_DIR="$RUNTIME_DIR/logs"
PID_DIR="$RUNTIME_DIR/pids"
# --- 配置结束 ---

# start, stop, status 函数内容完全不需要修改，因为它们使用的是上面的变量
# ... (函数内容保持不变) ...

# 确保所有需要的目录都存在
mkdir -p "$LOGS_DIR"
mkdir -p "$PID_DIR"
mkdir -p "$TARGET_CONF_DIR"

# 启动所有节点
start() {
  echo "==> 准备启动所有 pulsenode 工作节点..."
  if [ ! -d "$MOCK_CONF_DIR" ]; then echo "错误: 找不到本地配置文件目录 '$MOCK_CONF_DIR'。"; return 1; fi
  echo "正在清理并从 '$MOCK_CONF_DIR' 复制新配置文件到 '$TARGET_CONF_DIR'..."
  rm -f "$TARGET_CONF_DIR"/*
  cp "$MOCK_CONF_DIR"/*.json "$TARGET_CONF_DIR/"
  
  cd "$BIN_DIR" || { echo "错误: 无法进入目录 '$BIN_DIR'"; return 1; }

  echo "==> 正在启动实例 (当前工作目录: $(pwd))..."
  for conf_file in "node/conf/testing/"*.json; do
    [ -f "$conf_file" ] || continue
    local node_name=$(basename "$conf_file" .json)
    local pid_file="../$PID_DIR/$node_name.pid"
    if [ -f "$pid_file" ] && ps -p $(cat "$pid_file") > /dev/null; then echo "节点 '$node_name' 已经在运行。"; continue; fi
    local conf_name=$(basename "$conf_file" .json)
    echo -n "正在启动节点 '$node_name'... "
    nohup "./$PROG_NAME" --environment="$ENVIRONMENT" --configfilename="$conf_name" > "../$LOGS_DIR/$node_name.log" 2>&1 &
    local pid=$!; echo $pid > "$pid_file"; echo "成功! [PID: $pid]"
  done
  cd ..
  echo "==> 所有节点启动完毕。"
}

# 停止所有节点
stop() {
  echo "==> 正在停止所有工作节点..."
  if [ -z "$(ls -A $PID_DIR/*.pid 2>/dev/null)" ]; then echo "没有正在运行的节点。"; else
    for pid_file in "$PID_DIR"/*.pid; do
      [ -f "$pid_file" ] || continue; local node_name=$(basename "$pid_file" .pid); local pid=$(cat "$pid_file")
      if [ -z "$pid" ]; then rm -f "$pid_file"; continue; fi
      if ps -p "$pid" > /dev/null; then echo -n "正在停止节点 '$node_name'... "; kill "$pid" &>/dev/null; sleep 0.5; if ps -p "$pid" > /dev/null; then kill -9 "$pid" &>/dev/null; fi; echo "已停止。"; fi
      rm -f "$pid_file"
    done
  fi
  echo "正在清理模拟配置文件..."; rm -f "$TARGET_CONF_DIR"/*
  echo "==> 所有节点停止完毕。"
}

# status 函数
status() {
  echo "==> 检查工作节点状态..."; if [ -z "$(ls -A $PID_DIR/*.pid 2>/dev/null)" ]; then echo "没有发现正在运行的节点。"; return; fi
  for pid_file in "$PID_DIR"/*.pid; do
    [ -f "$pid_file" ] || continue; local node_name=$(basename "$pid_file" .pid); local pid=$(cat "$pid_file")
    if ps -p "$pid" > /dev/null; then echo "[运行中] 节点 '$node_name' - PID: $pid"; else echo "[已停止] 节点 '$node_name' - (发现残留的PID文件)"; fi
  done
}

# 主逻辑
case "$1" in
  start|stop|status|restart)
    if [ "$1" == "restart" ]; then stop; sleep 1; fi; $1;;
  *)
    echo "用法: $0 {start|stop|status|restart}"; exit 1
esac
exit 0