package main

import (
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

func main() {
	if guiapi.EsbuildAvailable() {
		err := guiapi.BuildAssets()
		if err != nil {
			log.Fatal(err)
		}
	}

	app := NewApp()
	counter := &Counter{App: app}
	app.Server.RegisterComponent(counter)
	app.Server.RegisterPage("/counter", counter.RenderPage)

	registerTodoList(app.Server, app.DB)

	app.Server.Static("/dist/", "./dist")

	log.Println("listening on localhost:8000")
	err := http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
