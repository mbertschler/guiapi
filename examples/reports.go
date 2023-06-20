package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

type ChangeType int

const (
	ChangeCreate ChangeType = 1
	ChangeUpdate ChangeType = 2
	ChangeDelete ChangeType = 3
)

func (c ChangeType) String() string {
	switch c {
	case ChangeCreate:
		return "create"
	case ChangeUpdate:
		return "update"
	case ChangeDelete:
		return "delete"
	}
	return "unknown"
}

type ReportsChange func(change ChangeType, report *Report)

type ReportsDB struct {
	transaction   sync.Mutex
	lock          sync.Mutex
	globalUpdates map[int64]ReportsChange
	idUpdates     map[string]map[int64]ReportsChange
	reports       map[string]*Report
}

func (r *ReportsDB) AddGlobalChangeListener(fn ReportsChange) int64 {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.globalUpdates == nil {
		r.globalUpdates = make(map[int64]ReportsChange)
	}
	listenerID := time.Now().UnixNano()
	r.globalUpdates[listenerID] = fn
	return listenerID
}

func (r *ReportsDB) RemoveGlobalChangeListener(id int64) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.globalUpdates, id)
}

func (r *ReportsDB) AddIDChangeListener(id string, fn ReportsChange) int64 {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.idUpdates == nil {
		r.idUpdates = make(map[string]map[int64]ReportsChange)
	}
	if r.idUpdates[id] == nil {
		r.idUpdates[id] = make(map[int64]ReportsChange)
	}
	listenerID := time.Now().UnixNano()
	r.idUpdates[id][listenerID] = fn
	return listenerID
}

func (r *ReportsDB) RemoveIDChangeListener(id string, listenerID int64) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.idUpdates[id], listenerID)
}

func (r *ReportsDB) notify(change ChangeType, report *Report) {
	r.lock.Unlock()
	defer r.lock.Lock()
	for _, fn := range r.globalUpdates {
		fn(change, report)
	}
	for _, fn := range r.idUpdates[report.ID] {
		fn(change, report)
	}
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

func (r *ReportsDB) Transaction(fn func() error) error {
	r.transaction.Lock()
	defer r.transaction.Unlock()
	return fn()
}

func (r *ReportsDB) Create(report *Report) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.reports[report.ID] != nil {
		return fmt.Errorf("report with id %s already exists", report.ID)
	}
	r.reports[report.ID] = report
	r.notify(ChangeCreate, report)
	return nil
}

func (r *ReportsDB) Update(report *Report) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.reports[report.ID] == nil {
		return fmt.Errorf("report with id %s doesn't exists", report.ID)
	}
	r.reports[report.ID] = report
	r.notify(ChangeUpdate, report)
	return nil
}

func (r *ReportsDB) Delete(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	report := r.reports[id]
	if report == nil {
		return fmt.Errorf("report with id %s doesn't exists", report.ID)
	}
	delete(r.reports, id)
	r.notify(ChangeDelete, report)
	return nil
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
			"Start":     ContextCallable(r.Start),
			"Cancel":    ContextCallable(r.Cancel),
			"Refresh":   ContextCallable(r.Refresh),
			"SomeError": ContextCallable(r.SomeError),
		},
		Pages: map[string]guiapi.PageFunc{
			"/reports":    r.IndexPage,
			"/report/:id": r.ReportPage,
		},
	}
}

type ReportsPage struct {
	Content html.Block
	Stream  ReportsStream
}

type ReportsStream struct {
	ID       string
	Overview bool
}

func (r *ReportsPage) WriteHTML(w io.Writer) error {
	stream, err := json.Marshal(r.Stream)
	if err != nil {
		return err
	}
	block := html.Blocks{
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
				html.Div(attr.Id("error-box"), html.Text("there is an error message")),
				html.Script(nil, html.JS("var stream = "+string(stream)+";")), // before bundle, otherwise it isn't defined
				html.Script(attr.Src("/dist/bundle.js")),
			),
		),
	}
	return html.RenderMinified(w, block)
}

func (r *ReportsPage) Update() (*guiapi.Response, error) {
	out, err := html.RenderMinifiedString(r.Content)
	res := guiapi.ReplaceElement("#reports", out)
	res.Stream = r.Stream
	return res, err
}

func (r *Reports) IndexPage(ctx *guiapi.Context) (guiapi.Page, error) {
	main, err := r.indexBlock(ctx)
	if err != nil {
		return nil, err
	}
	return &ReportsPage{
		Content: main,
		Stream:  ReportsStream{Overview: true},
	}, nil
}

func (r *Reports) indexBlock(ctx *guiapi.Context) (html.Block, error) {
	main := html.Main(attr.Id("reports"),
		html.H1(nil, html.Text("Reports")),
		html.P(nil, html.Text("This is a demo for reports that take a long time to complete.")),
		html.H3(nil, html.Text("All Reports")),
		r.allReportsBlock(),
		html.P(nil, html.Text("Refreshing is very slow, it takes 2 seconds. That's why we show you a spinner.")),
		html.Div(nil,
			html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-func", "Reports.onRefresh"), html.Text("Refresh")),
			html.Span(attr.Id("refresh-spinner").Class("spinner").Style("display:none;")),
		),
		html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-action", "Reports.SomeError"), html.Text("Fake Error")),
		html.H3(nil, html.Text("New Report")),
		html.Div(nil, html.Input(attr.Class("new-report").Name("id").Placeholder("Give the new report a name").Type("text"))),
		html.Div(nil, html.Button(attr.Class("ga").Attr("ga-on", "click").Attr("ga-action", "Reports.Start").Attr("ga-values", ".new-report"), html.Text("Start"))),
	)
	return main, nil
}

func (r *Reports) allReportsBlock() html.Block {
	reports := r.DB.All()
	var items html.Blocks
	for _, report := range reports {
		text := fmt.Sprintf(": %s %s", report.Status, report.Started.Format(time.DateTime))
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
	main := html.Main(attr.Id("reports"),
		html.A(attr.Href("/reports").Class("ga").Attr("ga-link", nil), html.Text("< All Reports")),
		r.singleReportBlock(id),
	)
	return &ReportsPage{
		Content: main,
		Stream:  ReportsStream{ID: id},
	}, nil
}

func (r *Reports) singleReportBlock(id string) html.Block {
	report := r.DB.Get(id)
	if report == nil {
		return html.Div(attr.Id("single-report"), html.H1(nil, html.Text("Report "+id)),
			html.P(nil, html.Text(fmt.Sprintf("Report with ID %q doesn't exist", id))),
		)
	}
	return html.Div(attr.Id("single-report"), html.H1(nil, html.Text("Report "+id)),
		html.P(nil, html.Text("Blocks is a framework for building web applications in Go.")),
		html.Div(nil, html.Text(fmt.Sprintf("ID: %q", id))),
		html.Div(nil, html.Text(fmt.Sprintf("Status: %s", report.Status))),
		html.Div(nil, html.Text(fmt.Sprintf("Started: %s", report.Started.Format(time.DateTime)))),
	)
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
	err := r.DB.Create(report)
	if err != nil {
		return nil, err
	}
	go func() {
		// run the actual report
		time.Sleep(5 * time.Second)
		err := r.DB.Transaction(func() error {
			report := r.DB.Get(args.ID)
			report.Status = ReportStatusFinished
			log.Println("finished report", report.ID)
			return r.DB.Update(report)
		})
		if err != nil {
			log.Println(err)
		}
	}()
	ctx.Ctx.Params = httprouter.Params{httprouter.Param{Key: "id", Value: report.ID}}
	page, err := r.ReportPage(ctx.Ctx)
	if err != nil {
		return nil, err
	}
	update, err := page.Update()
	update.URL = "/report/" + report.ID
	return update, err
}

func (r *Reports) Cancel(ctx *Context, args *ReportsArgs) (*guiapi.Response, error) {
	err := r.DB.Transaction(func() error {
		report := r.DB.Get(args.ID)
		if report.Status == ReportStatusStarted {
			report.Status = ReportStatusCancelled
		}
		return r.DB.Update(report)
	})
	if err != nil {
		return nil, err
	}
	out, err := html.RenderMinifiedString(r.allReportsBlock())
	return guiapi.ReplaceElement("#all-reports", out), err
}

func (r *Reports) Refresh(ctx *Context, args *NoArgs) (*guiapi.Response, error) {
	time.Sleep(2 * time.Second)
	out, err := html.RenderMinifiedString(r.allReportsBlock())
	return guiapi.ReplaceElement("#all-reports", out), err
}

func (r *Reports) SomeError(ctx *Context, args *NoArgs) (*guiapi.Response, error) {
	return nil, errors.New("something bad happened (not really)")
}
