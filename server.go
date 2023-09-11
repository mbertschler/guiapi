package guiapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	httpRouter   *httprouter.Router
	pagesRouter  *httprouter.Router
	actions      map[string]ActionFunc
	streamRouter StreamRouter
}

func New(streamRouter StreamRouter) *Server {
	s := &Server{
		httpRouter:   httprouter.New(),
		pagesRouter:  httprouter.New(),
		actions:      map[string]ActionFunc{},
		streamRouter: streamRouter,
	}
	s.httpRouter.POST("/guiapi", s.withPageCtx(s.handle))
	s.httpRouter.GET("/guiapi/ws", s.withPageCtx(s.websocketHandler))
	return s
}

func (s *Server) withPageCtx(handler func(*PageCtx)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := &PageCtx{
			Writer:  w,
			Request: r,
			Params:  ps,
		}
		handler(c)
	}
}

func (s *Server) AddPage(path string, p PageFunc) {
	s.pageHTML(path, p)
	s.pageUpdate(path, p)
}

func (s *Server) AddFiles(baseURL string, fs http.FileSystem) {
	s.httpRouter.ServeFiles(baseURL+"*filepath", fs)
}

func (s *Server) AddAction(name string, fn ActionFunc) {
	s.actions[name] = fn
}

type Page interface {
	WriteHTML(io.Writer) error
	Update() (*Response, error)
}

type PageCtx struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  httprouter.Params
}

type PageFunc func(*PageCtx) (Page, error)

func (s *Server) pageHTML(path string, page PageFunc) {
	s.httpRouter.GET(path, s.withPageCtx(func(c *PageCtx) {
		res, err := page(c)
		if err != nil {
			log.Println("page error:", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = res.WriteHTML(c.Writer)
		if err != nil {
			log.Println("page.HTML error:", err)
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
}

func (s *Server) pageUpdate(path string, page PageFunc) {
	s.pagesRouter.GET(path, s.withPageCtx(func(c *PageCtx) {
		res, err := page(c)
		if err != nil {
			log.Println("page error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		resp, err := res.Update()
		if err != nil {
			log.Println("page.Update error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(c.Writer).Encode(resp)
		if err != nil {
			log.Println("write error:", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpRouter.ServeHTTP(w, r)
}
