package assets

type BuildOptions struct {
	Infile      string
	Outfile     string
	Bundle      bool
	Minify      bool
	Sourcemap   bool
	Log         bool
	EsbuildArgs []string
}

func DefaultBuildOptions() BuildOptions {
	return BuildOptions{
		Bundle:    true,
		Minify:    true,
		Sourcemap: true,
		Log:       true,
	}
}
