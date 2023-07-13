package guiapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

// Handle handles HTTP requests to the GUI API.
func (s *Server) Handle(c *Context) {
	var req Request
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

func (s *Server) process(c *Context, req *Request) *Response {
	var res = Response{
		Name: req.Name,
	}

	action, ok := s.actions[req.Name]
	if !ok {
		res.Error = &Error{
			Code:    "undefinedFunction",
			Message: fmt.Sprint(req.Name, " is not defined"),
		}
	} else {
		c.State = req.State
		r, err := action(c, req.Args)
		if err != nil {
			res.Error = &Error{
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

func (s *Server) processURL(c *Context, req *Request) {
	url, err := url.Parse(req.URL)
	if err != nil {
		log.Println("guiapi: error parsing url:", err)
		c.Writer.WriteHeader(400)
		c.Writer.Write([]byte(`{"error":"400 bad request"}`))
		return
	}
	handle, params, _ := s.pages.Lookup("GET", url.Path)
	if handle == nil {
		log.Println("guiapi: no handler found for", req.URL)
		c.Writer.WriteHeader(404)
		c.Writer.Write([]byte(`{"error":"404 page not found"}`))
		return
	}
	handle(c.Writer, c.Request, params)
}

// Request is the sent body of a GUI API call
type Request struct {
	// Name of the action that is called
	Name string `json:",omitempty"`
	// URL is the URL of the next page that should be loaded via guiapi.
	URL string `json:",omitempty"`
	// Args as object, gets parsed by the called function
	Args json.RawMessage `json:",omitempty"`
	// State is can be passed back and forth between the server and browser.
	// It is held in a Javascript variable, so there is one per browser tab.
	State json.RawMessage `json:",omitempty"`
}

type HandlerFunc func(*Context)

type ActionFunc func(c *Context, args json.RawMessage) (*Response, error)
