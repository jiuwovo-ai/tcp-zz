#!/bin/bash
set -e

# Port Forward Agent 一键安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/install.sh | bash -s -- --name "节点名" --key "密钥" --port 9090 --master "http://面板地址:8080"

# 解析参数
NODE_NAME="Node"
NODE_KEY=""
NODE_PORT=9090
MASTER_URL=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --name) NODE_NAME="$2"; shift 2 ;;
        --key) NODE_KEY="$2"; shift 2 ;;
        --port) NODE_PORT="$2"; shift 2 ;;
        --master) MASTER_URL="$2"; shift 2 ;;
        *) shift ;;
    esac
done

if [ -z "$NODE_KEY" ]; then
    echo "❌ 错误: 必须提供 --key 参数"
    exit 1
fi

if [ -z "$MASTER_URL" ]; then
    echo "❌ 错误: 必须提供 --master 参数"
    exit 1
fi

echo "🚀 开始安装 Port Forward Agent..."
echo "   节点名称: $NODE_NAME"
echo "   节点端口: $NODE_PORT"
echo "   主控地址: $MASTER_URL"

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
echo "📦 系统: $OS, 架构: $ARCH"

# 安装目录
INSTALL_DIR="/opt/port-forward-agent"
mkdir -p $INSTALL_DIR

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "📦 安装 Go..."
    curl -fsSL "https://go.dev/dl/go1.21.5.linux-$ARCH.tar.gz" | tar -C /usr/local -xzf -
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# 下载源码并编译
echo "⬇️ 下载 Agent 源码..."
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

# 从 GitHub 下载 agent 源码
curl -fsSL "https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/agent/main.go" -o main.go
curl -fsSL "https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/agent/go.mod" -o go.mod

# 编译
echo "🔨 编译中..."
go mod tidy
go build -o $INSTALL_DIR/port-forward-agent .

cd $INSTALL_DIR
rm -rf $TEMP_DIR

# 创建 systemd 服务
echo "📝 创建 systemd 服务..."
cat > /etc/systemd/system/port-forward-agent.service << EOF
[Unit]
Description=Port Forward Agent
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/port-forward-agent -name "$NODE_NAME" -key "$NODE_KEY" -port $NODE_PORT -master "$MASTER_URL"
Restart=always
RestartSec=5
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo "🚀 启动服务..."
systemctl daemon-reload
systemctl enable port-forward-agent
systemctl start port-forward-agent

# 检查状态
sleep 2
if systemctl is-active --quiet port-forward-agent; then
    echo ""
    echo "✅ Port Forward Agent 安装成功！"
    echo ""
    echo "📊 服务状态: 运行中"
    echo "📍 监听端口: $NODE_PORT"
    echo "🔗 节点名称: $NODE_NAME"
    echo "🌐 主控地址: $MASTER_URL"
    echo ""
    echo "常用命令:"
    echo "  查看状态: systemctl status port-forward-agent"
    echo "  查看日志: journalctl -u port-forward-agent -f"
    echo "  重启服务: systemctl restart port-forward-agent"
    echo "  停止服务: systemctl stop port-forward-agent"
else
    echo "❌ 服务启动失败，请检查日志: journalctl -u port-forward-agent -n 50"
    exit 1
fi
