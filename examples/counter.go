package main

import (
	"encoding/json"
	"fmt"

	"github.com/mbertschler/blocks/html"
	"github.com/mbertschler/blocks/html/attr"
	"github.com/mbertschler/guiapi"
)

type Counter struct {
	*DB
}

func (c *Counter) Component() *guiapi.ComponentConfig {
	return &guiapi.ComponentConfig{
		Name: "Counter",
		Actions: map[string]guiapi.Callable{
			"Increase": c.Increase,
			"Decrease": c.Decrease,
		},
		Pages: map[string]guiapi.PageFunc{
			"/counter": c.RenderPage,
		},
	}
}

type CounterPage struct {
	Content html.Block
}

func (c *CounterPage) HTML() (html.Block, error) {
	return html.Blocks{
		html.Doctype("html"),
		html.Html(nil,
			html.Head(nil,
				html.Meta(attr.Charset("utf-8")),
				html.Title(nil, html.Text("Guiapi Counter Example")),
				html.Link(attr.Rel("stylesheet").Href("https://cdn.jsdelivr.net/npm/simpledotcss@2.2.0/simple.min.css")),
				html.Link(attr.Rel("stylesheet").Href("/dist/bundle.css")),
			),
			html.Body(nil,
				html.Main(attr.Id("page"), c.Content),
				html.Hr(nil),
				html.A(attr.Href("/"), html.Text("TodoMVC Example")),
				html.A(attr.Href("/reports"), html.Text("Reports Example")),
				html.Script(attr.Src("/dist/bundle.js")),
			),
		),
	}, nil
}

func (c *CounterPage) Update() (*guiapi.Response, error) {
	return guiapi.ReplaceContent("#page", c.Content)
}

func (c *Counter) RenderPage(ctx *guiapi.Context) (guiapi.Page, error) {
	block, err := c.RenderBlock(ctx)
	if err != nil {
		return nil, err
	}
	main := html.Blocks{
		html.H1(nil, html.Text("guiapi")),
		html.P(nil, html.Text("guiapi is a framework for building web applications in Go.")),
		block,
	}
	return &CounterPage{Content: main}, nil
}

func (c *Counter) RenderBlock(ctx *guiapi.Context) (html.Block, error) {
	sess := sessionFromContext(ctx)
	counter, err := c.DB.GetCounter(sess.ID)
	if err != nil {
		return nil, err
	}
	block := html.Div(attr.Id("counter"),
		html.H3(nil, html.Text("Counter")),
		html.P(attr.Id("count"), html.Text(fmt.Sprintf("Current count: %d", counter.Count))),
		html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-action", "Counter.Decrease"), html.Text("-")),
		html.Text(" "),
		html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-action", "Counter.Increase"), html.Text("+")),
	)
	return block, nil
}

func (c *Counter) Increase(ctx *guiapi.Context, args json.RawMessage) (*guiapi.Response, error) {
	sess := sessionFromContext(ctx)
	counter, err := c.DB.GetCounter(sess.ID)
	if err != nil {
		return nil, err
	}
	counter.Count++
	err = c.DB.SetCounter(counter)
	if err != nil {
		return nil, err
	}
	return guiapi.ReplaceContent("#count", html.Text(fmt.Sprintf("Current count: %d", counter.Count)))
}

func (c *Counter) Decrease(ctx *guiapi.Context, args json.RawMessage) (*guiapi.Response, error) {
	sess := sessionFromContext(ctx)
	counter, err := c.DB.GetCounter(sess.ID)
	if err != nil {
		return nil, err
	}
	counter.Count--
	err = c.DB.SetCounter(counter)
	if err != nil {
		return nil, err
	}
	return guiapi.ReplaceContent("#count", html.Text(fmt.Sprintf("Current count: %d", counter.Count)))
}
