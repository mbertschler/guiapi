package api

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

type Stream struct {
	Name string // name of the stream to subscribe to
	// Args as object, gets encoded by the called function
	Args any `json:",omitempty"`
}
