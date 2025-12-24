# Port Forward Dashboard / 端口转发面板

<p align="center">
  <strong>🚀 高性能端口转发管理面板 | TCP/UDP 流量转发 | 多节点管理 | 实时监控</strong>
</p>

<p align="center">
  <a href="#-一键部署">一键部署</a> •
  <a href="#-功能特性">功能特性</a> •
  <a href="#-截图预览">截图预览</a> •
  <a href="#-技术栈">技术栈</a> •
  <a href="#-快速开始">快速开始</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue">
  <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/Platform-Linux-FCC624?style=flat-square&logo=linux" alt="Platform">
</p>

---

## ✨ 简介

**Port Forward Dashboard** 是一个功能强大的**端口转发管理面板**，支持 **TCP/UDP 协议转发**、**多节点分布式管理**、**实时流量监控**。适用于：

- 🌐 **内网穿透** - 将内网服务暴露到公网
- 🔀 **端口映射** - 灵活的端口转发规则管理
- 📊 **流量监控** - 实时查看转发流量和延迟
- 🖥️ **多节点管理** - 统一管理多台服务器的转发规则
- 🎮 **游戏加速** - 游戏服务器端口转发
- 🛡️ **NAT 转发** - NAT 网络环境下的端口映射

**关键词**: 端口转发, 端口映射, 流量转发, TCP转发, UDP转发, 内网穿透, NAT转发, 端口转发面板, 端口转发工具, 流量监控, 多节点管理, Port Forwarding, Traffic Forwarding

---

## 🚀 一键部署

```bash
bash <(curl -sL https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main/deploy.sh)
```

部署完成后访问 `http://服务器IP:8080`，默认账号 `admin`，密码 `admin123`

## 📐 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Vue 3)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐ │
│  │ Dashboard │  │ Tunnels  │  │  Login   │  │  ECharts 图表    │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────────┘ │
│                         │                                        │
│              WebSocket (实时数据) + REST API                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend (Go + Gin)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   API Server │  │  WebSocket   │  │    Auth (JWT)          │ │
│  │   (REST)     │  │    Hub       │  │                        │ │
│  └──────────────┘  └──────────────┘  └────────────────────────┘ │
│           │                │                    │                │
│           ▼                ▼                    ▼                │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │                   Forwarder Manager                          ││
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐         ││
│  │  │ Tunnel1 │  │ Tunnel2 │  │ Tunnel3 │  │ TunnelN │         ││
│  │  │ TCP/UDP │  │ TCP/UDP │  │ TCP/UDP │  │ TCP/UDP │         ││
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘         ││
│  └──────────────────────────────────────────────────────────────┘│
│           │                                                      │
│           ▼                                                      │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │                   System Monitor                             ││
│  │  CPU | Memory | Network I/O | Latency Probe                  ││
│  └──────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## 📦 技术栈

### 后端
- **Go 1.21+** - 高性能、低内存占用
- **Gin** - 轻量级 HTTP 框架
- **gorilla/websocket** - WebSocket 实时通信
- **gopsutil** - 系统监控
- **JWT** - 身份认证

### 前端
- **Vue 3** - 组合式 API
- **Vite** - 快速构建
- **Naive UI** - 现代化组件库
- **TailwindCSS** - 原子化 CSS
- **ECharts** - 实时图表
- **Pinia** - 状态管理

## 🚀 功能特性

### 端口转发
- ✅ TCP / UDP 协议支持
- ✅ IPv4 支持
- ✅ 动态添加/删除/修改规则
- ✅ 规则启用/停用
- ✅ 配置持久化

### 实时监控
- ✅ 每秒流量速率 (KB/s, MB/s)
- ✅ 总流量统计
- ✅ WebSocket 实时推送
- ✅ ECharts 动态折线图

### 延迟监控
- ✅ TCP 握手延迟检测
- ✅ 5秒间隔自动探测
- ✅ 状态分级显示 (🟢正常 🟡偏高 🔴超时)

### 系统状态
- ✅ CPU 使用率
- ✅ 内存使用率
- ✅ 网络吞吐量
- ✅ 运行时间

### 安全
- ✅ 账号密码登录
- ✅ JWT Token 认证
- ✅ 24小时 Token 过期

## 🛠️ 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- npm 或 pnpm

### 开发模式

```bash
# 1. 克隆项目后进入目录
cd port-forward-dashboard

# 2. 安装后端依赖
cd backend
go mod tidy

# 3. 安装前端依赖
cd ../frontend
npm install

# 4. 启动后端 (新终端)
cd backend
go run .

# 5. 启动前端 (新终端)
cd frontend
npm run dev
```

访问 http://localhost:3000

**默认账号**: admin / admin123

### 生产构建

```bash
# 构建前端
cd frontend
npm run build

# 构建后端
cd ../backend
go build -o port-forward-dashboard .

# 运行
./port-forward-dashboard
```

访问 http://localhost:8080

## 🐳 Docker 部署

```bash
# 构建镜像
docker build -t port-forward-dashboard .

# 运行容器
docker run -d \
  --name port-forward-dashboard \
  --network host \
  -v $(pwd)/config.json:/app/config.json \
  port-forward-dashboard
```

或使用 docker-compose:

```bash
docker-compose up -d
```

## 🔧 Systemd 部署

```bash
# 构建
make build

# 安装 (需要 root)
sudo make install

# 启动服务
sudo systemctl start port-forward-dashboard

# 查看状态
sudo systemctl status port-forward-dashboard

# 查看日志
journalctl -u port-forward-dashboard -f
```

## ⚙️ 配置文件

`config.json` 示例:

```json
{
  "port": 8080,
  "username": "admin",
  "password": "your-secure-password",
  "jwt_secret": "your-jwt-secret-change-this",
  "rules": []
}
```

**重要**: 生产环境请修改默认密码和 JWT 密钥！

## 📁 项目结构

```
port-forward-dashboard/
├── backend/
│   ├── main.go                 # 入口
│   ├── go.mod
│   └── internal/
│       ├── api/
│       │   ├── server.go       # HTTP 服务器
│       │   ├── auth.go         # JWT 认证
│       │   └── websocket.go    # WebSocket
│       ├── config/
│       │   └── config.go       # 配置管理
│       ├── forwarder/
│       │   ├── manager.go      # 转发管理器
│       │   └── tunnel.go       # 隧道实现
│       ├── models/
│       │   └── models.go       # 数据模型
│       └── monitor/
│           └── system.go       # 系统监控
├── frontend/
│   ├── src/
│   │   ├── api/                # API 封装
│   │   ├── composables/        # 组合式函数
│   │   ├── router/             # 路由
│   │   ├── stores/             # Pinia 状态
│   │   ├── utils/              # 工具函数
│   │   ├── views/              # 页面组件
│   │   ├── App.vue
│   │   └── main.js
│   ├── package.json
│   └── vite.config.js
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## 🔌 API 接口

### 认证
- `POST /api/login` - 登录

### 规则管理
- `GET /api/rules` - 获取所有规则
- `POST /api/rules` - 创建规则
- `PUT /api/rules/:id` - 更新规则
- `DELETE /api/rules/:id` - 删除规则
- `POST /api/rules/:id/toggle` - 启用/停用

### 监控
- `GET /api/dashboard` - 获取仪表板数据
- `GET /api/system` - 获取系统状态
- `GET /api/ws` - WebSocket 连接

## 🧠 核心实现说明

### 端口转发引擎
- 使用 Go 原生 `net` 包实现 TCP/UDP 转发
- 每个隧道独立 goroutine 管理
- 使用 `atomic` 包实现无锁流量统计
- 支持优雅关闭

### 流量统计
- 使用 `sync/atomic` 原子操作计数
- 每秒计算速率差值
- WebSocket 广播到所有客户端

### 延迟探测
- 每 5 秒 TCP 握手测试
- 超时设置 5 秒
- 状态分级: <100ms 正常, <300ms 偏高, >=300ms 或超时为异常

## 🔮 扩展建议

1. **流量限速**: 在 `copyWithStats` 中添加令牌桶限速
2. **告警通知**: 集成 Webhook/邮件通知
3. **数据持久化**: 使用 SQLite/PostgreSQL 存储历史数据
4. **HTTPS**: 配置 TLS 证书

---

## 🌟 Star History

如果这个项目对你有帮助，请给一个 ⭐ Star 支持一下！

---

## � 相关项目

- [frp](https://github.com/fatedier/frp) - 内网穿透工具
- [nps](https://github.com/ehang-io/nps) - 内网穿透代理服务器
- [gost](https://github.com/ginuerzh/gost) - GO 语言实现的安全隧道

---

## �📄 License

MIT License

---

<p align="center">
  <sub>Made with ❤️ by <a href="https://github.com/jiuwovo-ai">jiuwovo-ai</a></sub>
</p>

<!-- 
SEO Keywords / 搜索关键词:
端口转发, 端口转发面板, 端口映射, 流量转发, TCP转发, UDP转发, 
内网穿透, NAT转发, 端口转发工具, 流量监控, 多节点管理,
port forwarding, port forward panel, traffic forwarding, 
tcp forward, udp forward, nat forwarding, tunnel management,
端口转发管理, 服务器端口转发, Linux端口转发, 端口代理,
port proxy, network forwarding, 流量中转, 中转面板
-->
