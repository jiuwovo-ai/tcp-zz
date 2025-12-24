package forwarder

import (
	"fmt"
	"log"
	"sync"
	"time"

	"port-forward-dashboard/internal/models"
)

type Manager struct {
	tunnels   map[string]*Tunnel
	mu        sync.RWMutex
	startTime time.Time
}

func NewManager() *Manager {
	return &Manager{
		tunnels:   make(map[string]*Tunnel),
		startTime: time.Now(),
	}
}

func (m *Manager) AddRule(rule models.Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tunnels[rule.ID]; exists {
		return fmt.Errorf("rule %s already exists", rule.ID)
	}

	tunnel := NewTunnel(rule)
	m.tunnels[rule.ID] = tunnel

	if rule.Enabled {
		if err := tunnel.Start(); err != nil {
			log.Printf("Failed to start tunnel %s: %v", rule.ID, err)
		}
	}

	return nil
}

func (m *Manager) UpdateRule(rule models.Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[rule.ID]
	if !exists {
		return fmt.Errorf("rule %s not found", rule.ID)
	}

	wasRunning := tunnel.IsRunning()
	if wasRunning {
		tunnel.Stop()
	}

	tunnel.UpdateRule(rule)

	if rule.Enabled {
		if err := tunnel.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %v", err)
		}
	}

	return nil
}

func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("rule %s not found", id)
	}

	tunnel.Stop()
	delete(m.tunnels, id)
	return nil
}

func (m *Manager) ToggleRule(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("rule %s not found", id)
	}

	if enabled {
		return tunnel.Start()
	}
	tunnel.Stop()
	return nil
}

func (m *Manager) GetTunnelStatus(id string) (*models.TunnelStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", id)
	}

	return tunnel.GetStatus(), nil
}

func (m *Manager) GetAllStatus() []models.TunnelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]models.TunnelStatus, 0, len(m.tunnels))
	for _, tunnel := range m.tunnels {
		statuses = append(statuses, *tunnel.GetStatus())
	}
	return statuses
}

func (m *Manager) GetAllRules() []models.Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]models.Rule, 0, len(m.tunnels))
	for _, tunnel := range m.tunnels {
		rules = append(rules, tunnel.GetRule())
	}
	return rules
}

func (m *Manager) GetGlobalTraffic() models.GlobalTraffic {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var global models.GlobalTraffic
	for _, tunnel := range m.tunnels {
		stats := tunnel.GetTrafficStats()
		global.TotalIn += stats.TotalIn
		global.TotalOut += stats.TotalOut
		global.RateIn += stats.BytesInRate
		global.RateOut += stats.BytesOutRate
	}
	return global
}

func (m *Manager) GetActiveTunnelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, tunnel := range m.tunnels {
		if tunnel.IsRunning() {
			count++
		}
	}
	return count
}

func (m *Manager) GetUptime() int64 {
	return int64(time.Since(m.startTime).Seconds())
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tunnel := range m.tunnels {
		tunnel.Stop()
	}
}

func (m *Manager) UpdateAllRates() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tunnel := range m.tunnels {
		tunnel.UpdateRates()
	}
}
