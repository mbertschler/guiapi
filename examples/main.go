package main

import (
	"embed"
	"flag"
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

	reports := NewReportsComponent()
	counter := &Counter{DB: app.DB}
	todo := &TodoList{DB: app.DB}

	// functional options
	app.Server = guiapi.New(app.DB.sessionMiddleware, reports.StreamRouter)

	reports.Register(app.Server)
	counter.Register(app.Server)
	todo.Register(app.Server)

	return app
}

//go:embed dist/*
var distEmbedFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var exitAfterBuild bool
	flag.BoolVar(&exitAfterBuild, "build", false, "build assets and exit")
	flag.Parse()

	// move as options into guiapi New() call, or not?
	// what about exitAfterBuild?
	options := guiapi.NewBuildOptions("js/main.js", "dist/bundle.js")
	options.EsbuildArgs = []string{"--metafile=dist/meta.json"}

	// BuildAssets(), but called on the whole server after configuring?
	fs, err := guiapi.BuildOrUseBuiltAssets(options, distEmbedFS)
	if err != nil {
		log.Fatal(err)
	}

	if exitAfterBuild {
		return
	}

	app := NewApp()

	// move into guiapi
	app.Server.ServeFiles("/dist/", http.FS(fs))

	log.Println("listening on localhost:8000")
	err = http.ListenAndServe("localhost:8000", app.Server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
