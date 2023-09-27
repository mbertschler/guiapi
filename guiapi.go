package guiapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/mbertschler/guiapi/api"
)

// action is the sent body of a GUI API call
type action struct {
	// Name of the action that is called
	Name string `json:",omitempty"`
	// URL is the URL of the next page that should be loaded via guiapi.
	URL string `json:",omitempty"`
	// Args as object, gets parsed by the called function
	Args json.RawMessage `json:",omitempty"`
	// State is can be passed back and forth between the server and browser.
	// It is held in a JavaScript variable, so there is one per browser tab.
	State json.RawMessage `json:",omitempty"`
}

// handle handles HTTP requests to the GUI API.
func (s *Server) handle(c *PageCtx) {
	var req action
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		log.Println("guiapi: error decoding request:", err)
		return
	}
	if req.URL != "" {
		s.processURL(c, &req)
		return
	}
	resp := s.process(c, &req)
	err = json.NewEncoder(c.Writer).Encode(resp)
	if err != nil {
		log.Println("guiapi: error encoding response:", err)
		return
	}
}

func (s *Server) process(p *PageCtx, req *action) *Update {
	var res = Update{
		Name: req.Name,
	}

	action, ok := s.actions[req.Name]
	if !ok {
		res.Error = &api.Error{
			Code:    "undefinedFunction",
			Message: fmt.Sprint(req.Name, " is not defined"),
		}
	} else {
		actionCtx := ActionCtx{
			Writer:  p.Writer,
			Request: p.Request,
			State:   req.State,
			Args:    req.Args,
		}
		r, err := action(&actionCtx)
		if err != nil {
			res.Error = &api.Error{
				Code:    "error",
				Message: err.Error(),
			}
		}
		if r != nil {
			if err == nil {
				res.Error = r.Error
			}
			res = *r
			res.Name = req.Name
		}
	}
	return &res
}

func (s *Server) processURL(c *PageCtx, req *action) {
	url, err := url.Parse(req.URL)
	if err != nil {
		log.Println("guiapi: error parsing url:", err)
		c.Writer.WriteHeader(400)
		c.Writer.Write([]byte(`{"error":"400 bad request"}`))
		return
	}
	handle, params, _ := s.pagesRouter.Lookup("GET", url.Path)
	if handle == nil {
		log.Println("guiapi: no handler found for", req.URL)
		c.Writer.WriteHeader(404)
		c.Writer.Write([]byte(`{"error":"404 page not found"}`))
		return
	}
	handle(c.Writer, c.Request, params)
}

// ActionCtx is the context that is passed to an ActionFunc.
// It extends the Request and Writer from a typical HTTP request handler with
// State and Args fields from the Action call that were sent from the browser.
type ActionCtx struct {
	Writer  http.ResponseWriter
	Request *http.Request
	State   json.RawMessage
	Args    json.RawMessage
}

// ActionFunc is the action handler function that should return an Update in
// response to the call from the client. ActionFuncs are registered with the
// Server using the AddAction() function.
type ActionFunc func(c *ActionCtx) (*Update, error)
