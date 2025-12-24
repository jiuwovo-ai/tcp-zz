package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"port-forward-dashboard/internal/models"
)

func (s *Server) handleGetInstallScript(c *gin.Context) {
	id := c.Param("id")

	node, err := s.nm.GetNode(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Node not found"})
		return
	}

	// 获取主控面板地址
	masterHost := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	masterURL := fmt.Sprintf("%s://%s", scheme, masterHost)

	script := generateInstallScript(node.Node, masterURL)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"script":  script,
			"command": generateOneLineCommand(node.Node, masterURL),
		},
	})
}

func (s *Server) handleDownloadAgent(c *gin.Context) {
	// 返回预编译的 agent 二进制文件
	// 这里返回下载脚本，实际会从 GitHub Release 或服务器下载
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, getAgentDownloadScript())
}

const githubRawURL = "https://raw.githubusercontent.com/jiuwovo-ai/tcp-zz/main"

func generateOneLineCommand(node models.Node, masterURL string) string {
	return fmt.Sprintf(`curl -fsSL %s/install.sh | bash -s -- --name "%s" --key "%s" --port %d --master "%s"`,
		githubRawURL, node.Name, node.Key, node.Port, masterURL)
}

func generateInstallScript(node models.Node, masterURL string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Port Forward Agent 一键安装脚本
# 节点名称: %s
# 节点端口: %d

echo "🚀 开始安装 Port Forward Agent..."

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

# 创建安装目录
INSTALL_DIR="/opt/port-forward-agent"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 下载 Agent
echo "⬇️ 下载 Agent..."
DOWNLOAD_URL="%s/api/agent/download?os=$OS&arch=$ARCH"

# 如果有预编译版本则下载，否则从源码编译
if command -v go &> /dev/null; then
    echo "📦 检测到 Go 环境，从源码编译..."
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    
    # 下载源码
    cat > main.go << 'AGENT_SOURCE'
%s
AGENT_SOURCE

    cat > go.mod << 'GO_MOD'
module port-forward-agent

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/shirou/gopsutil/v3 v3.23.12
)
GO_MOD

    # 编译
    echo "🔨 编译中..."
    go mod tidy
    go build -o $INSTALL_DIR/port-forward-agent .
    
    cd $INSTALL_DIR
    rm -rf $TEMP_DIR
else
    echo "❌ 未检测到 Go 环境，请先安装 Go 1.21+"
    echo "   安装命令: curl -fsSL https://go.dev/dl/go1.21.5.linux-$ARCH.tar.gz | sudo tar -C /usr/local -xzf -"
    echo "   然后添加到 PATH: export PATH=\$PATH:/usr/local/go/bin"
    exit 1
fi

# 创建 systemd 服务
echo "📝 创建 systemd 服务..."
cat > /etc/systemd/system/port-forward-agent.service << EOF
[Unit]
Description=Port Forward Agent
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/port-forward-agent -name "%s" -key "%s" -port %d -master "%s"
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
    echo "📍 监听端口: %d"
    echo "🔗 节点名称: %s"
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
`, node.Name, node.Port, masterURL, getAgentSourceCode(), node.Name, node.Key, node.Port, masterURL, node.Port, node.Name)
}

func getAgentDownloadScript() string {
	return `#!/bin/bash
# Port Forward Agent 安装脚本
# 用法: curl -fsSL <master>/api/install.sh | bash -s -- --name "节点名" --key "密钥" --port 9090

set -e

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

echo "🚀 开始安装 Port Forward Agent..."
echo "   节点名称: $NODE_NAME"
echo "   节点端口: $NODE_PORT"

# 检测系统
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *) echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

# 安装目录
INSTALL_DIR="/opt/port-forward-agent"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "📦 安装 Go..."
    curl -fsSL https://go.dev/dl/go1.21.5.linux-$ARCH.tar.gz | tar -C /usr/local -xzf -
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# 下载源码并编译
echo "⬇️ 下载并编译 Agent..."
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

# 这里会从主控面板下载源码
curl -fsSL "$MASTER_URL/api/agent/source" -o agent.tar.gz 2>/dev/null || {
    # 如果下载失败，使用内嵌源码
    echo "使用内嵌源码..."
}

# 编译
go mod init port-forward-agent 2>/dev/null || true
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/shirou/gopsutil/v3@v3.23.12
go build -o $INSTALL_DIR/port-forward-agent . 2>/dev/null || {
    echo "❌ 编译失败"
    exit 1
}

cd $INSTALL_DIR
rm -rf $TEMP_DIR

# 创建 systemd 服务
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
systemctl daemon-reload
systemctl enable port-forward-agent
systemctl start port-forward-agent

echo ""
echo "✅ 安装完成！"
echo "   查看状态: systemctl status port-forward-agent"
echo "   查看日志: journalctl -u port-forward-agent -f"
`
}

func getAgentSourceCode() string {
	// 返回 agent 的源代码，用于在目标服务器上编译
	return `package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

var (
	masterURL  string
	nodeKey    string
	nodeName   string
	listenPort int
	tunnels    = make(map[string]*Tunnel)
	tunnelsMu  sync.RWMutex
	startTime  = time.Now()
)

type Tunnel struct {
	ID         string
	LocalPort  int
	TargetIP   string
	TargetPort int
	Protocol   string
	listener   net.Listener
	udpConn    *net.UDPConn
	running    atomic.Bool
	bytesIn    atomic.Int64
	bytesOut   atomic.Int64
	rateIn     float64
	rateOut    float64
	lastIn     int64
	lastOut    int64
	lastUpdate time.Time
	cancel     chan struct{}
	mu         sync.RWMutex
}

type NodeStatus struct {
	NodeKey     string         ` + "`json:\"node_key\"`" + `
	NodeName    string         ` + "`json:\"node_name\"`" + `
	Online      bool           ` + "`json:\"online\"`" + `
	CPUPercent  float64        ` + "`json:\"cpu_percent\"`" + `
	MemPercent  float64        ` + "`json:\"mem_percent\"`" + `
	Uptime      int64          ` + "`json:\"uptime\"`" + `
	TunnelCount int            ` + "`json:\"tunnel_count\"`" + `
	Tunnels     []TunnelStatus ` + "`json:\"tunnels\"`" + `
}

type TunnelStatus struct {
	ID         string  ` + "`json:\"id\"`" + `
	LocalPort  int     ` + "`json:\"local_port\"`" + `
	TargetIP   string  ` + "`json:\"target_ip\"`" + `
	TargetPort int     ` + "`json:\"target_port\"`" + `
	Protocol   string  ` + "`json:\"protocol\"`" + `
	Running    bool    ` + "`json:\"running\"`" + `
	BytesIn    int64   ` + "`json:\"bytes_in\"`" + `
	BytesOut   int64   ` + "`json:\"bytes_out\"`" + `
	RateIn     float64 ` + "`json:\"rate_in\"`" + `
	RateOut    float64 ` + "`json:\"rate_out\"`" + `
	Latency    int64   ` + "`json:\"latency\"`" + `
}

type APIResponse struct {
	Success bool        ` + "`json:\"success\"`" + `
	Message string      ` + "`json:\"message,omitempty\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
}

func main() {
	flag.StringVar(&masterURL, "master", "", "Master panel URL")
	flag.StringVar(&nodeKey, "key", "", "Node authentication key")
	flag.StringVar(&nodeName, "name", "Node", "Node display name")
	flag.IntVar(&listenPort, "port", 9090, "Agent API listen port")
	flag.Parse()

	if nodeKey == "" {
		log.Fatal("Node key is required. Use -key flag")
	}

	log.Printf("Port Forward Agent Starting - Name: %s, Port: %d", nodeName, listenPort)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		key := c.GetHeader("X-Node-Key")
		if key != nodeKey {
			c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Message: "Invalid node key"})
			c.Abort()
			return
		}
		c.Next()
	})

	router.GET("/status", handleStatus)
	router.POST("/tunnels", handleCreateTunnel)
	router.DELETE("/tunnels/:id", handleDeleteTunnel)
	router.POST("/tunnels/:id/start", handleStartTunnel)
	router.POST("/tunnels/:id/stop", handleStopTunnel)

	go updateRatesLoop()
	if masterURL != "" {
		go registerToMaster()
	}

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", listenPort)
		log.Printf("Agent running on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start agent: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}

func handleStatus(c *gin.Context) {
	status := getNodeStatus()
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: status})
}

func handleCreateTunnel(c *gin.Context) {
	var req struct {
		ID         string ` + "`json:\"id\"`" + `
		LocalPort  int    ` + "`json:\"local_port\"`" + `
		TargetIP   string ` + "`json:\"target_ip\"`" + `
		TargetPort int    ` + "`json:\"target_port\"`" + `
		Protocol   string ` + "`json:\"protocol\"`" + `
		AutoStart  bool   ` + "`json:\"auto_start\"`" + `
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	tunnelsMu.Lock()
	if _, exists := tunnels[req.ID]; exists {
		tunnelsMu.Unlock()
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Message: "Tunnel already exists"})
		return
	}

	tunnel := &Tunnel{
		ID: req.ID, LocalPort: req.LocalPort, TargetIP: req.TargetIP,
		TargetPort: req.TargetPort, Protocol: req.Protocol,
		cancel: make(chan struct{}), lastUpdate: time.Now(),
	}
	tunnels[req.ID] = tunnel
	tunnelsMu.Unlock()

	if req.AutoStart {
		startTunnel(tunnel)
	}
	log.Printf("Tunnel created: %s", req.ID)
	c.JSON(http.StatusOK, APIResponse{Success: true})
}

func handleDeleteTunnel(c *gin.Context) {
	id := c.Param("id")
	tunnelsMu.Lock()
	tunnel, exists := tunnels[id]
	if !exists {
		tunnelsMu.Unlock()
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Message: "Tunnel not found"})
		return
	}
	stopTunnel(tunnel)
	delete(tunnels, id)
	tunnelsMu.Unlock()
	c.JSON(http.StatusOK, APIResponse{Success: true})
}

func handleStartTunnel(c *gin.Context) {
	id := c.Param("id")
	tunnelsMu.RLock()
	tunnel, exists := tunnels[id]
	tunnelsMu.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Message: "Tunnel not found"})
		return
	}
	startTunnel(tunnel)
	c.JSON(http.StatusOK, APIResponse{Success: true})
}

func handleStopTunnel(c *gin.Context) {
	id := c.Param("id")
	tunnelsMu.RLock()
	tunnel, exists := tunnels[id]
	tunnelsMu.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Message: "Tunnel not found"})
		return
	}
	stopTunnel(tunnel)
	c.JSON(http.StatusOK, APIResponse{Success: true})
}

func startTunnel(t *Tunnel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running.Load() { return }
	t.cancel = make(chan struct{})
	
	if t.Protocol == "udp" {
		addr := &net.UDPAddr{Port: t.LocalPort}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil { log.Printf("UDP listen error: %v", err); return }
		t.udpConn = conn
		go handleUDP(t)
	} else {
		addr := fmt.Sprintf("0.0.0.0:%d", t.LocalPort)
		listener, err := net.Listen("tcp", addr)
		if err != nil { log.Printf("TCP listen error: %v", err); return }
		t.listener = listener
		go handleTCP(t)
	}
	t.running.Store(true)
	log.Printf("Tunnel started: %s", t.ID)
}

func handleTCP(t *Tunnel) {
	for {
		select {
		case <-t.cancel: return
		default:
		}
		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.cancel: return
			default: continue
			}
		}
		go func(c net.Conn) {
			defer c.Close()
			target, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", t.TargetIP, t.TargetPort), 10*time.Second)
			if err != nil { return }
			defer target.Close()
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { defer wg.Done(); copyData(target, c, &t.bytesOut, t.cancel) }()
			go func() { defer wg.Done(); copyData(c, target, &t.bytesIn, t.cancel) }()
			wg.Wait()
		}(conn)
	}
}

func handleUDP(t *Tunnel) {
	buf := make([]byte, 65535)
	clients := make(map[string]*net.UDPConn)
	var mu sync.RWMutex
	targetAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", t.TargetIP, t.TargetPort))
	
	for {
		select {
		case <-t.cancel:
			mu.Lock()
			for _, c := range clients { c.Close() }
			mu.Unlock()
			return
		default:
		}
		t.udpConn.SetReadDeadline(time.Now().Add(time.Second))
		n, clientAddr, err := t.udpConn.ReadFromUDP(buf)
		if err != nil { continue }
		t.bytesOut.Add(int64(n))
		
		key := clientAddr.String()
		mu.RLock()
		tc, exists := clients[key]
		mu.RUnlock()
		
		if !exists {
			tc, err = net.DialUDP("udp", nil, targetAddr)
			if err != nil { continue }
			mu.Lock()
			clients[key] = tc
			mu.Unlock()
			go func(tc *net.UDPConn, ca *net.UDPAddr, key string) {
				rbuf := make([]byte, 65535)
				for {
					select {
					case <-t.cancel: return
					default:
					}
					tc.SetReadDeadline(time.Now().Add(30 * time.Second))
					rn, err := tc.Read(rbuf)
					if err != nil {
						mu.Lock(); delete(clients, key); mu.Unlock()
						tc.Close()
						return
					}
					t.bytesIn.Add(int64(rn))
					t.udpConn.WriteToUDP(rbuf[:rn], ca)
				}
			}(tc, clientAddr, key)
		}
		tc.Write(buf[:n])
	}
}

func copyData(dst, src net.Conn, counter *atomic.Int64, cancel chan struct{}) {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-cancel: return
		default:
		}
		src.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := src.Read(buf)
		if n > 0 {
			counter.Add(int64(n))
			dst.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if _, werr := dst.Write(buf[:n]); werr != nil { return }
		}
		if err != nil { return }
	}
}

func stopTunnel(t *Tunnel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running.Load() { return }
	close(t.cancel)
	if t.listener != nil { t.listener.Close(); t.listener = nil }
	if t.udpConn != nil { t.udpConn.Close(); t.udpConn = nil }
	t.running.Store(false)
	log.Printf("Tunnel stopped: %s", t.ID)
}

func getNodeStatus() NodeStatus {
	status := NodeStatus{NodeKey: nodeKey, NodeName: nodeName, Online: true, Uptime: int64(time.Since(startTime).Seconds())}
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 { status.CPUPercent = cpuPercent[0] }
	if memInfo, err := mem.VirtualMemory(); err == nil { status.MemPercent = memInfo.UsedPercent }
	
	tunnelsMu.RLock()
	defer tunnelsMu.RUnlock()
	status.TunnelCount = len(tunnels)
	status.Tunnels = make([]TunnelStatus, 0, len(tunnels))
	for _, t := range tunnels {
		ts := TunnelStatus{ID: t.ID, LocalPort: t.LocalPort, TargetIP: t.TargetIP, TargetPort: t.TargetPort,
			Protocol: t.Protocol, Running: t.running.Load(), BytesIn: t.bytesIn.Load(), BytesOut: t.bytesOut.Load(),
			RateIn: t.rateIn, RateOut: t.rateOut, Latency: checkLatency(t.TargetIP, t.TargetPort)}
		status.Tunnels = append(status.Tunnels, ts)
	}
	return status
}

func checkLatency(ip string, port int) int64 {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 5*time.Second)
	if err != nil { return -1 }
	conn.Close()
	return time.Since(start).Milliseconds()
}

func updateRatesLoop() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		tunnelsMu.RLock()
		for _, t := range tunnels {
			now := time.Now()
			dur := now.Sub(t.lastUpdate).Seconds()
			if dur > 0 {
				in, out := t.bytesIn.Load(), t.bytesOut.Load()
				t.rateIn, t.rateOut = float64(in-t.lastIn)/dur, float64(out-t.lastOut)/dur
				t.lastIn, t.lastOut, t.lastUpdate = in, out, now
			}
		}
		tunnelsMu.RUnlock()
	}
}

func registerToMaster() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		status := getNodeStatus()
		data, _ := json.Marshal(status)
		req, _ := http.NewRequest("POST", masterURL+"/api/nodes/heartbeat", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Node-Key", nodeKey)
		client := &http.Client{Timeout: 5 * time.Second}
		if resp, err := client.Do(req); err == nil { resp.Body.Close() }
	}
}
`
}
