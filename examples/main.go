package main

import (
	"embed"
	"flag"
	"log"
	"net/http"

	"github.com/mbertschler/guiapi"
)

//go:embed dist/*
var distEmbedFS embed.FS

func setup() (*guiapi.Server, error) {
	db := NewDB()

	reports := NewReportsComponent(db)
	counter := &Counter{DB: db}
	todo := &TodoList{DB: db}

	options := guiapi.DefaultOptions()
	options.DistFS = distEmbedFS
	options.Assets.Infile = "js/main.js"
	options.Assets.Outfile = "dist/bundle.js"
	options.Assets.EsbuildArgs = []string{"--metafile=dist/meta.json"}

	// better struct options
	server := guiapi.New(options, reports.StreamRouter)

	distFS, err := server.BuildAssets()
	if err != nil {
		return nil, err
	}

	// move into guiapi?
	server.AddFiles("/dist/", http.FS(distFS))

	reports.Register(server)
	counter.Register(server)
	todo.Register(server)

	return server, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var exitAfterBuild bool
	flag.BoolVar(&exitAfterBuild, "build", false, "build assets and exit")
	flag.Parse()

	server, err := setup()
	if err != nil {
		log.Fatal(err)
	}

	if exitAfterBuild {
		return
	}

	log.Println("listening on localhost:8000")
	err = http.ListenAndServe("localhost:8000", server)
	if err != nil {
		log.Fatal(err)
	}
}
