package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/mbertschler/guiapi"
)

type App struct {
	DB     *DB
	Server *guiapi.Server
}

func NewApp() *App {
	app := &App{}
	app.DB = NewDB()
	app.Server = guiapi.New(app.DB.sessionMiddleware)
	return app
}

//go:embed dist/*
var distEmbedFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) > 1 && os.Args[1] == "build" {
		err := guiapi.BuildAssets()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	app := NewApp()

	app.Server.RegisterComponent(&Counter{DB: app.DB})
	app.Server.RegisterComponent(&TodoList{DB: app.DB})
	app.Server.RegisterComponent(NewReportsComponent())

	var dist fs.FS = distEmbedFS
	if guiapi.EsbuildAvailable() {
		err := guiapi.BuildAssets()
		if err != nil {
			log.Fatal(err)
		}
		dist = os.DirFS("dist")
	} else {
		var err error
		dist, err = fs.Sub(dist, "dist")
		if err != nil {
			log.Fatal(err)
		}
	}
	app.Server.ServeFiles("/dist/", http.FS(dist))

	log.Println("listening on localhost:8000")
	err := http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
