//go:build !no_esbuild

package assets

import (
	"fmt"

	"github.com/evanw/esbuild/pkg/cli"
)

func EsbuildAvailable() bool {
	return true
}

func BuildAssets(options BuildOptions) error {
	flags := []string{
		options.Infile,
		"--outfile=" + options.Outfile,
	}
	if options.Bundle {
		flags = append(flags, "--bundle")
	}
	if options.Minify {
		flags = append(flags, "--minify")
	}
	if options.Sourcemap {
		flags = append(flags, "--sourcemap")
	}
	if options.Log {
		flags = append(flags, "--log-level=info")
	} else {
		flags = append(flags, "--log-level=warning")
	}
	flags = append(flags, options.EsbuildArgs...)
	returnCode := cli.Run(flags)
	if returnCode != 0 {
		return fmt.Errorf("esbuild failed with code %d", returnCode)
	}
	return nil
}
