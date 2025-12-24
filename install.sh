#!/bin/bash
set -e

# Port Forward Agent 一键安装脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# 打印带颜色的信息
print_banner() {
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${WHITE}${BOLD}🚀 Port Forward Agent 一键安装脚本${NC}                         ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}   ${PURPLE}GitHub: github.com/jiuwovo-ai/tcp-zz${NC}                        ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${BLUE}[${WHITE}${BOLD}$1${NC}${BLUE}]${NC} ${WHITE}$2${NC}"
}

print_info() {
    echo -e "    ${CYAN}➜${NC} $1"
}

print_success() {
    echo -e "    ${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "    ${RED}✗${NC} $1"
}

print_warning() {
    echo -e "    ${YELLOW}!${NC} $1"
}

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

# 显示 Banner
print_banner

# 参数验证
if [ -z "$NODE_KEY" ]; then
    print_error "错误: 必须提供 --key 参数"
    exit 1
fi

if [ -z "$MASTER_URL" ]; then
    print_error "错误: 必须提供 --master 参数"
    exit 1
fi

# 显示配置信息
echo -e "${WHITE}${BOLD}📋 安装配置${NC}"
echo -e "────────────────────────────────────────"
print_info "节点名称: ${GREEN}$NODE_NAME${NC}"
print_info "节点端口: ${GREEN}$NODE_PORT${NC}"
print_info "主控地址: ${GREEN}$MASTER_URL${NC}"
echo ""

# 检测系统架构
print_step "1/5" "检测系统环境"
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) print_error "不支持的架构: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
print_success "系统: ${CYAN}$OS${NC}, 架构: ${CYAN}$ARCH${NC}"
echo ""

# 安装目录
INSTALL_DIR="/opt/port-forward-agent"
mkdir -p $INSTALL_DIR

# 检查 Go 环境
print_step "2/5" "检查 Go 环境"
if ! command -v go &> /dev/null; then
    print_warning "未检测到 Go，正在安装..."
    curl -fsSL "https://go.dev/dl/go1.21.5.linux-$ARCH.tar.gz" 2>/dev/null | tar -C /usr/local -xzf -
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    print_success "Go 安装完成"
else
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go 已安装: ${CYAN}$GO_VERSION${NC}"
fi
echo ""

# 下载源码并编译
print_step "3/5" "下载 Agent 源码"
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

curl -fsSL "https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/agent/main.go" -o main.go 2>/dev/null
curl -fsSL "https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/agent/go.mod" -o go.mod 2>/dev/null
print_success "源码下载完成"
echo ""

print_step "4/5" "编译 Agent"
go mod tidy 2>/dev/null
go build -o $INSTALL_DIR/port-forward-agent . 2>/dev/null
print_success "编译完成"

cd $INSTALL_DIR
rm -rf $TEMP_DIR
echo ""

# 创建 systemd 服务
print_step "5/5" "配置系统服务"
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

systemctl daemon-reload >/dev/null 2>&1
systemctl enable port-forward-agent >/dev/null 2>&1
systemctl start port-forward-agent >/dev/null 2>&1
print_success "服务配置完成"
echo ""

# 检查状态
sleep 2
if systemctl is-active --quiet port-forward-agent; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}   ${WHITE}${BOLD}✅ Port Forward Agent 安装成功！${NC}                            ${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${WHITE}${BOLD}📊 服务状态${NC}"
    echo -e "────────────────────────────────────────"
    print_info "服务状态: ${GREEN}运行中${NC}"
    print_info "监听端口: ${CYAN}$NODE_PORT${NC}"
    print_info "节点名称: ${CYAN}$NODE_NAME${NC}"
    print_info "主控地址: ${CYAN}$MASTER_URL${NC}"
    echo ""
    echo -e "${WHITE}${BOLD}📝 常用命令${NC}"
    echo -e "────────────────────────────────────────"
    print_info "查看状态: ${YELLOW}systemctl status port-forward-agent${NC}"
    print_info "查看日志: ${YELLOW}journalctl -u port-forward-agent -f${NC}"
    print_info "重启服务: ${YELLOW}systemctl restart port-forward-agent${NC}"
    print_info "停止服务: ${YELLOW}systemctl stop port-forward-agent${NC}"
    echo ""
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║${NC}                                                                ${RED}║${NC}"
    echo -e "${RED}║${NC}   ${WHITE}${BOLD}❌ 服务启动失败${NC}                                              ${RED}║${NC}"
    echo -e "${RED}║${NC}                                                                ${RED}║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    print_error "请检查日志: journalctl -u port-forward-agent -n 50"
    exit 1
fi
