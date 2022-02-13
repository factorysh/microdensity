package assets

import "embed"

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed musaraigne.webp
)
