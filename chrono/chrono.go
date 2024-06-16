package chrono

import "time"

// NowFunc is a function that returns the current time (real of fake).
type NowFunc func() time.Time

// Now returns current time. Can be overridden in tests using chronotest package.
var Now NowFunc = time.Now
