package chronotest

import (
	"testing"
	"time"

	"github.com/nevkontakte/pat/chrono"
)

// OverrideNow overrides chrono.Now to return the fixed time for the duration
// of the test. Automatically restores to time.Now at the end of the test.
func OverrideNow(t *testing.T, now time.Time) {
	chrono.Now = func() time.Time { return now }
	t.Cleanup(func() { chrono.Now = time.Now })
}
