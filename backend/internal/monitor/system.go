package monitor

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"port-forward-dashboard/internal/models"
)

type SystemMonitor struct {
	lastNetIn  uint64
	lastNetOut uint64
	lastUpdate time.Time
	netRateIn  float64
	netRateOut float64
	mu         sync.RWMutex
}

func NewSystemMonitor() *SystemMonitor {
	m := &SystemMonitor{
		lastUpdate: time.Now(),
	}

	// 初始化网络计数器
	if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
		m.lastNetIn = counters[0].BytesRecv
		m.lastNetOut = counters[0].BytesSent
	}

	go m.updateLoop()
	return m
}

func (m *SystemMonitor) updateLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.updateNetworkRates()
	}
}

func (m *SystemMonitor) updateNetworkRates() {
	counters, err := net.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	duration := now.Sub(m.lastUpdate).Seconds()
	if duration < 0.1 {
		return
	}

	currentIn := counters[0].BytesRecv
	currentOut := counters[0].BytesSent

	m.netRateIn = float64(currentIn-m.lastNetIn) / duration
	m.netRateOut = float64(currentOut-m.lastNetOut) / duration

	m.lastNetIn = currentIn
	m.lastNetOut = currentOut
	m.lastUpdate = now
}

func (m *SystemMonitor) GetStats(activeTunnels int, uptime int64) models.SystemStats {
	stats := models.SystemStats{
		ActiveTunnels: activeTunnels,
		Uptime:        uptime,
	}

	// CPU
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		stats.CPUPercent = cpuPercent[0]
	}

	// Memory
	if memInfo, err := mem.VirtualMemory(); err == nil {
		stats.MemoryPercent = memInfo.UsedPercent
		stats.MemoryUsed = memInfo.Used
		stats.MemoryTotal = memInfo.Total
	}

	// Network
	if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
		stats.NetBytesIn = counters[0].BytesRecv
		stats.NetBytesOut = counters[0].BytesSent
	}

	m.mu.RLock()
	stats.NetRateIn = m.netRateIn
	stats.NetRateOut = m.netRateOut
	m.mu.RUnlock()

	return stats
}
