package guiapi

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mbertschler/blocks/html"
)

type Server struct {
	engine      *gin.Engine
	guiapi      *Handler
	middleware  gin.HandlerFunc
	withSession *gin.RouterGroup
}

func New() *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	guiapi := NewGuiapi()

	s := &Server{
		engine:      engine,
		guiapi:      guiapi,
		middleware:  func(c *gin.Context) { c.Next() },
		withSession: engine.Group(""),
	}
	s.withSession.Use(func(c *gin.Context) { s.middleware(c) })
	s.withSession.POST("/guiapi", guiapi.Handle)
	return s
}

func (s *Server) SessionMiddleware(handler gin.HandlerFunc) {
	s.middleware = handler
}

type Component interface {
	Component() *ComponentConfig
}

type ComponentConfig struct {
	Name    string
	Actions map[string]Callable
	Pages   map[string]PageFunc
}

func (s *Server) RegisterComponent(c Component) {
	config := c.Component()
	for name, fn := range config.Actions {
		s.SetFunc(config.Name+"."+name, fn)
	}
}

func (s *Server) RegisterPage(path string, fn PageFunc) {
	s.Page(path, fn)
}

// Static serves static files from the given directory.
func (s *Server) StaticDir(path, dir string) {
	s.engine.Static(path, dir)
}

// Static serves static files from the given directory.
func (s *Server) StaticFS(url string, fs http.FileSystem) {
	s.engine.StaticFS(url, fs)
}

func (s *Server) SetFunc(name string, fn Callable) {
	s.guiapi.SetFunc(name, fn)
}

type PageFunc func(*gin.Context) (html.Block, error)

func (s *Server) Page(path string, page PageFunc) {
	s.withSession.GET(path, func(c *gin.Context) {
		pageBlock, err := page(c)
		if err != nil {
			log.Println("Page error:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		err = html.RenderMinified(c.Writer, pageBlock)
		if err != nil {
			log.Println("RenderMinified error:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	})
}

func (s *Server) Handler() http.Handler {
	return s.engine.Handler()
}
