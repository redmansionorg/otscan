package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/workshop1/otscan/internal/cache"
	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/service"
	"github.com/workshop1/otscan/internal/store"
)

type Server struct {
	cfg          *config.Config
	rpcClient    *rpc.Client
	db           *store.DB
	cache        *cache.Cache
	nodeSvc      *service.NodeService
	batchSvc     *service.BatchService
	claimSvc     *service.ClaimService
	wsHub        *WSHub
	router       *gin.Engine
}

func NewServer(cfg *config.Config, rpcClient *rpc.Client, db *store.DB, c *cache.Cache) *Server {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:       cfg,
		rpcClient: rpcClient,
		db:        db,
		cache:     c,
		nodeSvc:   service.NewNodeService(db, c, rpcClient, cfg),
		batchSvc:  service.NewBatchService(db, rpcClient, cfg),
		claimSvc:  service.NewClaimService(db, rpcClient, cfg),
		wsHub:     NewWSHub(),
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), corsMiddleware())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", s.handleHealth)
		v1.GET("/config", s.handleConfig)
		v1.GET("/dashboard", s.handleDashboard)
		v1.GET("/nodes", s.handleListNodes)
		v1.GET("/nodes/:name", s.handleGetNode)
		v1.GET("/batches", s.handleListBatches)
		v1.GET("/batches/:id", s.handleGetBatch)
		v1.GET("/batches/:id/ruids", s.handleGetBatchRUIDs)
		v1.GET("/claims", s.handleSearchClaims)
		v1.GET("/claims/conflicts/:auid", s.handleCheckConflict)
		v1.GET("/claims/:ruid", s.handleGetClaim)
		v1.GET("/conflicts", s.handleListConflicts)
		v1.GET("/stats/claims", s.handleClaimStats)
		v1.POST("/verify", s.handleVerify)
		v1.GET("/proof/:batchId", s.handleGetProof)
		v1.GET("/nodes/:name/history", s.handleNodeHistory)
		v1.GET("/nodes/:name/calendar", s.handleNodeCalendar)
		v1.GET("/nodes/:name/calendar-url-status", s.handleNodeCalendarURLStatus)
		v1.GET("/ws", s.handleWS)
	}

	// Serve React SPA from web/dist
	s.serveSPA(r)

	s.router = r
	return s
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

// Hub returns the WebSocket hub for external event broadcasting.
func (s *Server) Hub() *WSHub {
	return s.wsHub
}

// pickNode returns the RPC URL of the first configured node (reference node).
func (s *Server) pickNode() string {
	return s.cfg.Nodes[0].RPCURL
}

// findNodeURL returns the RPC URL for a given node name, or empty string.
func (s *Server) findNodeURL(name string) string {
	for _, n := range s.cfg.Nodes {
		if n.Name == name {
			return n.RPCURL
		}
	}
	return ""
}

func (s *Server) serveSPA(r *gin.Engine) {
	distPath := "web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return
	}

	// Hashed assets: long cache
	assetsGroup := r.Group("/assets")
	assetsGroup.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	})
	assetsGroup.Static("/", filepath.Join(distPath, "assets"))

	r.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))

	// index.html: no cache so new deploys take effect immediately
	r.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.File(filepath.Join(distPath, "index.html"))
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// JSON response helpers
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func okJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

func errJSON(c *gin.Context, code int, msg string) {
	c.JSON(code, APIResponse{Success: false, Error: msg})
}
