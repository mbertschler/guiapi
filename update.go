package guiapi

import (
	"github.com/mbertschler/guiapi/api"
)

// Update is the returned body of a GUI API call.
type Update struct {
	Name   string           `json:",omitempty"` // Name of the action that was called
	URL    string           `json:",omitempty"` // URL that was loaded
	Error  *api.Error       `json:",omitempty"` // Error that occurred while handling the action
	HTML   []api.HTMLUpdate `json:",omitempty"` // DOM updates to apply
	JS     []api.JSCall     `json:",omitempty"` // JS calls to execute
	State  any              `json:",omitempty"` // State to pass back to the browser
	Stream []api.Stream     `json:",omitempty"` // Stream to subscribe to via websocket
}

// JSCall returns a new Update that will call the registered JavaScript
// function that is identified by the name, and will pass the arguments.
func JSCall(name string, args any) *Update {
	u := &Update{}
	u.AddJSCall(name, args)
	return u
}

// AddJSCall adds a new JSCall Update that will call the registered JavaScript
// function that is identified by the name, and will pass the arguments.
func (u *Update) AddJSCall(name string, args any) {
	u.JS = append(u.JS, api.JSCall{
		Name: name,
		Args: args,
	})
}

// Stream returns a new Update that will connect to a stream
// with the passed name and arguments.
func Stream(name string, args any) *Update {
	u := &Update{}
	u.AddStream(name, args)
	return u
}

// AddStream adds new stream Update that will connect to
// a stream with the passed name and arguments.
func (u *Update) AddStream(name string, args any) {
	u.Stream = append(u.Stream, api.Stream{
		Name: name,
		Args: args,
	})
}

// ReplaceContent returns a new Update that replaces the content of the element
// that gets selected by the passed selector with the HTML content.
// The selector gets passed to document.querySelector,
// so it can be any valid CSS selector.
func ReplaceContent(selector, content string) *Update {
	u := &Update{}
	u.AddReplaceContent(selector, content)
	return u
}

// AddReplaceContent adds a new HTML Update that replaces the content of the element
// that gets selected by the passed selector with the HTML content.
// The selector gets passed to document.querySelector,
// so it can be any valid CSS selector.
func (u *Update) AddReplaceContent(selector, content string) {
	u.HTML = append(u.HTML, api.HTMLUpdate{
		Operation: api.HTMLReplaceContent,
		Selector:  selector,
		Content:   content,
	})
}

// ReplaceElement returns a new Update that replaces the whole element
// that gets selected by the passed selector with the HTML content.
// The selector gets passed to document.querySelector,
// so it can be any valid CSS selector.
func ReplaceElement(selector, content string) *Update {
	u := &Update{}
	u.AddReplaceElement(selector, content)
	return u
}

// AddReplaceElement adds a HTML Update that replaces the whole element
// that gets selected by the passed selector with the HTML content.
// The selector gets passed to document.querySelector,
// so it can be any valid CSS selector.
func (u *Update) AddReplaceElement(selector, content string) {
	u.HTML = append(u.HTML, api.HTMLUpdate{
		Operation: api.HTMLReplaceElement,
		Selector:  selector,
		Content:   content,
	})
}

// InsertBefore returns a new Update that inserts HTML content on the
// same level before the passed selector. The selector gets passed
// to document.querySelector, so it can be any valid CSS selector.
func InsertBefore(selector, content string) *Update {
	u := &Update{}
	u.AddInsertBefore(selector, content)
	return u
}

// AddInsertBefore adds a HTML update that inserts HTML content on the
// same level before the passed selector. The selector gets passed
// to document.querySelector, so it can be any valid CSS selector.
func (u *Update) AddInsertBefore(selector, content string) {
	u.HTML = append(u.HTML, api.HTMLUpdate{
		Operation: api.HTMLInsertBefore,
		Selector:  selector,
		Content:   content,
	})
}

// InsertAfter returns a new Update that inserts HTML content on the
// same level after the passed selector. The selector gets passed
// to document.querySelector, so it can be any valid CSS selector.
func InsertAfter(selector, content string) *Update {
	u := &Update{}
	u.AddInsertAfter(selector, content)
	return u
}

// AddInsertAfter adds a HTML update that inserts HTML content on the
// same level after the passed selector. The selector gets passed
// to document.querySelector, so it can be any valid CSS selector.
func (u *Update) AddInsertAfter(selector, content string) {
	u.HTML = append(u.HTML, api.HTMLUpdate{
		Operation: api.HTMLInsertAfter,
		Selector:  selector,
		Content:   content,
	})
}
