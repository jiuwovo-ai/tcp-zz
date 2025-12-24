package models

type Node struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	Key        string  `json:"key"`
	Online     bool    `json:"online"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	Uptime     int64   `json:"uptime"`
	LastSeen   int64   `json:"last_seen"`
	CreatedAt  int64   `json:"created_at"`
}

type NodeRule struct {
	ID         string `json:"id"`
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	LocalPort  int    `json:"local_port"`
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  int64  `json:"created_at"`
}

type NodeStatus struct {
	NodeKey     string             `json:"node_key"`
	NodeName    string             `json:"node_name"`
	Online      bool               `json:"online"`
	CPUPercent  float64            `json:"cpu_percent"`
	MemPercent  float64            `json:"mem_percent"`
	NetRateIn   float64            `json:"net_rate_in"`
	NetRateOut  float64            `json:"net_rate_out"`
	Uptime      int64              `json:"uptime"`
	TunnelCount int                `json:"tunnel_count"`
	Tunnels     []NodeTunnelStatus `json:"tunnels"`
}

type NodeTunnelStatus struct {
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

type NodeWithStatus struct {
	Node
	TunnelCount   int                `json:"tunnel_count"`
	ActiveTunnels int                `json:"active_tunnels"`
	TotalIn       int64              `json:"total_in"`
	TotalOut      int64              `json:"total_out"`
	Tunnels       []NodeTunnelStatus `json:"tunnels,omitempty"`
}
