package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/mbertschler/guiapi"
)

type App struct {
	DB     *DB
	Server *Server
}

func NewApp() *App {
	app := &App{}
	app.DB = NewDB()
	app.Server = NewServer(app.DB)
	return app
}

//go:embed dist/*
var distFS embed.FS

func main() {

	app := NewApp()
	counter := &Counter{App: app}
	app.Server.RegisterComponent(counter)
	app.Server.RegisterPage("/counter", counter.RenderPage)

	registerTodoList(app.Server, app.DB)

	if guiapi.EsbuildAvailable() {
		err := guiapi.BuildAssets()
		if err != nil {
			log.Fatal(err)
		}
		app.Server.Static("/dist/", "./dist")
	} else {
		app.Server.engine.StaticFileFS("/dist/", "/dist/", http.FS(distFS))
	}

	log.Println("listening on localhost:8000")
	err := http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
