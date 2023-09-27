package guiapi

import (
	"encoding/json"

	"github.com/mbertschler/guiapi/api"
)

// Request is the sent body of a GUI API call
type Request struct {
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

// Response is the returned body of a GUI API call
type Response struct {
	// Name of the action that was called
	Name string `json:",omitempty"`
	// URL that was loaded.
	URL    string           `json:",omitempty"`
	Error  *api.Error       `json:",omitempty"`
	HTML   []api.HTMLUpdate `json:",omitempty"` // DOM updates to apply
	JS     []api.JSCall     `json:",omitempty"` // JS calls to execute
	State  any              `json:",omitempty"` // State to pass back to the browser
	Stream []api.Stream     `json:",omitempty"` // Stream to subscribe to via websocket
}

func (r *Response) AddJSCall(name string, args any) {
	r.JS = append(r.JS, api.JSCall{
		Name: name,
		Args: args,
	})
}

func JSCall(name string, args any) *Response {
	r := &Response{}
	r.AddJSCall(name, args)
	return r
}

func (r *Response) AddStream(name string, args any) {
	r.Stream = append(r.Stream, api.Stream{
		Name: name,
		Args: args,
	})
}

func Stream(name string, args any) *Response {
	r := &Response{}
	r.AddJSCall(name, args)
	return r
}

// ReplaceContent is a helper function that returns a Result that replaces
// the element content chosen by the selector with the passed Block.
func ReplaceContent(selector, content string) *Response {
	r := &Response{}
	r.AddReplaceContent(selector, content)
	return r
}

// ReplaceContent is a helper function that returns a Result that replaces
// the element content chosen by the selector with the passed Block.
func (r *Response) AddReplaceContent(selector, content string) {
	r.HTML = append(r.HTML, api.HTMLUpdate{
		Operation: api.HTMLReplaceContent,
		Selector:  selector,
		Content:   content,
	})
}

// ReplaceElement is a helper function that returns a Result that
// replaces the whole element chosen by the selector with the passed Block.
func ReplaceElement(selector, content string) *Response {
	r := &Response{}
	r.AddReplaceElement(selector, content)
	return r
}

// ReplaceElement is a helper function that returns a Result that
// replaces the whole element chosen by the selector with the passed Block.
func (r *Response) AddReplaceElement(selector, content string) {
	r.HTML = append(r.HTML, api.HTMLUpdate{
		Operation: api.HTMLReplaceElement,
		Selector:  selector,
		Content:   content,
	})
}

// InsertBefore is a helper function that returns a Result that
// inserts a block on the same level before the passed selector.
func InsertBefore(selector, content string) *Response {
	r := &Response{}
	r.AddInsertBefore(selector, content)
	return r
}

// InsertBefore is a helper function that returns a Result that
// inserts a block on the same level before the passed selector.
func (r *Response) AddInsertBefore(selector, content string) {
	r.HTML = append(r.HTML, api.HTMLUpdate{
		Operation: api.HTMLInsertBefore,
		Selector:  selector,
		Content:   content,
	})
}

// InsertAfter is a helper function that returns a Result that
// inserts a block on the same level after the passed selector.
func InsertAfter(selector, content string) *Response {
	r := &Response{}
	r.AddInsertAfter(selector, content)
	return r
}

// InsertAfter is a helper function that returns a Result that
// inserts a block on the same level after the passed selector.
func (r *Response) AddInsertAfter(selector, content string) {
	r.HTML = append(r.HTML, api.HTMLUpdate{
		Operation: api.HTMLInsertAfter,
		Selector:  selector,
		Content:   content,
	})
}
