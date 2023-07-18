package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

type Page struct {
	*guiapi.PageCtx
	Sess *Session
}

type Action struct {
	*guiapi.ActionCtx
	Sess  *Session
	State TodoListState
}

type ActionFunc[T any] func(c *Action, args *T) (*guiapi.Response, error)

func ContextAction[T any](db *DB, fn ActionFunc[T]) guiapi.ActionFunc {
	return func(c *guiapi.ActionCtx) (*guiapi.Response, error) {
		var input T
		if c.Args != nil {
			err := json.Unmarshal(c.Args, &input)
			if err != nil {
				return nil, err
			}
		}

		ctx := &Action{
			ActionCtx: c,
			Sess:      db.Session(c.Writer, c.Request),
		}

		err := json.Unmarshal(c.State, &ctx.State)
		if err != nil {
			return nil, err
		}
		return fn(ctx, &input)
	}
}

type PageFunc func(c *Page) (guiapi.Page, error)

func PageWrapper(db *DB) func(PageFunc) guiapi.PageFunc {
	return func(pf PageFunc) guiapi.PageFunc {
		return func(ctx *guiapi.PageCtx) (guiapi.Page, error) {
			sess := db.Session(ctx.Writer, ctx.Request)
			return pf(&Page{PageCtx: ctx, Sess: sess})
		}
	}
}

type TodoPage struct {
	Content html.Block
	State   TodoListState
}

func (t *TodoPage) WriteHTML(w io.Writer) error {
	stateJSON, err := json.Marshal(t.State)
	if err != nil {
		return err
	}
	block := html.Blocks{
		html.Doctype("html"),
		html.Html(attr.Lang("en"),
			html.Head(nil,
				html.Meta(attr.Charset("utf-8")),
				html.Meta(attr.Name("viewport").Content("width=device-width, initial-scale=1")),
				html.Title(nil, html.Text("Guiapi • TodoMVC")),
				html.Link(attr.Rel("stylesheet").Href("https://cdn.jsdelivr.net/npm/todomvc-app-css@2.4.2/index.min.css")),
				html.Link(attr.Rel("stylesheet").Href("/dist/bundle.css")),
			),
			html.Body(nil,
				html.Main(attr.Id("page"), t.Content),
				html.Elem("footer", attr.Class("info"),
					html.P(nil, html.Text("Double-click to edit a todo")),
					html.P(nil, html.Text("Template by "), html.A(attr.Href("http://sindresorhus.com"), html.Text("Sindre Sorhus"))),
					html.P(nil, html.Text("Created by "), html.A(attr.Href("https://github.com/mbertschler"), html.Text("Martin Bertschler"))),
					html.P(nil, html.Text("Part of "), html.A(attr.Href("http://todomvc.com"), html.Text("TodoMVC"))),
					html.Hr(nil),
					html.P(attr.Class("biglink"), html.A(attr.Href("/counter"), html.Text("Counter Example"))),
					html.P(attr.Class("biglink"), html.A(attr.Href("/reports"), html.Text("Reports Example"))),
				),
				html.Script(nil, html.JS("var state = "+string(stateJSON)+";")),
				html.Script(attr.Src("/dist/bundle.js")),
			),
		),
	}
	return html.RenderMinified(w, block)
}

func (t *TodoPage) Update() (*guiapi.Response, error) {
	out, err := html.RenderMinifiedString(t.Content)
	if err != nil {
		return nil, err
	}
	res := guiapi.ReplaceContent("#page", out)
	res.State = t.State
	return res, nil
}

type TodoList struct {
	*DB
}

func (t *TodoList) Register(s *guiapi.Server) {
	s.AddPage("/", t.RenderFullPage(TodoListPageAll))
	s.AddPage("/active", t.RenderFullPage(TodoListPageActive))
	s.AddPage("/completed", t.RenderFullPage(TodoListPageCompleted))

	s.AddAction("TodoList.NewTodo", ContextAction(t.DB, t.NewTodo))
	s.AddAction("TodoList.ToggleItem", ContextAction(t.DB, t.ToggleItem))
	s.AddAction("TodoList.ToggleAll", ContextAction(t.DB, t.ToggleAll))
	s.AddAction("TodoList.DeleteItem", ContextAction(t.DB, t.DeleteItem))
	s.AddAction("TodoList.ClearCompleted", ContextAction(t.DB, t.ClearCompleted))
	s.AddAction("TodoList.EditItem", ContextAction(t.DB, t.EditItem))
	s.AddAction("TodoList.UpdateItem", ContextAction(t.DB, t.UpdateItem))
}

const (
	TodoListPageAll       = "all"
	TodoListPageActive    = "active"
	TodoListPageCompleted = "completed"
)

type TodoListState struct {
	Page string
}

type TodoListProps struct {
	Page       string
	Todos      *StoredTodo
	EditItemID int
}

func (t *TodoList) RenderFullPage(page string) guiapi.PageFunc {
	wrap := PageWrapper(t.DB)
	return wrap(func(ctx *Page) (guiapi.Page, error) {
		content, err := t.renderPageContent(ctx, page)
		if err != nil {
			return nil, err
		}
		return &TodoPage{Content: content}, nil
	})
}

func (t *TodoList) renderPageContent(ctx *Page, page string) (html.Block, error) {
	props, err := t.todoListProps(ctx.Sess, page)
	if err != nil {
		return nil, err
	}
	return t.renderBlock(props)
}

type PageArgs struct {
	Page string
}

func (t *TodoList) renderBlock(props *TodoListProps) (html.Block, error) {
	var main, footer html.Block
	if len(props.Todos.Items) > 0 {
		var err error
		main, err = t.renderMainBlock(props.Todos, props.Page, props.EditItemID)
		if err != nil {
			return nil, err
		}
		footer, err = t.renderFooterBlock(props.Todos, props.Page)
		if err != nil {
			return nil, err
		}
	}

	block := html.Elem("section", attr.Class("todoapp"),
		html.Elem("header", attr.Class("header"),
			html.H1(nil, html.Text("todos")),
			html.Input(attr.Class("new-todo ga").Name("new-todo").Placeholder("What needs to be done?").
				Autofocus("").Attr("ga-on", "keydown").Attr("ga-func", "newTodoKeydown")),
		),
		main,
		footer,
	)
	return block, nil
}

func (t *TodoList) renderMainBlock(todos *StoredTodo, page string, editItemID int) (html.Block, error) {
	items := html.Blocks{}
	for _, item := range todos.Items {
		if page == TodoListPageActive && item.Done {
			continue
		}
		if page == TodoListPageCompleted && !item.Done {
			continue
		}
		items.Add(t.renderItem(&item, editItemID))
	}
	main := html.Elem("section", attr.Class("main"),
		html.Input(attr.Id("toggle-all").Class("toggle-all").Type("checkbox")),
		html.Label(attr.Class("ga").For("toggle-all").Attr("ga-on", "click").Attr("ga-action", "TodoList.ToggleAll"),
			html.Text("Mark all as complete")),
		html.Ul(attr.Class("todo-list"),
			items,
		),
	)
	return main, nil
}

func (t *TodoList) renderItem(item *StoredTodoItem, editItemID int) html.Block {
	if item.ID == editItemID {
		return t.renderItemEdit(item, editItemID)
	}

	liAttrs := attr.Attr("ga-on", "dblclick").
		Attr("ga-action", "TodoList.EditItem").
		Attr("ga-args", fmt.Sprintf(`{"id":%d}`, item.ID))
	inputAttrs := attr.Class("toggle ga").Type("checkbox").
		Attr("ga-on", "click").Attr("ga-action", "TodoList.ToggleItem").
		Attr("ga-args", fmt.Sprintf(`{"id":%d}`, item.ID))
	if item.Done {
		liAttrs = liAttrs.Class("completed ga")
		inputAttrs = inputAttrs.Checked("")
	} else {
		liAttrs = liAttrs.Class("active ga")
	}

	id := fmt.Sprintf("todo-%d", item.ID)
	li := html.Li(liAttrs,
		html.Div(attr.Class("view"),
			html.Input(inputAttrs.Id(id)),
			html.Label(attr.For(id), html.Text(item.Text)),
			html.Button(attr.Class("destroy ga").
				Attr("ga-on", "click").Attr("ga-action", "TodoList.DeleteItem").
				Attr("ga-args", fmt.Sprintf(`{"id":%d}`, item.ID))),
		),
	)
	return li
}

func (t *TodoList) renderItemEdit(item *StoredTodoItem, editItemID int) html.Block {
	li := html.Li(attr.Class("editing"),
		html.Div(attr.Class("view"),
			html.Input(attr.Class("edit ga").Attr("ga-init", "initEdit").
				Attr("ga-args", fmt.Sprintf(`{"id":%d}`, item.ID)).Value(item.Text)),
		),
	)
	return li
}

func (t *TodoList) renderFooterBlock(todos *StoredTodo, page string) (html.Block, error) {
	var allClass, activeClass, completedClass string
	switch page {
	case TodoListPageAll:
		allClass = "selected"
	case TodoListPageActive:
		activeClass = "selected"
	case TodoListPageCompleted:
		completedClass = "selected"
	default:
		allClass = "selected"
	}

	leftCount := 0
	someDone := false
	for _, item := range todos.Items {
		if !item.Done {
			leftCount++
		} else {
			someDone = true
		}
	}
	itemsLeftText := " items left"
	if leftCount == 1 {
		itemsLeftText = " item left"
	}

	var clearCompletedButton html.Block
	if someDone {
		clearCompletedButton = html.Button(attr.Class("clear-completed ga").Attr("ga-on", "click").Attr("ga-action", "TodoList.ClearCompleted"),
			html.Text("Clear completed"))
	}

	footer := html.Elem("footer", attr.Class("footer"),
		html.Span(attr.Class("todo-count"),
			html.Strong(nil, html.Text(fmt.Sprint(leftCount))),
			html.Text(itemsLeftText),
		),
		html.Ul(attr.Class("filters"),
			html.Li(nil,
				html.A(attr.Class(allClass+" ga").Href("/").Attr("ga-link", nil), html.Text("All")),
			),
			html.Li(nil,
				html.A(attr.Class(activeClass+" ga").Href("/active").Attr("ga-link", nil), html.Text("Active")),
			),
			html.Li(nil,
				html.A(attr.Class(completedClass+" ga").Href("/completed").Attr("ga-link", nil), html.Text("Completed")),
			),
		),
		clearCompletedButton,
	)
	return footer, nil
}

type NewTodoArgs struct {
	Text string `json:"text"`
}

func (t *TodoList) NewTodo(ctx *Action, input *NewTodoArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		var highestID int
		for _, item := range todos.Items {
			if item.ID > highestID {
				highestID = item.ID
			}
		}
		input.Text = strings.TrimSpace(input.Text)
		todos.Items = append(todos.Items, StoredTodoItem{ID: highestID + 1, Text: input.Text})
		return t.DB.SetTodo(todos)
	})
}

func (t *TodoList) ToggleItem(ctx *Action, args *IDArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		for i, item := range todos.Items {
			if item.ID == args.ID {
				todos.Items[i].Done = !todos.Items[i].Done
			}
		}
		return t.DB.SetTodo(todos)
	})
}

func (t *TodoList) ToggleAll(ctx *Action, args *NoArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		allDone := true
		for _, item := range todos.Items {
			if !item.Done {
				allDone = false
				break
			}
		}

		for i := range todos.Items {
			todos.Items[i].Done = !allDone
		}
		return t.DB.SetTodo(todos)
	})
}

func (t *TodoList) DeleteItem(ctx *Action, args *IDArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		var newItems []StoredTodoItem
		for _, item := range todos.Items {
			if item.ID == args.ID {
				continue
			}
			newItems = append(newItems, item)
		}
		todos.Items = newItems
		return t.DB.SetTodo(todos)
	})
}

type NoArgs struct{}

func (t *TodoList) ClearCompleted(ctx *Action, _ *NoArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		var newItems []StoredTodoItem
		for _, item := range todos.Items {
			if item.Done {
				continue
			}
			newItems = append(newItems, item)
		}
		todos.Items = newItems
		return t.DB.SetTodo(todos)
	})
}

type IDArgs struct {
	ID int `json:"id"`
}

func (t *TodoList) EditItem(ctx *Action, args *IDArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, _ *StoredTodo) error {
		props.EditItemID = args.ID
		return nil
	})
}

type UpdateItemArgs struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func (t *TodoList) UpdateItem(ctx *Action, args *UpdateItemArgs) (*guiapi.Response, error) {
	return t.updateTodoList(ctx, func(props *TodoListProps, todos *StoredTodo) error {
		for i, item := range todos.Items {
			if item.ID == args.ID {
				todos.Items[i].Text = strings.TrimSpace(args.Text)
			}
		}
		return t.DB.SetTodo(todos)
	})
}

func (t *TodoList) updateTodoList(ctx *Action, fn func(*TodoListProps, *StoredTodo) error) (*guiapi.Response, error) {
	props, err := t.todoListProps(ctx.Sess, ctx.State.Page)
	if err != nil {
		return nil, err
	}

	err = fn(props, props.Todos)
	if err != nil {
		return nil, err
	}

	appBlock, err := t.renderBlock(props)
	if err != nil {
		return nil, err
	}
	out, err := html.RenderMinifiedString(appBlock)
	if err != nil {
		return nil, err
	}
	return guiapi.ReplaceContent(".todoapp", out), nil
}

func (t *TodoList) todoListProps(sess *Session, page string) (*TodoListProps, error) {
	todos, err := t.DB.GetTodo(sess.ID)
	if err != nil {
		return nil, err
	}

	return &TodoListProps{
		Page:  page,
		Todos: todos,
	}, nil
}
