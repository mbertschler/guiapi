package main

import (
	"context"
	"encoding/json"

	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/guiapi/api"
	"github.com/mbertschler/html"
)

func (r *Reports) Stream(ctx context.Context, msg json.RawMessage, res chan<- *guiapi.Update) error {
	var stream ReportsStream
	err := json.Unmarshal(msg, &stream)
	if err != nil {
		return err
	}
	if stream.Overview {
		return r.overviewStream(ctx, res)
	}
	if stream.ID != "" {
		return r.detailStream(ctx, stream.ID, res)
	}
	return nil
}

func (r *Reports) overviewStream(ctx context.Context, results chan<- *guiapi.Update) error {
	listener := r.DB.AddGlobalChangeListener(func(change ChangeType, report *Report) {
		out, err := html.RenderMinifiedString(r.allReportsBlock())
		res := guiapi.ReplaceElement("#all-reports", out)
		if err != nil {
			res.Error = &api.Error{Message: err.Error()}
		}
		results <- res
	})

	<-ctx.Done()
	r.DB.RemoveGlobalChangeListener(listener)
	return nil
}

func (r *Reports) detailStream(ctx context.Context, id string, results chan<- *guiapi.Update) error {
	listener := r.DB.AddIDChangeListener(id, func(change ChangeType, report *Report) {
		out, err := html.RenderMinifiedString(r.singleReportBlock(id))
		res := guiapi.ReplaceElement("#single-report", out)
		if err != nil {
			res.Error = &api.Error{Message: err.Error()}
		}
		results <- res
	})

	<-ctx.Done()
	r.DB.RemoveIDChangeListener(id, listener)
	return nil
}
