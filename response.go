package guiapi

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

func (r *Response) AddJSCall(name string, args any) {
	r.JS = append(r.JS, JSCall{
		Name: name,
		Args: args,
	})
}

func (r *Response) AddHTMLUpdate(operation HTMLOp, selector, content string) {
	r.HTML = append(r.HTML, HTMLUpdate{
		Operation: operation,
		Selector:  selector,
		Content:   content,
	})
}

// ReplaceContent is a helper function that returns a Result that replaces
// the element content chosen by the selector with the passed Block.
func ReplaceContent(selector, content string) *Response {
	r := &Response{}
	r.ReplaceContent(selector, content)
	return r
}

// ReplaceContent is a helper function that returns a Result that replaces
// the element content chosen by the selector with the passed Block.
func (r *Response) ReplaceContent(selector, content string) {
	r.HTML = append(r.HTML, HTMLUpdate{
		Operation: HTMLReplaceContent,
		Selector:  selector,
		Content:   content,
	})
}

// ReplaceElement is a helper function that returns a Result that
// replaces the whole element chosen by the selector with the passed Block.
func ReplaceElement(selector, content string) *Response {
	r := &Response{}
	r.ReplaceElement(selector, content)
	return r
}

// ReplaceElement is a helper function that returns a Result that
// replaces the whole element chosen by the selector with the passed Block.
func (r *Response) ReplaceElement(selector, content string) {
	r.HTML = append(r.HTML, HTMLUpdate{
		Operation: HTMLReplaceElement,
		Selector:  selector,
		Content:   content,
	})
}

// InsertBefore is a helper function that returns a Result that
// inserts a block on the same level before the passed selector.
func InsertBefore(selector, content string) *Response {
	r := &Response{}
	r.InsertBefore(selector, content)
	return r
}

// InsertBefore is a helper function that returns a Result that
// inserts a block on the same level before the passed selector.
func (r *Response) InsertBefore(selector, content string) {
	r.HTML = append(r.HTML, HTMLUpdate{
		Operation: HTMLInsertBefore,
		Selector:  selector,
		Content:   content,
	})
}

// InsertAfter is a helper function that returns a Result that
// inserts a block on the same level after the passed selector.
func InsertAfter(selector, content string) *Response {
	r := &Response{}
	r.InsertAfter(selector, content)
	return r
}

// InsertAfter is a helper function that returns a Result that
// inserts a block on the same level after the passed selector.
func (r *Response) InsertAfter(selector, content string) {
	r.HTML = append(r.HTML, HTMLUpdate{
		Operation: HTMLInsertAfter,
		Selector:  selector,
		Content:   content,
	})
}
