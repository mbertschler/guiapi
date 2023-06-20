package guiapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/html"
)

// NewGuiapi returns an empty handler
func NewGuiapi(sr StreamRouter) *Handler {
	return &Handler{
		Functions:    map[string]Callable{},
		Router:       httprouter.New(),
		StreamRouter: sr,
	}
}

// SetFunc sets a callable GUI API function in the handler.
func (h *Handler) SetFunc(name string, fn Callable) {
	h.Functions[name] = fn
}

// ReplaceContent is a helper function that returns a Result that replaces
// the element content chosen by the selector with the passed Block.
func ReplaceContent(selector string, block html.Block) (*Response, error) {
	out, err := html.RenderMinifiedString(block)
	if err != nil {
		return nil, err
	}
	ret := &Response{
		HTML: []HTMLUpdate{
			{
				Operation: HTMLReplaceContent,
				Selector:  selector,
				Content:   out,
			},
		},
	}
	return ret, nil
}

// ReplaceElement is a helper function that returns a Result that
// replaces the whole element chosen by the selector with the passed Block.
func ReplaceElement(selector string, block html.Block) (*Response, error) {
	out, err := html.RenderMinifiedString(block)
	if err != nil {
		return nil, err
	}
	ret := &Response{
		HTML: []HTMLUpdate{
			{
				Operation: HTMLReplaceElement,
				Selector:  selector,
				Content:   out,
			},
		},
	}
	return ret, nil
}

// InsertBefore is a helper function that returns a Result that
// inserts a block on the same level before the passed selector.
func InsertBefore(selector string, block html.Block) (*Response, error) {
	out, err := html.RenderMinifiedString(block)
	if err != nil {
		return nil, err
	}
	ret := &Response{
		HTML: []HTMLUpdate{
			{
				Operation: HTMLInsertBefore,
				Selector:  selector,
				Content:   out,
			},
		},
	}
	return ret, nil
}

// Redirect lets the browser navigate to a given path
func Redirect(path string) (*Response, error) {
	ret := &Response{
		JS: []JSCall{
			{
				Name: "redirect",
				Args: path,
			},
		},
	}
	return ret, nil
}

// Handle handles HTTP requests to the GUI API.
func (h *Handler) Handle(c *Context) {
	var req Request
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		log.Println("guiapi: error decoding request:", err)
		return
	}
	if req.URL != "" {
		h.processURL(c, &req)
		return
	}
	resp := h.process(c, &req)
	err = json.NewEncoder(c.Writer).Encode(resp)
	if err != nil {
		log.Println("guiapi: error encoding response:", err)
		return
	}
}

func (h *Handler) process(c *Context, req *Request) *Response {
	var res = Response{
		Name: req.Name,
	}

	fn, ok := h.Functions[req.Name]
	if !ok {
		res.Error = &Error{
			Code:    "undefinedFunction",
			Message: fmt.Sprint(req.Name, " is not defined"),
		}
	} else {
		c.State = req.State
		r, err := fn(c, req.Args)
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

func (h *Handler) processURL(c *Context, req *Request) {
	url, err := url.Parse(req.URL)
	if err != nil {
		log.Println("guiapi: error parsing url:", err)
		c.Writer.WriteHeader(400)
		c.Writer.Write([]byte(`{"error":"400 bad request"}`))
		return
	}
	handle, params, _ := h.Router.Lookup("GET", url.Path)
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

type Handler struct {
	Router       *httprouter.Router
	Functions    map[string]Callable
	StreamRouter StreamRouter
}

type HandlerFunc func(*Context)

type Callable func(c *Context, args json.RawMessage) (*Response, error)

// Response is the returned body of a GUI API call
type Response struct {
	// Name of the action that was called
	Name string `json:",omitempty"`
	// URL that was loaded.
	URL    string       `json:",omitempty"`
	Error  *Error       `json:",omitempty"`
	HTML   []HTMLUpdate `json:",omitempty"` // DOM updates to apply
	JS     []JSCall     `json:",omitempty"` // JS calls to execute
	State  any          `json:",omitempty"` // State to pass back to the browser
	Stream any          `json:",omitempty"` // Stream to subscribe to via websocket
}

type Error struct {
	Code    string
	Message string
}

type HTMLOp int8

const (
	HTMLReplaceContent HTMLOp = 1
	HTMLReplaceElement HTMLOp = 2
	HTMLInsertBefore   HTMLOp = 3
	HTMLInsertAfter    HTMLOp = 4
)

type HTMLUpdate struct {
	Operation HTMLOp // how to apply this update
	Selector  string // querySelector syntax: #id .class
	Content   string `json:",omitempty"`
}

type JSCall struct {
	Name string // name of the function to call
	// Args as object, gets encoded by the called function
	Args any `json:",omitempty"`
}

func (r *Response) AddJSResponse(name string, args any) {
	r.JS = append(r.JS, JSCall{Name: name, Args: args})
}
