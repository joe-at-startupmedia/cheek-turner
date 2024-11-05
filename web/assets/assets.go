package assets

import "embed"

// see https://github.com/golang/go/issues/46056

//go:embed templates static
var Assets embed.FS
