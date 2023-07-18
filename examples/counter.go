package main

import (
	"fmt"
	"io"

	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

type Counter struct {
	*DB
}

func (c *Counter) Register(s *guiapi.Server) {
	s.AddPage("/counter", c.RenderPage)

	s.AddAction("Counter.Increase", c.Increase)
	s.AddAction("Counter.Decrease", c.Decrease)
}

type CounterPage struct {
	Content html.Block
}

func (c *CounterPage) WriteHTML(w io.Writer) error {
	block := html.Blocks{
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
	}
	return html.RenderMinified(w, block)
}

func (c *CounterPage) Update() (*guiapi.Response, error) {
	out, err := html.RenderMinifiedString(c.Content)
	return guiapi.ReplaceContent("#page", out), err
}

func (c *Counter) RenderPage(ctx *guiapi.PageCtx) (guiapi.Page, error) {
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

func (c *Counter) RenderBlock(ctx *guiapi.PageCtx) (html.Block, error) {
	sess := c.DB.Session(ctx.Writer, ctx.Request)
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

func (c *Counter) Increase(ctx *guiapi.ActionCtx) (*guiapi.Response, error) {
	sess := c.DB.Session(ctx.Writer, ctx.Request)
	counter, err := c.DB.GetCounter(sess.ID)
	if err != nil {
		return nil, err
	}
	counter.Count++
	err = c.DB.SetCounter(counter)
	if err != nil {
		return nil, err
	}

	out, err := html.RenderMinifiedString(html.Text(fmt.Sprintf("Current count: %d", counter.Count)))
	return guiapi.ReplaceContent("#count", out), err
}

func (c *Counter) Decrease(ctx *guiapi.ActionCtx) (*guiapi.Response, error) {
	sess := c.DB.Session(ctx.Writer, ctx.Request)
	counter, err := c.DB.GetCounter(sess.ID)
	if err != nil {
		return nil, err
	}
	counter.Count--
	err = c.DB.SetCounter(counter)
	if err != nil {
		return nil, err
	}

	out, err := html.RenderMinifiedString(html.Text(fmt.Sprintf("Current count: %d", counter.Count)))
	return guiapi.ReplaceContent("#count", out), err
}
