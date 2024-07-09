package static

import "embed"

//go:embed *
var staticFiles embed.FS

func FS() embed.FS {
	return staticFiles
}
