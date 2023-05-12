//go:build !production

package guiapi

import (
	"fmt"
	"log"

	"github.com/evanw/esbuild/pkg/cli"
)

func EsbuildAvailable() bool {
	return true
}

func BuildAssets() error {
	log.Println("building browser assets")
	options := []string{
		"js/main.js",
		"--bundle",
		"--outfile=dist/bundle.js",
		"--minify",
		"--sourcemap",
	}
	returnCode := cli.Run(options)
	if returnCode != 0 {
		return fmt.Errorf("esbuild failed with code %d", returnCode)
	}
	return nil
}
