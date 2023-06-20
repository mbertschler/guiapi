package main

import (
	"context"
	"encoding/json"

	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/html"
)

func (r *Reports) StreamRouter(ctx context.Context, msg []byte, res chan<- *guiapi.Response) error {
	var stream ReportsStream
	err := json.Unmarshal(msg, &stream)
	if err != nil {
		return err
	}
	if stream.Overview {
		return r.OverviewStream(ctx, res)
	}
	if stream.ID != "" {
		return r.DetailStream(ctx, stream.ID, res)
	}
	return nil
}

func (r *Reports) OverviewStream(ctx context.Context, results chan<- *guiapi.Response) error {
	listener := r.DB.AddGlobalChangeListener(func(change ChangeType, report *Report) {
		out, err := html.RenderMinifiedString(r.allReportsBlock())
		res := guiapi.ReplaceElement("#all-reports", out)
		if err != nil {
			res.Error = &guiapi.Error{Message: err.Error()}
		}
		results <- res
	})

	<-ctx.Done()
	r.DB.RemoveGlobalChangeListener(listener)
	return nil
}

func (r *Reports) DetailStream(ctx context.Context, id string, results chan<- *guiapi.Response) error {
	listener := r.DB.AddIDChangeListener(id, func(change ChangeType, report *Report) {
		out, err := html.RenderMinifiedString(r.singleReportBlock(id))
		res := guiapi.ReplaceElement("#single-report", out)
		if err != nil {
			res.Error = &guiapi.Error{Message: err.Error()}
		}
		results <- res
	})

	<-ctx.Done()
	r.DB.RemoveIDChangeListener(id, listener)
	return nil
}
