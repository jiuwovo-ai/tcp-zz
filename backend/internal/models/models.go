package models

import (
	"sync/atomic"
)

type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

type Rule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	LocalPort  int      `json:"local_port"`
	TargetIP   string   `json:"target_ip"`
	TargetPort int      `json:"target_port"`
	Protocol   Protocol `json:"protocol"`
	Enabled    bool     `json:"enabled"`
	CreatedAt  int64    `json:"created_at"`
}

type TrafficStats struct {
	BytesIn      atomic.Int64 `json:"-"`
	BytesOut     atomic.Int64 `json:"-"`
	BytesInRate  float64      `json:"bytes_in_rate"`
	BytesOutRate float64      `json:"bytes_out_rate"`
	TotalIn      int64        `json:"total_in"`
	TotalOut     int64        `json:"total_out"`
	Connections  atomic.Int32 `json:"-"`
	ConnCount    int32        `json:"connections"`
}

type LatencyInfo struct {
	Latency   int64  `json:"latency"` // ms
	Status    string `json:"status"`  // normal, warning, error
	LastCheck int64  `json:"last_check"`
}

type TunnelStatus struct {
	Rule     Rule         `json:"rule"`
	Traffic  TrafficStats `json:"traffic"`
	Latency  LatencyInfo  `json:"latency"`
	Running  bool         `json:"running"`
	NodeHost string       `json:"node_host,omitempty"`
}

type SystemStats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryTotal   uint64  `json:"memory_total"`
	NetBytesIn    uint64  `json:"net_bytes_in"`
	NetBytesOut   uint64  `json:"net_bytes_out"`
	NetRateIn     float64 `json:"net_rate_in"`
	NetRateOut    float64 `json:"net_rate_out"`
	ActiveTunnels int     `json:"active_tunnels"`
	Uptime        int64   `json:"uptime"`
}

type GlobalTraffic struct {
	TotalIn  int64   `json:"total_in"`
	TotalOut int64   `json:"total_out"`
	RateIn   float64 `json:"rate_in"`
	RateOut  float64 `json:"rate_out"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DashboardData struct {
	System    SystemStats    `json:"system"`
	Global    GlobalTraffic  `json:"global"`
	Tunnels   []TunnelStatus `json:"tunnels"`
	Timestamp int64          `json:"timestamp"`
}

type TrafficHistory struct {
	Timestamp int64   `json:"timestamp"`
	RateIn    float64 `json:"rate_in"`
	RateOut   float64 `json:"rate_out"`
}
