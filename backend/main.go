package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"port-forward-dashboard/internal/api"
	"port-forward-dashboard/internal/config"
	"port-forward-dashboard/internal/forwarder"
	"port-forward-dashboard/internal/node"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 Port Forward Dashboard Starting...")

	// 加载配置
	cfg := config.Load()

	// 初始化转发管理器（本地转发）
	fm := forwarder.NewManager()

	// 从配置恢复本地规则
	for _, rule := range cfg.Rules {
		if err := fm.AddRule(rule); err != nil {
			log.Printf("Failed to restore rule %s: %v", rule.ID, err)
		}
	}

	// 初始化节点管理器
	nm := node.NewManager()

	// 从配置恢复节点和节点规则
	nm.RestoreRules(cfg.Nodes, cfg.NodeRules)

	// 启动 API 服务器
	server := api.NewServer(cfg, fm, nm)
	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("✅ Server running on http://0.0.0.0:%d", cfg.Port)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	fm.StopAll()
	config.Save(cfg, fm.GetAllRules())
	log.Println("Goodbye!")
}
