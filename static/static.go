package static

import "embed"

// StaticFS is an embedded file system with static assets.
//
//go:embed cat css favicon site.webmanifest
var StaticFS embed.FS
