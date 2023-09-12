package guiapi

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/mbertschler/guiapi/assets"
)

func NewBuildOptions(infile, outfile string) assets.BuildOptions {
	return assets.BuildOptions{
		Infile:    infile,
		Outfile:   outfile,
		Bundle:    true,
		Minify:    true,
		Sourcemap: true,
		Log:       true,
	}
}

func BuildOrUseBuiltAssets(options assets.BuildOptions, built fs.FS) (fs.FS, error) {
	dir := filepath.Dir(options.Outfile)

	if assets.EsbuildAvailable() {
		err := assets.BuildAssets(options)
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

func (s *Server) BuildAssets() error {
	return nil
}
