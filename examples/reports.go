package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mbertschler/blocks/html"
	"github.com/mbertschler/blocks/html/attr"
	"github.com/mbertschler/guiapi"
)

type ReportsDB struct {
	lock    sync.Mutex
	reports map[string]*Report
}

func (r *ReportsDB) Get(id string) *Report {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.reports[id]
}

func (r *ReportsDB) All() []*Report {
	r.lock.Lock()
	defer r.lock.Unlock()
	out := make([]*Report, 0, len(r.reports))
	for _, report := range r.reports {
		out = append(out, report)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Started.Before(out[j].Started)
	})
	return out
}

func (r *ReportsDB) Update(fn func(reports map[string]*Report)) {
	r.lock.Lock()
	defer r.lock.Unlock()
	fn(r.reports)
}

func (r *ReportsDB) Set(report *Report) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.reports[report.ID] = report
}

func (r *ReportsDB) Delete(id string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.reports, id)
}

func NewReportsComponent() *Reports {
	return &Reports{
		DB: &ReportsDB{
			reports: make(map[string]*Report),
		},
	}
}

type Report struct {
	ID      string
	Started time.Time
	Status  string
}

const (
	ReportStatusStarted   = "started"
	ReportStatusFinished  = "finished"
	ReportStatusCancelled = "cancelled"
)

type Reports struct {
	DB *ReportsDB
}

func (r *Reports) Component() *guiapi.ComponentConfig {
	return &guiapi.ComponentConfig{
		Name: "Reports",
		Actions: map[string]guiapi.Callable{
			"Start":   ContextCallable(r.Start),
			"Cancel":  ContextCallable(r.Cancel),
			"Refresh": ContextCallable(r.Refresh),
		},
		Pages: map[string]guiapi.PageFunc{
			"/reports":    r.IndexPage,
			"/report/:id": r.ReportPage,
		},
	}
}

type ReportsPage struct {
	Content html.Block
}

func (r *ReportsPage) HTML() (html.Block, error) {
	return html.Blocks{
		html.Doctype("html"),
		html.Html(nil,
			html.Head(nil,
				html.Meta(attr.Charset("utf-8")),
				html.Title(nil, html.Text("Blocks")),
				html.Link(attr.Rel("stylesheet").Href("https://cdn.jsdelivr.net/npm/simpledotcss@2.2.0/simple.min.css")),
				html.Link(attr.Rel("stylesheet").Href("/dist/bundle.css")),
			),
			html.Body(nil,
				r.Content,
				html.Hr(nil),
				html.A(attr.Href("/"), html.Text("TodoMVC Example")),
				html.A(attr.Href("/counter"), html.Text("Counter Example")),
				html.Script(attr.Src("/dist/bundle.js")),
			),
		),
	}, nil
}

func (r *ReportsPage) Update() (*guiapi.Response, error) {
	return guiapi.ReplaceElement("#reports", r.Content)
}

func (r *Reports) IndexPage(ctx *guiapi.Context) (guiapi.Page, error) {
	main, err := r.indexBlock(ctx)
	if err != nil {
		return nil, err
	}
	return &ReportsPage{Content: main}, nil
}

func (r *Reports) indexBlock(ctx *guiapi.Context) (html.Block, error) {
	main := html.Main(attr.Id("reports"),
		html.H1(nil, html.Text("Reports")),
		html.P(nil, html.Text("This is a demo for reports that take a long time to complete.")),
		html.H3(nil, html.Text("All Reports")),
		r.allReportsBlock(ctx),
		html.P(nil, html.Text("Refreshing is very slow, it takes 2 seconds. That's why we show you a spinner.")),
		html.Div(nil,
			html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-func", "Reports.onRefresh"), html.Text("Refresh")),
			html.Span(attr.Id("refresh-spinner").Class("spinner").Style("display:none;")),
		),
		html.H3(nil, html.Text("New Report")),
		html.Div(nil, html.Input(attr.Class("new-report").Name("id").Placeholder("Give the new report a name").Type("text"))),
		html.Div(nil, html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-action", "Reports.Start").Attr("ga-values", ".new-report"), html.Text("Start"))),
	)
	return main, nil
}

func (r *Reports) allReportsBlock(ctx *guiapi.Context) html.Block {
	reports := r.DB.All()
	var items html.Blocks
	for _, report := range reports {
		text := fmt.Sprintf(": %s %v", report.Status, report.Started)
		items.Add(html.Li(nil,
			html.A(attr.Href("/report/"+report.ID).Class("ga").Attr("ga-link", nil), html.Text(report.ID)),
			html.Text(text),
		))
	}

	block := html.Ul(attr.Id("all-reports"), items)
	if len(reports) == 0 {
		block = html.P(attr.Id("all-reports"), html.Text("No reports yet."))
	}
	return block
}

func (r *Reports) ReportPage(ctx *guiapi.Context) (guiapi.Page, error) {
	id := ctx.Params.ByName("id")
	report := r.DB.Get(id)
	main := html.Main(attr.Id("reports"),
		html.A(attr.Href("/reports").Class("ga").Attr("ga-link", nil), html.Text("< All Reports")),
		html.H1(nil, html.Text("Report "+id)),
		html.P(nil, html.Text("Blocks is a framework for building web applications in Go.")),
		html.P(nil, html.Text(fmt.Sprintf("%q: %+v", id, report))),
	)
	return &ReportsPage{Content: main}, nil
}

type ReportsArgs struct {
	ID string `json:"id"`
}

func (r *Reports) Start(ctx *Context, args *ReportsArgs) (*guiapi.Response, error) {
	report := &Report{
		ID:      args.ID,
		Started: time.Now(),
		Status:  ReportStatusStarted,
	}
	r.DB.Set(report)
	return guiapi.ReplaceElement("#all-reports", r.allReportsBlock(ctx.Ctx))
}

func (r *Reports) Cancel(ctx *Context, args *ReportsArgs) (*guiapi.Response, error) {
	r.DB.Update(func(reports map[string]*Report) {
		report := reports[args.ID]
		if report.Status == ReportStatusStarted {
			report.Status = ReportStatusCancelled
		}
	})
	return guiapi.ReplaceElement("#all-reports", r.allReportsBlock(ctx.Ctx))
}

func (r *Reports) Refresh(ctx *Context, args *NoArgs) (*guiapi.Response, error) {
	time.Sleep(2 * time.Second)
	return guiapi.ReplaceElement("#all-reports", r.allReportsBlock(ctx.Ctx))
}
