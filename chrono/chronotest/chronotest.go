package chronotest

import (
	"testing"
	"time"

	"github.com/nevkontakte/pat/chrono"
)

// OverrideNow overrides chrono.Now to return the fixed time for the duration
// of the test. Automatically restores to time.Now at the end of the test.
func OverrideNow(t *testing.T, now time.Time) {
	orig := chrono.Now
	chrono.Now = func() time.Time { return now }
	t.Cleanup(func() { chrono.Now = orig })
}

func OverrideScope(now time.Time, f func()) {
	orig := chrono.Now
	chrono.Now = func() time.Time { return now }
	defer func() { chrono.Now = orig }()
	f()
}
