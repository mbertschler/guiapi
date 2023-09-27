package guiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Server contains all the registered Pages, Actions, Files and Streams.
// It implements the http.Handler interface, so it can be directly passed
// to a function like http.ListenAndServe(). When a request comes in, the
// server will handle GET requests for pages, POST requests for actions,
// and WebSocket requests for streams.
type Server struct {
	httpRouter  *httprouter.Router
	pagesRouter *httprouter.Router
	actions     map[string]ActionFunc
	streams     map[string]StreamFunc
}

// New returns a new guiapi Server. After registering all the Pages, Actions, Files and Streams,
// the server can be directly used as a http.Handler.
func New() *Server {
	s := &Server{
		httpRouter:  httprouter.New(),
		pagesRouter: httprouter.New(),
		actions:     map[string]ActionFunc{},
		streams:     map[string]StreamFunc{},
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

// AddPage registers a PageFunc with the passed path and page handler function on the server.
// The URL path can contain placeholders in the form of :name, which then get passed as
// Params in a PageCtx.
//
// See https://github.com/julienschmidt/httprouter for more info.
func (s *Server) AddPage(path string, fn PageFunc) {
	s.pageHTML(path, fn)
	s.pageUpdate(path, fn)
}

// AddFiles registers a http.FileSystem with the passed baseURL on the server.
// The files can also be in a subdirectory of the baseURL. This function is the
// main way of serving static files from the guiapi server.
func (s *Server) AddFiles(baseURL string, fs http.FileSystem) {
	s.httpRouter.ServeFiles(baseURL+"*filepath", fs)
}

// AddAction registers an ActionFunc with the passed name and handler function on the server.
func (s *Server) AddAction(name string, fn ActionFunc) {
	s.actions[name] = fn
}

// Page gets returned from a PageFunc. The page needs to be able to
// write the HTML representation of the page to an io.Writer.
// It can be extended into an UpdateablePage by implementing Update().
type Page interface {
	WriteHTML(io.Writer) error
}

// UpdateablePage is a Page that can also produce a relative Update.
// This means that the browser can just update the relevant parts of the page
// and doesn't have to reload the whole page.
type UpdateablePage interface {
	Page
	Update() (*Update, error)
}

// PageCtx is the context that is passed to a PageFunc.
// It extends the Request and Writer from a typical HTTP request handler with
// Params from the HTTP router.
//
// For more info on httrouter.Params see: https://github.com/julienschmidt/httprouter
type PageCtx struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  httprouter.Params // params from placeholders in the URL
}

// PageFunc is the page handler function that should return a Page value in
// response to a HTTP request or guiapi page request. PageFuncs are registered
// with the using the AddPage() function.
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
		updater, ok := res.(UpdateablePage)
		if !ok {
			err := fmt.Sprintf("page %q is not updateable", path)
			log.Println(err)
			http.Error(c.Writer, err, http.StatusNotImplemented)
			return
		}
		resp, err := updater.Update()
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

// ServeHTTP implements the http.Handler interface. This means that the Server
// can directly passed to a function like http.ListenAndServe().
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpRouter.ServeHTTP(w, r)
}

// Router returns the underlying HTTP router. This way any additional endpoints
// can be added to the HTTP server, bypassing guiapi.
//
// See https://github.com/julienschmidt/httprouter for more info.
func (s *Server) Router() *httprouter.Router {
	return s.httpRouter
}

// AddStream registers a StreamFunc with the passed name and handler function on the server.
func (s *Server) AddStream(name string, fn StreamFunc) {
	s.streams[name] = fn
}
