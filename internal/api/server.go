package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imide/debateshare/internal/config"
	"github.com/imide/debateshare/internal/logger"
	"github.com/imide/debateshare/internal/sse"
	"github.com/imide/debateshare/internal/storage"
	"github.com/imide/debateshare/web"
	"gorm.io/gorm"
)

type Server struct {
	Router *gin.Engine
	db     *gorm.DB
	store  *storage.Store
	hub    *sse.Hub
	cfg    *config.Config
}

func NewServer(db *gorm.DB, store *storage.Store, cfg *config.Config, hub *sse.Hub) *Server {
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = nil

	r := gin.New()
	r.MaxMultipartMemory = 10 << 20

	r.Use(logger.Middleware())
	r.Use(gin.Recovery())
	r.Use()

	s := &Server{
		Router: r,
		db:     db,
		hub:    hub,
		store:  store,
		cfg:    cfg,
	}

	return s
}

func (s *Server) Init() {
	s.RegisterRoutes()

	// use build.go for static file serving with embedded filesystem
	web.ServeStatic(s.Router)
}

func (s *Server) RegisterRoutes() {
	baseURL := s.cfg.Server.BaseURL
	if baseURL == "" {
		baseURL = "/"
	}

	if !strings.HasPrefix(baseURL, "/") {
		baseURL = "/" + baseURL
	}

	routeBase := strings.TrimSuffix(baseURL, "/")

	s.Router.Use(func(c *gin.Context) {
		c.Set("base_url", baseURL)
		c.Next()
	})

	apiGroup := s.Router.Group(routeBase)
	if routeBase != "" {
		apiGroup = apiGroup.Group("")
	}
	api := apiGroup.Group("/api")
	{
		api.POST("/rooms", s.createRoom)
		api.POST("/rooms/:code/join", s.joinRoom)
		api.GET("/rooms/:code", s.getRoom)
		api.GET("/rooms/:code/files", s.listFiles)
		api.POST("/rooms/:code/files", s.uploadFile)
		api.PUT("/rooms/:code/files/:id", s.replaceFile)
		api.DELETE("/rooms/:code/files/:id", s.deleteFile)
		api.GET("/rooms/:code/files/:id/download", s.downloadFile)
		api.GET("/rooms/:code/files/zip", s.getRoomZip)
		api.GET("/rooms/:code/events", s.roomEvents)
	}

	if routeBase != "" {
		// serve root path and index.html
		s.Router.GET(routeBase, web.ServeIndex)
		s.Router.GET(routeBase+"/", web.ServeIndex)
		s.Router.GET(routeBase+"/index.html", web.ServeIndex)
	}

	// register the catch-all handler for SPA routing
	web.ServeStatic(s.Router)
}
