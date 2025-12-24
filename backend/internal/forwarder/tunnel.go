package forwarder

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"port-forward-dashboard/internal/models"
)

type Tunnel struct {
	rule         models.Rule
	stats        *models.TrafficStats
	latency      *models.LatencyInfo
	running      atomic.Bool
	listener     net.Listener
	udpConn      *net.UDPConn
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	lastBytesIn  int64
	lastBytesOut int64
	lastUpdate   time.Time
}

func NewTunnel(rule models.Rule) *Tunnel {
	return &Tunnel{
		rule:       rule,
		stats:      &models.TrafficStats{},
		latency:    &models.LatencyInfo{Status: "unknown"},
		lastUpdate: time.Now(),
	}
}

func (t *Tunnel) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running.Load() {
		return nil
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())

	var err error
	switch t.rule.Protocol {
	case models.TCP:
		err = t.startTCP()
	case models.UDP:
		err = t.startUDP()
	default:
		err = t.startTCP()
	}

	if err != nil {
		return err
	}

	t.running.Store(true)
	t.rule.Enabled = true

	// 启动延迟检测
	go t.latencyProbe()

	log.Printf("✅ Tunnel %s started: :%d -> %s:%d (%s)",
		t.rule.Name, t.rule.LocalPort, t.rule.TargetIP, t.rule.TargetPort, t.rule.Protocol)

	return nil
}

func (t *Tunnel) startTCP() error {
	addr := fmt.Sprintf("0.0.0.0:%d", t.rule.LocalPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	t.listener = listener

	go t.acceptTCP()
	return nil
}

func (t *Tunnel) acceptTCP() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.ctx.Done():
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		t.stats.Connections.Add(1)
		go t.handleTCPConn(conn)
	}
}

func (t *Tunnel) handleTCPConn(clientConn net.Conn) {
	defer func() {
		clientConn.Close()
		t.stats.Connections.Add(-1)
	}()

	targetAddr := fmt.Sprintf("%s:%d", t.rule.TargetIP, t.rule.TargetPort)
	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", targetAddr, err)
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Target (上行)
	go func() {
		defer wg.Done()
		t.copyWithStats(targetConn, clientConn, &t.stats.BytesOut)
	}()

	// Target -> Client (下行)
	go func() {
		defer wg.Done()
		t.copyWithStats(clientConn, targetConn, &t.stats.BytesIn)
	}()

	wg.Wait()
}

func (t *Tunnel) copyWithStats(dst, src net.Conn, counter *atomic.Int64) {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		src.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := src.Read(buf)
		if n > 0 {
			counter.Add(int64(n))
			dst.SetWriteDeadline(time.Now().Add(30 * time.Second))
			_, werr := dst.Write(buf[:n])
			if werr != nil {
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				select {
				case <-t.ctx.Done():
				default:
				}
			}
			return
		}
	}
}

func (t *Tunnel) startUDP() error {
	addr := &net.UDPAddr{Port: t.rule.LocalPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	t.udpConn = conn

	go t.handleUDP()
	return nil
}

func (t *Tunnel) handleUDP() {
	buf := make([]byte, 65535)
	clients := make(map[string]*net.UDPConn)
	var clientsMu sync.RWMutex

	targetAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", t.rule.TargetIP, t.rule.TargetPort))
	if err != nil {
		log.Printf("Failed to resolve target: %v", err)
		return
	}

	for {
		select {
		case <-t.ctx.Done():
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
			case <-t.ctx.Done():
				return
			default:
				continue
			}
		}

		t.stats.BytesOut.Add(int64(n))

		clientKey := clientAddr.String()
		clientsMu.RLock()
		targetConn, exists := clients[clientKey]
		clientsMu.RUnlock()

		if !exists {
			targetConn, err = net.DialUDP("udp", nil, targetAddr)
			if err != nil {
				log.Printf("Failed to dial target: %v", err)
				continue
			}

			clientsMu.Lock()
			clients[clientKey] = targetConn
			clientsMu.Unlock()

			// 启动反向转发
			go func(tc *net.UDPConn, ca *net.UDPAddr, key string) {
				rbuf := make([]byte, 65535)
				for {
					select {
					case <-t.ctx.Done():
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

					t.stats.BytesIn.Add(int64(rn))
					t.udpConn.WriteToUDP(rbuf[:rn], ca)
				}
			}(targetConn, clientAddr, clientKey)
		}

		targetConn.Write(buf[:n])
	}
}

func (t *Tunnel) latencyProbe() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.checkLatency()
		}
	}
}

func (t *Tunnel) checkLatency() {
	targetAddr := fmt.Sprintf("%s:%d", t.rule.TargetIP, t.rule.TargetPort)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	latency := time.Since(start).Milliseconds()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.latency.LastCheck = time.Now().Unix()

	if err != nil {
		t.latency.Latency = -1
		t.latency.Status = "error"
		return
	}
	conn.Close()

	t.latency.Latency = latency
	if latency < 100 {
		t.latency.Status = "normal"
	} else if latency < 300 {
		t.latency.Status = "warning"
	} else {
		t.latency.Status = "error"
	}
}

func (t *Tunnel) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running.Load() {
		return
	}

	if t.cancel != nil {
		t.cancel()
	}

	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}

	if t.udpConn != nil {
		t.udpConn.Close()
		t.udpConn = nil
	}

	t.running.Store(false)
	t.rule.Enabled = false

	log.Printf("🛑 Tunnel %s stopped", t.rule.Name)
}

func (t *Tunnel) UpdateRule(rule models.Rule) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rule = rule
}

func (t *Tunnel) GetRule() models.Rule {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.rule
}

func (t *Tunnel) IsRunning() bool {
	return t.running.Load()
}

func (t *Tunnel) UpdateRates() {
	now := time.Now()
	duration := now.Sub(t.lastUpdate).Seconds()
	if duration < 0.1 {
		return
	}

	currentIn := t.stats.BytesIn.Load()
	currentOut := t.stats.BytesOut.Load()

	t.stats.BytesInRate = float64(currentIn-t.lastBytesIn) / duration
	t.stats.BytesOutRate = float64(currentOut-t.lastBytesOut) / duration
	t.stats.TotalIn = currentIn
	t.stats.TotalOut = currentOut
	t.stats.ConnCount = t.stats.Connections.Load()

	t.lastBytesIn = currentIn
	t.lastBytesOut = currentOut
	t.lastUpdate = now
}

func (t *Tunnel) GetTrafficStats() models.TrafficStats {
	return models.TrafficStats{
		BytesInRate:  t.stats.BytesInRate,
		BytesOutRate: t.stats.BytesOutRate,
		TotalIn:      t.stats.BytesIn.Load(),
		TotalOut:     t.stats.BytesOut.Load(),
		ConnCount:    t.stats.Connections.Load(),
	}
}

func (t *Tunnel) GetStatus() *models.TunnelStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return &models.TunnelStatus{
		Rule:    t.rule,
		Traffic: t.GetTrafficStats(),
		Latency: *t.latency,
		Running: t.running.Load(),
	}
}
