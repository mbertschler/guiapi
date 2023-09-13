package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mbertschler/guiapi"
	"github.com/mbertschler/guiapi/assets"
)

//go:embed dist/*
var prebuiltAssets embed.FS

func assetsFS() (fs.FS, error) {
	if !assets.EsbuildAvailable() {
		// in production, esbuild is not available
		// and assets are compiled into the binary
		assets, err := fs.Sub(prebuiltAssets, "dist")
		if err != nil {
			return nil, err
		}
		return assets, nil
	}

	build := assets.DefaultBuildOptions()
	build.Infile = "js/main.js"
	build.Outfile = "dist/bundle.js"
	build.EsbuildArgs = []string{"--metafile=dist/meta.json"}

	dir := filepath.Dir(build.Outfile)

	err := assets.BuildAssets(build)
	if err != nil {
		return nil, err
	}
	return os.DirFS(dir), nil
}

func setupServer(assetsFS fs.FS) *guiapi.Server {
	db := NewDB()

	reports := NewReportsComponent(db)
	counter := &Counter{DB: db}
	todo := &TodoList{DB: db}

	server := guiapi.New()

	server.AddFiles("/dist/", http.FS(assetsFS))

	reports.Register(server)
	counter.Register(server)
	todo.Register(server)

	return server
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var exitAfterBuild bool
	flag.BoolVar(&exitAfterBuild, "build", false, "build assets and exit")
	flag.Parse()

	fs, err := assetsFS()
	if err != nil {
		log.Fatal(err)
	}

	if exitAfterBuild {
		log.Println("built assets, exit after build flag provided")
		return
	}

	server := setupServer(fs)

	log.Println("listening on localhost:8000")
	err = http.ListenAndServe("localhost:8000", server)
	if err != nil {
		log.Fatal(err)
	}
}
