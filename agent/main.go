package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	ID         string `json:"id"`
	LocalPort  int    `json:"local_port"`
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`

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
	NodeKey     string         `json:"node_key"`
	NodeName    string         `json:"node_name"`
	Online      bool           `json:"online"`
	CPUPercent  float64        `json:"cpu_percent"`
	MemPercent  float64        `json:"mem_percent"`
	NetRateIn   float64        `json:"net_rate_in"`
	NetRateOut  float64        `json:"net_rate_out"`
	Uptime      int64          `json:"uptime"`
	TunnelCount int            `json:"tunnel_count"`
	Tunnels     []TunnelStatus `json:"tunnels"`
}

type TunnelStatus struct {
	ID         string  `json:"id"`
	LocalPort  int     `json:"local_port"`
	TargetIP   string  `json:"target_ip"`
	TargetPort int     `json:"target_port"`
	Protocol   string  `json:"protocol"`
	Running    bool    `json:"running"`
	BytesIn    int64   `json:"bytes_in"`
	BytesOut   int64   `json:"bytes_out"`
	RateIn     float64 `json:"rate_in"`
	RateOut    float64 `json:"rate_out"`
	Latency    int64   `json:"latency"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	flag.StringVar(&masterURL, "master", "", "Master panel URL (e.g., http://192.168.1.1:8080)")
	flag.StringVar(&nodeKey, "key", "", "Node authentication key")
	flag.StringVar(&nodeName, "name", "Node", "Node display name")
	flag.IntVar(&listenPort, "port", 9090, "Agent API listen port")
	flag.Parse()

	if nodeKey == "" {
		log.Fatal("Node key is required. Use -key flag")
	}

	log.Printf("🚀 Port Forward Agent Starting...")
	log.Printf("   Node Name: %s", nodeName)
	log.Printf("   Node Key: %s", nodeKey[:8]+"...")
	log.Printf("   Listen Port: %d", listenPort)

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
	router.POST("/uninstall", handleUninstall)

	go updateRatesLoop()

	if masterURL != "" {
		go registerToMaster()
	}

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", listenPort)
		log.Printf("✅ Agent running on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start agent: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	stopAllTunnels()
}

func handleStatus(c *gin.Context) {
	status := getNodeStatus()
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: status})
}

func handleCreateTunnel(c *gin.Context) {
	var req struct {
		ID         string `json:"id"`
		LocalPort  int    `json:"local_port"`
		TargetIP   string `json:"target_ip"`
		TargetPort int    `json:"target_port"`
		Protocol   string `json:"protocol"`
		AutoStart  bool   `json:"auto_start"`
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
		ID:         req.ID,
		LocalPort:  req.LocalPort,
		TargetIP:   req.TargetIP,
		TargetPort: req.TargetPort,
		Protocol:   req.Protocol,
		cancel:     make(chan struct{}),
		lastUpdate: time.Now(),
	}
	tunnels[req.ID] = tunnel
	tunnelsMu.Unlock()

	if req.AutoStart {
		if err := startTunnel(tunnel); err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
			return
		}
	}

	log.Printf("✅ Tunnel created: %s (:%d -> %s:%d)", req.ID, req.LocalPort, req.TargetIP, req.TargetPort)
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Tunnel created"})
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

	log.Printf("🗑️ Tunnel deleted: %s", id)
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Tunnel deleted"})
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

	if err := startTunnel(tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Tunnel started"})
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
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Tunnel stopped"})
}

func startTunnel(t *Tunnel) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running.Load() {
		return nil
	}

	t.cancel = make(chan struct{})

	var err error
	if t.Protocol == "udp" {
		err = t.startUDP()
	} else {
		err = t.startTCP()
	}

	if err != nil {
		return err
	}

	t.running.Store(true)
	log.Printf("▶️ Tunnel started: %s", t.ID)
	return nil
}

func (t *Tunnel) startTCP() error {
	addr := fmt.Sprintf("0.0.0.0:%d", t.LocalPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}
	t.listener = listener

	go func() {
		for {
			select {
			case <-t.cancel:
				return
			default:
			}

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-t.cancel:
					return
				default:
					continue
				}
			}

			go t.handleTCPConn(conn)
		}
	}()

	return nil
}

func (t *Tunnel) handleTCPConn(clientConn net.Conn) {
	defer clientConn.Close()

	targetAddr := fmt.Sprintf("%s:%d", t.TargetIP, t.TargetPort)
	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		t.copyWithStats(targetConn, clientConn, &t.bytesOut)
	}()

	go func() {
		defer wg.Done()
		t.copyWithStats(clientConn, targetConn, &t.bytesIn)
	}()

	wg.Wait()
}

func (t *Tunnel) copyWithStats(dst, src net.Conn, counter *atomic.Int64) {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-t.cancel:
			return
		default:
		}

		src.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := src.Read(buf)
		if n > 0 {
			counter.Add(int64(n))
			dst.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (t *Tunnel) startUDP() error {
	addr := &net.UDPAddr{Port: t.LocalPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen UDP on port %d: %v", t.LocalPort, err)
	}
	t.udpConn = conn

	go t.handleUDP()
	return nil
}

func (t *Tunnel) handleUDP() {
	buf := make([]byte, 65535)
	clients := make(map[string]*net.UDPConn)
	var clientsMu sync.RWMutex

	targetAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", t.TargetIP, t.TargetPort))
	if err != nil {
		log.Printf("Failed to resolve target: %v", err)
		return
	}

	for {
		select {
		case <-t.cancel:
			clientsMu.Lock()
			for _, c := range clients {
				c.Close()
			}
			clientsMu.Unlock()
			return
		default:
		}

		t.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, clientAddr, err := t.udpConn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-t.cancel:
				return
			default:
				continue
			}
		}

		t.bytesOut.Add(int64(n))

		clientKey := clientAddr.String()
		clientsMu.RLock()
		targetConn, exists := clients[clientKey]
		clientsMu.RUnlock()

		if !exists {
			targetConn, err = net.DialUDP("udp", nil, targetAddr)
			if err != nil {
				continue
			}

			clientsMu.Lock()
			clients[clientKey] = targetConn
			clientsMu.Unlock()

			go func(tc *net.UDPConn, ca *net.UDPAddr, key string) {
				rbuf := make([]byte, 65535)
				for {
					select {
					case <-t.cancel:
						return
					default:
					}

					tc.SetReadDeadline(time.Now().Add(30 * time.Second))
					rn, err := tc.Read(rbuf)
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							clientsMu.Lock()
							delete(clients, key)
							clientsMu.Unlock()
							tc.Close()
							return
						}
						continue
					}

					t.bytesIn.Add(int64(rn))
					t.udpConn.WriteToUDP(rbuf[:rn], ca)
				}
			}(targetConn, clientAddr, clientKey)
		}

		targetConn.Write(buf[:n])
	}
}

func stopTunnel(t *Tunnel) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running.Load() {
		return
	}

	close(t.cancel)

	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}

	if t.udpConn != nil {
		t.udpConn.Close()
		t.udpConn = nil
	}

	t.running.Store(false)
	log.Printf("⏹️ Tunnel stopped: %s", t.ID)
}

func stopAllTunnels() {
	tunnelsMu.Lock()
	defer tunnelsMu.Unlock()

	for _, t := range tunnels {
		stopTunnel(t)
	}
}

func getNodeStatus() NodeStatus {
	status := NodeStatus{
		NodeKey:  nodeKey,
		NodeName: nodeName,
		Online:   true,
		Uptime:   int64(time.Since(startTime).Seconds()),
	}

	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		status.CPUPercent = cpuPercent[0]
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		status.MemPercent = memInfo.UsedPercent
	}

	if counters, err := psnet.IOCounters(false); err == nil && len(counters) > 0 {
		status.NetRateIn = float64(counters[0].BytesRecv)
		status.NetRateOut = float64(counters[0].BytesSent)
	}

	tunnelsMu.RLock()
	defer tunnelsMu.RUnlock()

	status.TunnelCount = len(tunnels)
	status.Tunnels = make([]TunnelStatus, 0, len(tunnels))

	for _, t := range tunnels {
		ts := TunnelStatus{
			ID:         t.ID,
			LocalPort:  t.LocalPort,
			TargetIP:   t.TargetIP,
			TargetPort: t.TargetPort,
			Protocol:   t.Protocol,
			Running:    t.running.Load(),
			BytesIn:    t.bytesIn.Load(),
			BytesOut:   t.bytesOut.Load(),
			RateIn:     t.rateIn,
			RateOut:    t.rateOut,
			Latency:    checkLatency(t.TargetIP, t.TargetPort),
		}
		status.Tunnels = append(status.Tunnels, ts)
	}

	return status
}

func checkLatency(ip string, port int) int64 {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 5*time.Second)
	if err != nil {
		return -1
	}
	conn.Close()
	return time.Since(start).Milliseconds()
}

func updateRatesLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		tunnelsMu.RLock()
		for _, t := range tunnels {
			now := time.Now()
			duration := now.Sub(t.lastUpdate).Seconds()
			if duration > 0 {
				currentIn := t.bytesIn.Load()
				currentOut := t.bytesOut.Load()
				t.rateIn = float64(currentIn-t.lastIn) / duration
				t.rateOut = float64(currentOut-t.lastOut) / duration
				t.lastIn = currentIn
				t.lastOut = currentOut
				t.lastUpdate = now
			}
		}
		tunnelsMu.RUnlock()
	}
}

func registerToMaster() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		status := getNodeStatus()
		data, _ := json.Marshal(status)

		req, _ := http.NewRequest("POST", masterURL+"/api/nodes/heartbeat", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Node-Key", nodeKey)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
			continue
		}
		resp.Body.Close()
	}
}

func handleUninstall(c *gin.Context) {
	log.Println("🛑 Received uninstall command from master panel")

	// 先返回成功响应
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: "Uninstalling..."})

	// 异步执行卸载
	go func() {
		time.Sleep(500 * time.Millisecond)

		// 停止所有隧道
		stopAllTunnels()

		log.Println("🗑️ Stopping and disabling service...")

		// 停止并禁用 systemd 服务
		exec.Command("systemctl", "stop", "port-forward-agent").Run()
		exec.Command("systemctl", "disable", "port-forward-agent").Run()

		// 删除服务文件
		os.Remove("/etc/systemd/system/port-forward-agent.service")
		exec.Command("systemctl", "daemon-reload").Run()

		// 删除安装目录
		os.RemoveAll("/opt/port-forward-agent")

		log.Println("✅ Uninstall completed, exiting...")

		// 退出进程
		os.Exit(0)
	}()
}
