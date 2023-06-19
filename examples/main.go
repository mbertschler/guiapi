package main

import (
	"embed"
	"flag"
	"log"
	"net/http"

	"github.com/mbertschler/guiapi"
)

type App struct {
	DB      *DB
	Server  *guiapi.Server
	Reports *Reports
}

func NewApp() *App {
	app := &App{}
	app.DB = NewDB()
	app.Reports = NewReportsComponent()
	app.Server = guiapi.New(app.DB.sessionMiddleware, app.Reports.StreamRouter)
	return app
}

//go:embed dist/*
var distEmbedFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var exitAfterBuild bool
	flag.BoolVar(&exitAfterBuild, "build", false, "build assets and exit")
	flag.Parse()

	options := guiapi.NewBuildOptions("js/main.js", "dist/bundle.js")
	options.EsbuildArgs = []string{"--metafile=dist/meta.json"}
	fs, err := guiapi.BuildOrUseBuiltAssets(options, distEmbedFS)
	if err != nil {
		log.Fatal(err)
	}

	if exitAfterBuild {
		return
	}

	app := NewApp()

	app.Server.RegisterComponent(&Counter{DB: app.DB})
	app.Server.RegisterComponent(&TodoList{DB: app.DB})
	app.Server.RegisterComponent(app.Reports)

	app.Server.ServeFiles("/dist/", http.FS(fs))

	log.Println("listening on localhost:8000")
	err = http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
