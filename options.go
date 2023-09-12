package guiapi

import (
	"io/fs"

	"github.com/mbertschler/guiapi/assets"
)

func DefaultOptions() *Options {
	return &Options{
		Assets: assets.DefaultBuildOptions(),
	}
}

type Options struct {
	DistFS fs.FS
	Assets assets.BuildOptions
}
