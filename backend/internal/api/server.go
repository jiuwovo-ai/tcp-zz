package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"port-forward-dashboard/internal/config"
	"port-forward-dashboard/internal/forwarder"
	"port-forward-dashboard/internal/models"
	"port-forward-dashboard/internal/monitor"
	"port-forward-dashboard/internal/node"
)

type Server struct {
	cfg     *config.Config
	fm      *forwarder.Manager
	nm      *node.Manager
	router  *gin.Engine
	monitor *monitor.SystemMonitor
	hub     *WSHub
}

func NewServer(cfg *config.Config, fm *forwarder.Manager, nm *node.Manager) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		cfg:     cfg,
		fm:      fm,
		nm:      nm,
		router:  gin.New(),
		monitor: monitor.NewSystemMonitor(),
		hub:     NewWSHub(),
	}

	s.setupRoutes()
	go s.hub.Run()
	go s.broadcastLoop()

	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(gin.Recovery())
	s.router.Use(corsMiddleware())

	// 静态文件服务
	s.router.Static("/assets", "./static/assets")
	s.router.StaticFile("/", "./static/index.html")
	s.router.StaticFile("/favicon.ico", "./static/favicon.ico")

	// API 路由
	api := s.router.Group("/api")
	{
		api.POST("/login", s.handleLogin)

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(authMiddleware(s.cfg.JWTSecret))
		{
			auth.GET("/dashboard", s.handleDashboard)
			auth.GET("/rules", s.handleGetRules)
			auth.POST("/rules", s.handleCreateRule)
			auth.PUT("/rules/:id", s.handleUpdateRule)
			auth.DELETE("/rules/:id", s.handleDeleteRule)
			auth.POST("/rules/:id/toggle", s.handleToggleRule)
			auth.GET("/system", s.handleSystemStats)

			// 节点管理
			auth.GET("/nodes", s.handleGetNodes)
			auth.POST("/nodes", s.handleCreateNode)
			auth.PUT("/nodes/:id", s.handleUpdateNode)
			auth.DELETE("/nodes/:id", s.handleDeleteNode)

			// 节点规则管理
			auth.GET("/node-rules", s.handleGetNodeRules)
			auth.POST("/node-rules", s.handleCreateNodeRule)
			auth.PUT("/node-rules/:id", s.handleUpdateNodeRule)
			auth.DELETE("/node-rules/:id", s.handleDeleteNodeRule)
			auth.POST("/node-rules/:id/toggle", s.handleToggleNodeRule)
		}

		// 节点心跳（不需要JWT认证，使用节点Key认证）
		api.POST("/nodes/heartbeat", s.handleNodeHeartbeat)

		// WebSocket（自己验证token）
		api.GET("/ws", s.handleWebSocket)

		// 安装脚本（需要认证）
		auth.GET("/nodes/:id/install", s.handleGetInstallScript)
	}

	// 公开的安装脚本下载（不需要认证）
	s.router.GET("/api/install.sh", s.handleDownloadAgent)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.Port)
	return s.router.Run(addr)
}

func (s *Server) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	if req.Username != s.cfg.Username || req.Password != s.cfg.Password {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Invalid credentials"})
		return
	}

	token, expiresAt, err := generateToken(req.Username, s.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
		},
	})
}

func (s *Server) handleDashboard(c *gin.Context) {
	// 获取节点隧道数据（面板只管理节点，不转发流量）
	tunnels := s.nm.GetAllTunnelStatus()
	totalIn, totalOut, rateIn, rateOut := s.nm.GetGlobalTraffic()
	activeCount := s.nm.GetActiveTunnelCount()

	globalTraffic := models.GlobalTraffic{
		TotalIn:  totalIn,
		TotalOut: totalOut,
		RateIn:   rateIn,
		RateOut:  rateOut,
	}

	data := models.DashboardData{
		System:    s.monitor.GetStats(activeCount, s.monitor.GetUptime()),
		Global:    globalTraffic,
		Tunnels:   tunnels,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: data})
}

func (s *Server) handleGetRules(c *gin.Context) {
	rules := s.fm.GetAllRules()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rules})
}

func (s *Server) handleCreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	rule.ID = generateID()
	rule.CreatedAt = time.Now().Unix()

	if err := s.fm.AddRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// 保存配置
	config.Save(s.cfg, s.fm.GetAllRules())

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rule})
}

func (s *Server) handleUpdateRule(c *gin.Context) {
	id := c.Param("id")

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	rule.ID = id
	if err := s.fm.UpdateRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	config.Save(s.cfg, s.fm.GetAllRules())

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rule})
}

func (s *Server) handleDeleteRule(c *gin.Context) {
	id := c.Param("id")

	if err := s.fm.DeleteRule(id); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	config.Save(s.cfg, s.fm.GetAllRules())

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Rule deleted"})
}

func (s *Server) handleToggleRule(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	if err := s.fm.ToggleRule(id, req.Enabled); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	config.Save(s.cfg, s.fm.GetAllRules())

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (s *Server) handleSystemStats(c *gin.Context) {
	stats := s.monitor.GetStats(s.fm.GetActiveTunnelCount(), s.fm.GetUptime())
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(c *gin.Context) {
	// 支持从 URL 参数获取 token
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Missing token"})
		return
	}

	// 验证 token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WSClient{
		hub:  s.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (s *Server) broadcastLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取节点隧道数据（面板只管理节点，不转发流量）
		tunnels := s.nm.GetAllTunnelStatus()
		totalIn, totalOut, rateIn, rateOut := s.nm.GetGlobalTraffic()
		activeCount := s.nm.GetActiveTunnelCount()

		globalTraffic := models.GlobalTraffic{
			TotalIn:  totalIn,
			TotalOut: totalOut,
			RateIn:   rateIn,
			RateOut:  rateOut,
		}

		data := models.DashboardData{
			System:    s.monitor.GetStats(activeCount, s.monitor.GetUptime()),
			Global:    globalTraffic,
			Tunnels:   tunnels,
			Timestamp: time.Now().Unix(),
		}

		msg := models.WSMessage{
			Type:    "dashboard",
			Payload: data,
		}

		s.hub.Broadcast(msg)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
