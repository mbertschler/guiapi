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
