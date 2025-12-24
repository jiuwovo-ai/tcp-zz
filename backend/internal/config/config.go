package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"port-forward-dashboard/internal/models"
)

type Config struct {
	Port      int               `json:"port"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
	JWTSecret string            `json:"jwt_secret"`
	Rules     []models.Rule     `json:"rules"`
	Nodes     []models.Node     `json:"nodes"`
	NodeRules []models.NodeRule `json:"node_rules"`
	mu        sync.RWMutex
}

const configFile = "config.json"

func Load() *Config {
	cfg := &Config{
		Port:      8080,
		Username:  "admin",
		Password:  "admin123",
		JWTSecret: "your-secret-key-change-in-production",
		Rules:     []models.Rule{},
		Nodes:     []models.Node{},
		NodeRules: []models.NodeRule{},
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Println("No config file found, using defaults")
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		log.Printf("Failed to parse config: %v", err)
		return cfg
	}

	return cfg
}

func Save(cfg *Config, rules []models.Rule) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.Rules = rules
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal config: %v", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

func (cfg *Config) Save() {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal config: %v", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}
