#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 打印带颜色的消息
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 打印 Banner
print_banner() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                                                           ║"
    echo "║        Port Forward Dashboard - 一键部署脚本              ║"
    echo "║                                                           ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

# 检测系统类型
detect_os() {
    if [ -f /etc/debian_version ]; then
        OS="debian"
        PKG_MANAGER="apt"
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
        PKG_MANAGER="yum"
    else
        print_error "不支持的操作系统"
        exit 1
    fi
    print_info "检测到系统: $OS"
}

# 安装依赖
install_dependencies() {
    print_info "正在安装依赖..."
    
    if [ "$PKG_MANAGER" = "apt" ]; then
        apt update -y
        apt install -y git curl wget
        
        # 安装 Node.js 18+
        if ! command -v node &> /dev/null || [ $(node -v | cut -d'.' -f1 | tr -d 'v') -lt 18 ]; then
            print_info "安装 Node.js 18..."
            curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
            apt install -y nodejs
        fi
        
        # 安装 Go 1.21+
        if ! command -v go &> /dev/null; then
            print_info "安装 Go 1.21..."
            wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
            rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
            rm go1.21.5.linux-amd64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        fi
    else
        yum update -y
        yum install -y git curl wget
        
        # 安装 Node.js
        curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
        yum install -y nodejs
        
        # 安装 Go
        if ! command -v go &> /dev/null; then
            wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
            rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
            rm go1.21.5.linux-amd64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        fi
    fi
    
    print_success "依赖安装完成"
}

# 克隆或更新仓库
clone_or_update_repo() {
    INSTALL_DIR="/opt/port-forward-dashboard"
    REPO_URL="https://github.com/jiuwovo-ai/tcp-zz.git"
    
    if [ -d "$INSTALL_DIR" ]; then
        print_info "检测到已有安装，正在更新..."
        cd "$INSTALL_DIR"
        git pull origin main
    else
        print_info "正在克隆仓库..."
        git clone "$REPO_URL" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi
    
    print_success "代码获取完成"
}

# 构建前端
build_frontend() {
    print_info "正在构建前端..."
    cd /opt/port-forward-dashboard/frontend
    npm install
    npm run build
    print_success "前端构建完成"
}

# 构建后端
build_backend() {
    print_info "正在构建后端..."
    cd /opt/port-forward-dashboard/backend
    export PATH=$PATH:/usr/local/go/bin
    go build -o port-forward-panel .
    print_success "后端构建完成"
}

# 创建 systemd 服务
create_service() {
    print_info "正在创建系统服务..."
    
    cat > /etc/systemd/system/port-forward-panel.service << EOF
[Unit]
Description=Port Forward Dashboard Panel
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/port-forward-dashboard/backend
ExecStart=/opt/port-forward-dashboard/backend/port-forward-panel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable port-forward-panel
    systemctl restart port-forward-panel
    
    print_success "服务创建完成"
}

# 显示完成信息
show_completion() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    部署完成！                             ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # 获取服务器 IP
    SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
    
    echo -e "${CYAN}访问地址:${NC} http://${SERVER_IP}:8080"
    echo -e "${CYAN}默认账号:${NC} admin"
    echo -e "${CYAN}默认密码:${NC} admin123"
    echo ""
    echo -e "${YELLOW}管理命令:${NC}"
    echo "  启动服务: systemctl start port-forward-panel"
    echo "  停止服务: systemctl stop port-forward-panel"
    echo "  重启服务: systemctl restart port-forward-panel"
    echo "  查看状态: systemctl status port-forward-panel"
    echo "  查看日志: journalctl -u port-forward-panel -f"
    echo ""
}

# 主函数
main() {
    print_banner
    check_root
    detect_os
    install_dependencies
    clone_or_update_repo
    build_frontend
    build_backend
    create_service
    show_completion
}

main "$@"
