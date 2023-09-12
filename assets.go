package guiapi

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/mbertschler/guiapi/assets"
)

func (s *Server) BuildAssets() (fs.FS, error) {
	dir := filepath.Dir(s.options.Assets.Outfile)

	if assets.EsbuildAvailable() {
		err := assets.BuildAssets(s.options.Assets)
		if err != nil {
			log.Fatal(err)
		}
		return os.DirFS(dir), nil
	}

	if s.options.Assets.Log {
		log.Println("esbuild not available, using prebuilt assets")
	}
	return fs.Sub(s.options.DistFS, dir)
}
