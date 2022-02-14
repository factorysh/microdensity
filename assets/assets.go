package assets

import "embed"

var (
	// used to ensure embed import
	//go:embed musaraigne.webp
	//go:embed microdensity.png
	F embed.FS
)
