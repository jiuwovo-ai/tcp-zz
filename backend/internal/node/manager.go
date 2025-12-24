package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"port-forward-dashboard/internal/models"
)

type Manager struct {
	nodes  map[string]*NodeInfo
	rules  map[string]*models.NodeRule
	mu     sync.RWMutex
	client *http.Client
}

type NodeInfo struct {
	Node      models.Node
	Status    *models.NodeStatus
	LastCheck time.Time
}

func NewManager() *Manager {
	m := &Manager{
		nodes:  make(map[string]*NodeInfo),
		rules:  make(map[string]*models.NodeRule),
		client: &http.Client{Timeout: 10 * time.Second},
	}
	go m.healthCheckLoop()
	return m
}

func (m *Manager) AddNode(node models.Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[node.ID]; exists {
		return fmt.Errorf("node %s already exists", node.ID)
	}

	m.nodes[node.ID] = &NodeInfo{
		Node:      node,
		LastCheck: time.Now(),
	}

	return nil
}

func (m *Manager) UpdateNode(node models.Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.nodes[node.ID]
	if !exists {
		return fmt.Errorf("node %s not found", node.ID)
	}

	info.Node = node
	return nil
}

func (m *Manager) DeleteNode(id string) error {
	m.mu.Lock()
	info, exists := m.nodes[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("node %s not found", id)
	}

	// 删除该节点的所有规则
	for ruleID, rule := range m.rules {
		if rule.NodeID == id {
			delete(m.rules, ruleID)
		}
	}

	delete(m.nodes, id)
	m.mu.Unlock()

	// 发送卸载命令到节点（异步，不阻塞）
	go m.sendUninstallToNode(info)

	return nil
}

func (m *Manager) sendUninstallToNode(info *NodeInfo) {
	url := fmt.Sprintf("http://%s:%d/uninstall", info.Node.Host, info.Node.Port)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		// 节点可能已经离线，忽略错误
		return
	}
	defer resp.Body.Close()
}

func (m *Manager) GetNode(id string) (*models.NodeWithStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s not found", id)
	}

	return m.buildNodeWithStatus(info), nil
}

func (m *Manager) GetAllNodes() []models.NodeWithStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]models.NodeWithStatus, 0, len(m.nodes))
	for _, info := range m.nodes {
		result = append(result, *m.buildNodeWithStatus(info))
	}
	return result
}

func (m *Manager) buildNodeWithStatus(info *NodeInfo) *models.NodeWithStatus {
	nws := &models.NodeWithStatus{
		Node: info.Node,
	}

	if info.Status != nil {
		nws.TunnelCount = info.Status.TunnelCount
		nws.Tunnels = info.Status.Tunnels

		for _, t := range info.Status.Tunnels {
			if t.Running {
				nws.ActiveTunnels++
			}
			nws.TotalIn += t.BytesIn
			nws.TotalOut += t.BytesOut
		}
	}

	return nws
}

func (m *Manager) AddRule(rule models.NodeRule) error {
	m.mu.Lock()
	info, exists := m.nodes[rule.NodeID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("node %s not found", rule.NodeID)
	}
	m.rules[rule.ID] = &rule
	m.mu.Unlock()

	// 发送到节点
	return m.sendRuleToNode(info, &rule, true)
}

func (m *Manager) UpdateRule(rule models.NodeRule) error {
	m.mu.Lock()
	oldRule, exists := m.rules[rule.ID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("rule %s not found", rule.ID)
	}

	info, nodeExists := m.nodes[rule.NodeID]
	if !nodeExists {
		m.mu.Unlock()
		return fmt.Errorf("node %s not found", rule.NodeID)
	}

	// 如果节点变了，先从旧节点删除
	if oldRule.NodeID != rule.NodeID {
		if oldInfo, ok := m.nodes[oldRule.NodeID]; ok {
			m.mu.Unlock()
			m.deleteRuleFromNode(oldInfo, oldRule.ID)
			m.mu.Lock()
		}
	}

	m.rules[rule.ID] = &rule
	m.mu.Unlock()

	return m.sendRuleToNode(info, &rule, rule.Enabled)
}

func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	rule, exists := m.rules[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("rule %s not found", id)
	}

	info, nodeExists := m.nodes[rule.NodeID]
	delete(m.rules, id)
	m.mu.Unlock()

	if nodeExists {
		return m.deleteRuleFromNode(info, id)
	}
	return nil
}

func (m *Manager) ToggleRule(id string, enabled bool) error {
	m.mu.RLock()
	rule, exists := m.rules[id]
	if !exists {
		m.mu.RUnlock()
		return fmt.Errorf("rule %s not found", id)
	}

	info, nodeExists := m.nodes[rule.NodeID]
	m.mu.RUnlock()

	if !nodeExists {
		return fmt.Errorf("node %s not found", rule.NodeID)
	}

	rule.Enabled = enabled

	if enabled {
		return m.startRuleOnNode(info, id)
	}
	return m.stopRuleOnNode(info, id)
}

func (m *Manager) GetAllRules() []models.NodeRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]models.NodeRule, 0, len(m.rules))
	for _, rule := range m.rules {
		result = append(result, *rule)
	}
	return result
}

func (m *Manager) GetRulesByNode(nodeID string) []models.NodeRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.NodeRule
	for _, rule := range m.rules {
		if rule.NodeID == nodeID {
			result = append(result, *rule)
		}
	}
	return result
}

// GetAllTunnelStatus 返回所有节点隧道的状态，用于仪表盘显示
func (m *Manager) GetAllTunnelStatus() []models.TunnelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.TunnelStatus

	// 遍历所有规则，而不是只遍历有状态的节点
	for _, rule := range m.rules {
		info, nodeExists := m.nodes[rule.NodeID]

		// 获取节点 IP
		var nodeHost string
		if nodeExists {
			nodeHost = info.Node.Host
		}

		// 默认状态：离线
		status := models.TunnelStatus{
			Rule: models.Rule{
				ID:         rule.ID,
				Name:       rule.Name,
				LocalPort:  rule.LocalPort,
				TargetIP:   rule.TargetIP,
				TargetPort: rule.TargetPort,
				Protocol:   models.Protocol(rule.Protocol),
				Enabled:    rule.Enabled,
			},
			Traffic: models.TrafficStats{},
			Latency: models.LatencyInfo{
				Latency: -1,
				Status:  "error",
			},
			Running:  false,
			NodeHost: nodeHost,
		}

		// 如果节点在线且有状态，更新实际数据
		if nodeExists && info.Status != nil {
			for _, tunnel := range info.Status.Tunnels {
				if tunnel.ID == rule.ID {
					status.Traffic = models.TrafficStats{
						TotalIn:      tunnel.BytesIn,
						TotalOut:     tunnel.BytesOut,
						BytesInRate:  tunnel.RateIn,
						BytesOutRate: tunnel.RateOut,
					}
					status.Latency = models.LatencyInfo{
						Latency: tunnel.Latency,
						Status:  getLatencyStatus(tunnel.Latency),
					}
					status.Running = tunnel.Running
					break
				}
			}
		}

		result = append(result, status)
	}
	return result
}

// GetGlobalTraffic 返回所有节点的全局流量统计
func (m *Manager) GetGlobalTraffic() (totalIn, totalOut int64, rateIn, rateOut float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, info := range m.nodes {
		if info.Status == nil {
			continue
		}
		for _, tunnel := range info.Status.Tunnels {
			totalIn += tunnel.BytesIn
			totalOut += tunnel.BytesOut
			rateIn += tunnel.RateIn
			rateOut += tunnel.RateOut
		}
	}
	return
}

// GetActiveTunnelCount 返回所有节点的活跃隧道数
func (m *Manager) GetActiveTunnelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, info := range m.nodes {
		if info.Status == nil {
			continue
		}
		for _, tunnel := range info.Status.Tunnels {
			if tunnel.Running {
				count++
			}
		}
	}
	return count
}

func getLatencyStatus(latency int64) string {
	if latency < 0 {
		return "error"
	}
	if latency > 200 {
		return "warning"
	}
	return "normal"
}

func (m *Manager) sendRuleToNode(info *NodeInfo, rule *models.NodeRule, autoStart bool) error {
	url := fmt.Sprintf("http://%s:%d/tunnels", info.Node.Host, info.Node.Port)

	payload := map[string]interface{}{
		"id":          rule.ID,
		"local_port":  rule.LocalPort,
		"target_ip":   rule.TargetIP,
		"target_port": rule.TargetPort,
		"protocol":    rule.Protocol,
		"auto_start":  autoStart,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("node returned error: %s", string(body))
	}

	return nil
}

func (m *Manager) deleteRuleFromNode(info *NodeInfo, ruleID string) error {
	url := fmt.Sprintf("http://%s:%d/tunnels/%s", info.Node.Host, info.Node.Port, ruleID)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func (m *Manager) startRuleOnNode(info *NodeInfo, ruleID string) error {
	url := fmt.Sprintf("http://%s:%d/tunnels/%s/start", info.Node.Host, info.Node.Port, ruleID)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func (m *Manager) stopRuleOnNode(info *NodeInfo, ruleID string) error {
	url := fmt.Sprintf("http://%s:%d/tunnels/%s/stop", info.Node.Host, info.Node.Port, ruleID)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func (m *Manager) healthCheckLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.checkAllNodes()
	}
}

func (m *Manager) checkAllNodes() {
	m.mu.RLock()
	nodes := make([]*NodeInfo, 0, len(m.nodes))
	for _, info := range m.nodes {
		nodes = append(nodes, info)
	}
	m.mu.RUnlock()

	for _, info := range nodes {
		m.checkNode(info)
	}
}

func (m *Manager) checkNode(info *NodeInfo) {
	url := fmt.Sprintf("http://%s:%d/status", info.Node.Host, info.Node.Port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Node-Key", info.Node.Key)

	resp, err := m.client.Do(req)
	if err != nil {
		m.mu.Lock()
		info.Node.Online = false
		m.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	var result struct {
		Success bool              `json:"success"`
		Data    models.NodeStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.mu.Lock()
		info.Node.Online = false
		m.mu.Unlock()
		return
	}

	m.mu.Lock()
	info.Node.Online = true
	info.Node.CPUPercent = result.Data.CPUPercent
	info.Node.MemPercent = result.Data.MemPercent
	info.Node.Uptime = result.Data.Uptime
	info.Node.LastSeen = time.Now().Unix()
	info.Status = &result.Data
	info.LastCheck = time.Now()
	m.mu.Unlock()
}

func (m *Manager) HandleHeartbeat(status models.NodeStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, info := range m.nodes {
		if info.Node.Key == status.NodeKey {
			info.Node.Online = true
			info.Node.CPUPercent = status.CPUPercent
			info.Node.MemPercent = status.MemPercent
			info.Node.Uptime = status.Uptime
			info.Node.LastSeen = time.Now().Unix()
			info.Status = &status
			info.LastCheck = time.Now()
			return
		}
	}
}

func (m *Manager) GetGlobalStats() (totalIn, totalOut int64, activeNodes, activeTunnels int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, info := range m.nodes {
		if info.Node.Online {
			activeNodes++
		}
		if info.Status != nil {
			for _, t := range info.Status.Tunnels {
				if t.Running {
					activeTunnels++
				}
				totalIn += t.BytesIn
				totalOut += t.BytesOut
			}
		}
	}
	return
}

func (m *Manager) RestoreRules(nodes []models.Node, rules []models.NodeRule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, n := range nodes {
		m.nodes[n.ID] = &NodeInfo{
			Node:      n,
			LastCheck: time.Now(),
		}
	}

	for _, r := range rules {
		rule := r
		m.rules[r.ID] = &rule
	}
}

func (m *Manager) GetNodesForSave() []models.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]models.Node, 0, len(m.nodes))
	for _, info := range m.nodes {
		result = append(result, info.Node)
	}
	return result
}
