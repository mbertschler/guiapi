package guiapi

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func NewBuildOptions(infile, outfile string) BuildOptions {
	return BuildOptions{
		Infile:    infile,
		Outfile:   outfile,
		Bundle:    true,
		Minify:    true,
		Sourcemap: true,
		Log:       true,
	}
}

type BuildOptions struct {
	Infile      string
	Outfile     string
	Bundle      bool
	Minify      bool
	Sourcemap   bool
	Log         bool
	EsbuildArgs []string
}

func BuildOrUseBuiltAssets(options BuildOptions, built fs.FS) (fs.FS, error) {
	dir := filepath.Dir(options.Outfile)

	if EsbuildAvailable() {
		err := BuildAssets(options)
		if err != nil {
			log.Fatal(err)
		}
		return os.DirFS(dir), nil
	}

	if options.Log {
		log.Println("esbuild not available, using prebuilt assets")
	}
	return fs.Sub(built, dir)
}
