package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"github.com/mbertschler/guiapi"
)

func Setup(distFS fs.FS) *guiapi.Server {
	db := NewDB()

	reports := NewReportsComponent(db)
	counter := &Counter{DB: db}
	todo := &TodoList{DB: db}

	// better struct options
	server := guiapi.New(reports.StreamRouter)

	// move into guiapi?
	server.AddFiles("/dist/", http.FS(distFS))

	reports.Register(server)
	counter.Register(server)
	todo.Register(server)

	return server
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

	server := Setup(fs)

	log.Println("listening on localhost:8000")
	err = http.ListenAndServe("localhost:8000", server)
	if err != nil {
		log.Fatal(err)
	}
}
