package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/mbertschler/guiapi"
)

type App struct {
	DB     *DB
	Server *guiapi.Server
}

func NewApp() *App {
	app := &App{}
	app.DB = NewDB()
	app.Server = guiapi.New()
	return app
}

//go:embed dist/*
var distEmbedFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := NewApp()
	app.Server.SessionMiddleware(app.DB.sessionMiddleware)

	app.Server.RegisterComponent(&Counter{App: app})
	app.Server.RegisterComponent(&TodoList{DB: app.DB})

	if guiapi.EsbuildAvailable() {
		err := guiapi.BuildAssets()
		if err != nil {
			log.Fatal(err)
		}
		app.Server.StaticDir("/dist/", "./dist")
	} else {
		fs, err := fs.Sub(distEmbedFS, "dist")
		if err != nil {
			log.Fatal(err)
		}
		app.Server.StaticFS("/dist/", http.FS(fs))
	}

	log.Println("listening on localhost:8000")
	err := http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
