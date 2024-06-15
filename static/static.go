package static

import "embed"

// StaticFS is an embedded file system with static assets.
//
//go:embed cat
var StaticFS embed.FS
