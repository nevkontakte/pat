package static

import "embed"

// StaticFS is an embedded file system with static assets.
//
//go:embed cat css
var StaticFS embed.FS
