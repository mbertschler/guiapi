package guiapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/blocks/html"
)

type Server struct {
	router     *httprouter.Router
	guiapi     *Handler
	middleware Middleware
}

func New(middleware Middleware) *Server {
	s := &Server{
		router:     httprouter.New(),
		guiapi:     NewGuiapi(),
		middleware: middleware,
	}
	s.router.POST("/guiapi", s.wrapMiddleware(s.guiapi.Handle))
	return s
}

func (s *Server) wrapMiddleware(handler HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := &Context{
			Writer:  w,
			Request: r,
			Params:  ps,
		}
		if s.middleware != nil {
			s.middleware(c, handler)
		} else {
			handler(c)
		}
	}
}

type Middleware func(c *Context, next HandlerFunc)

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
	for path, fn := range config.Pages {
		s.page(path, fn)
		s.pageUpdate(path, fn)
	}
}

func (s *Server) ServeFiles(url string, fs http.FileSystem) {
	s.router.ServeFiles(url+"*filepath", fs)
}

func (s *Server) SetFunc(name string, fn Callable) {
	s.guiapi.SetFunc(name, fn)
}

type Layout interface {
	Name() string
	RenderPage(*Page) (html.Block, error)
	RenderUpdate(*Page) (*Response, error)
}

type Page struct {
	Layout    Layout
	State     any
	Fragments map[string]html.Block
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  httprouter.Params
	Session any
	State   json.RawMessage
}

type PageFunc func(*Context) (*Page, error)

func (s *Server) page(path string, page PageFunc) {
	s.router.GET(path, s.wrapMiddleware(func(c *Context) {
		pageBlock, err := page(c)
		if err != nil {
			log.Println("Page error:", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		block, err := pageBlock.Layout.RenderPage(pageBlock)
		if err != nil {
			log.Println("Layout error:", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = html.RenderMinified(c.Writer, block)
		if err != nil {
			log.Println("RenderMinified error:", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
}

func (s *Server) pageUpdate(path string, page PageFunc) {
	s.guiapi.Router.GET(path, s.wrapMiddleware(func(c *Context) {
		pageBlock, err := page(c)
		if err != nil {
			log.Println("Page error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		resp, err := pageBlock.Layout.RenderUpdate(pageBlock)
		if err != nil {
			log.Println("Layout error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		buf, err := json.Marshal(resp)
		if err != nil {
			log.Println("JSON error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		n, err := c.Writer.Write(buf)
		if err != nil {
			log.Println("Write error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if n != len(buf) {
			log.Println("Write error: wrote", n, "bytes, expected", len(buf))
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))
}

func (s *Server) Handler() http.Handler {
	return s.router
}
