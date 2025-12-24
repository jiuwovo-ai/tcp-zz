package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"port-forward-dashboard/internal/models"
)

func (s *Server) handleGetNodes(c *gin.Context) {
	nodes := s.nm.GetAllNodes()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: nodes})
}

func (s *Server) handleCreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	node.ID = generateID()
	node.CreatedAt = time.Now().Unix()
	node.Online = false

	if err := s.nm.AddNode(node); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: node})
}

func (s *Server) handleUpdateNode(c *gin.Context) {
	id := c.Param("id")

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	node.ID = id
	if err := s.nm.UpdateNode(node); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: node})
}

func (s *Server) handleDeleteNode(c *gin.Context) {
	id := c.Param("id")

	if err := s.nm.DeleteNode(id); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Node deleted"})
}

func (s *Server) handleGetNodeRules(c *gin.Context) {
	nodeID := c.Query("node_id")
	var rules []models.NodeRule

	if nodeID != "" {
		rules = s.nm.GetRulesByNode(nodeID)
	} else {
		rules = s.nm.GetAllRules()
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rules})
}

func (s *Server) handleCreateNodeRule(c *gin.Context) {
	var rule models.NodeRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	rule.ID = generateID()
	rule.CreatedAt = time.Now().Unix()

	if err := s.nm.AddRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rule})
}

func (s *Server) handleUpdateNodeRule(c *gin.Context) {
	id := c.Param("id")

	var rule models.NodeRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	rule.ID = id
	if err := s.nm.UpdateRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: rule})
}

func (s *Server) handleDeleteNodeRule(c *gin.Context) {
	id := c.Param("id")

	if err := s.nm.DeleteRule(id); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Rule deleted"})
}

func (s *Server) handleToggleNodeRule(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	if err := s.nm.ToggleRule(id, req.Enabled); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	s.saveNodeConfig()

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (s *Server) handleNodeHeartbeat(c *gin.Context) {
	var status models.NodeStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request"})
		return
	}

	s.nm.HandleHeartbeat(status)

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (s *Server) saveNodeConfig() {
	nodes := s.nm.GetNodesForSave()
	rules := s.nm.GetAllRules()
	s.cfg.Nodes = nodes
	s.cfg.NodeRules = rules
	s.cfg.Save()
}
