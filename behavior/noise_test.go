package behavior

import (
	"math"
	"testing"
	"time"
)

// checkDeterminism verifies that repeated calls to noise.At with the same
// timestamp return the same value.
func checkDeterminism(t *testing.T, noise TemporalNoise, ts time.Time) {
	val1 := noise.At(ts)
	val2 := noise.At(ts)
	if val1 != val2 {
		t.Errorf("Property 1 (determinism) failed for ts %v: expected %v, got %v",
			ts, val1, val2)
	}
}

// checkRange verifies that the noise value returned by noise.At(ts) is within
// the [0, 1] range.
func checkRange(t *testing.T, noise TemporalNoise, ts time.Time) {
	val := noise.At(ts)
	if val < 0.0 || val > 1.0 {
		t.Errorf("Property 3 (range) failed for ts %v: noise.At() returned value %v which is out of [0, 1] range",
			ts, val)
	}
}

// checkUniqueness verifies that noise values for different timestamps are
// different.
//
// Technically, they *could* collide, but the probability is extremely low
// (~1/2^53), so we pretty much ignore it.
func checkUniqueness(t *testing.T, noise TemporalNoise, ts1, ts2 time.Time) {
	val1 := noise.At(ts1)
	val2 := noise.At(ts2)

	if !ts1.Equal(ts2) {
		// Timestamps are different. Expect noise values to be different.
		if val1 == val2 {
			t.Errorf("Property 2 (uniqueness for different times) failed: ts1 %v and ts2 %v are different, but both produced noise value %v",
				ts1, ts2, val1)
		}
	}
}

func FuzzMd5NoiseProperties(f *testing.F) {
	// Seed corpus with diverse inputs:
	// t1_nanos (int64), seed_bytes ([]byte), t2_nanos (int64)
	f.Add(int64(0), []byte("seed1"), int64(1))                              // Epoch, simple seed, t2 just after t1
	f.Add(int64(1000000000), []byte("seed2"), int64(1000000000))            // Same time
	f.Add(int64(1678886400000000000), []byte{}, int64(1678886400000000001)) // Empty seed, t2 just after t1
	f.Add(int64(-1000000000), []byte{0x01, 0x02}, int64(2000000000))        // Pre-epoch and post-epoch
	f.Add(int64(0), []byte(nil), int64(1))                                  // Nil seed
	f.Add(int64(math.MaxInt64), []byte("maxTime"), int64(0))                // Max int64 nanos (Year 2262)
	f.Add(int64(0), []byte("minTime"), int64(math.MinInt64))                // Min int64 nanos (Year 1677)

	f.Fuzz(func(t *testing.T, timeNano1 int64, seed []byte, timeNano2 int64) {
		t.Logf("Fuzzing Md5Noise with parameters: timeNano1=%d, seed=%x, timeNano2=%d", timeNano1, seed, timeNano2)

		noise := Md5Noise{Seed: seed}

		// Create time.Time objects from nanoseconds since epoch.
		// Md5Noise.At uses t.UTC(), so we do the same for consistency.
		ts1 := time.Unix(0, timeNano1).UTC()
		ts2 := time.Unix(0, timeNano2).UTC()

		// Check determinism for both timestamps.
		checkDeterminism(t, noise, ts1)
		checkDeterminism(t, noise, ts2)

		// Check that different timestamps generate different values.
		checkUniqueness(t, noise, ts1, ts2)

		// Check range for both timestamps.
		checkRange(t, noise, ts1)
		checkRange(t, noise, ts2)
	})
}

func FuzzSmoothNoiseProperties(f *testing.F) {
	const (
		minNanosPeriod = int64(1) // Smallest valid period (1ns) for SmoothNoise
		defaultPeriod  = int64(time.Second)
		smallPeriod    = int64(time.Millisecond * 100)
		largePeriod    = int64(time.Hour)
	)

	// Seed corpus: t1Nanos, seedBytes, t2Nanos, periodNanos
	f.Add(int64(0), []byte("s_seed1"), int64(1), defaultPeriod)
	f.Add(int64(1e9), []byte("s_seed2"), int64(1e9), smallPeriod) // Same time
	f.Add(int64(1678886400000000000), []byte{}, int64(1678886400000000001), largePeriod)
	f.Add(int64(-1e9), []byte{0x01, 0x02, 0x03}, int64(2e9), defaultPeriod)
	f.Add(int64(0), []byte(nil), int64(1), smallPeriod) // Nil seed
	f.Add(int64(math.MaxInt64), []byte("s_maxTime"), int64(0), largePeriod)
	f.Add(int64(0), []byte("s_minTime"), int64(math.MinInt64), defaultPeriod)
	f.Add(int64(0), []byte("s_tinyPeriod"), int64(100), minNanosPeriod)           // Test with minimal period
	f.Add(int64(0), []byte("s_zeroOffset"), smallPeriod/2, smallPeriod)           // t1 near start of period, t2 in middle
	f.Add(int64(0), []byte("s_crossPeriod"), smallPeriod*2, smallPeriod)          // t1 at start, t2 in next period
	f.Add(int64(0), []byte("s_largePeriodVal"), int64(100), int64(math.MaxInt64)) // Max int64 for period

	f.Fuzz(func(t *testing.T, timeNano1 int64, seed []byte, timeNano2 int64, periodNanos int64) {
		t.Logf("Fuzzing SmoothNoise with parameters: timeNano1=%d, seed=%x, timeNano2=%d, periodNanosInput=%d",
			timeNano1, seed, timeNano2, periodNanos)

		// Ensure periodNanos is valid for SmoothNoise.
		// A period less than minNanosPeriod (1ns) might be problematic or invalid.
		if periodNanos < minNanosPeriod {
			periodNanos = minNanosPeriod
		}
		smoothPeriodDuration := time.Duration(periodNanos)
		underlyingNoise := Md5Noise{Seed: seed}
		smoothNoise := SmoothNoise{Underlying: underlyingNoise, Period: smoothPeriodDuration}

		ts1 := time.Unix(0, timeNano1).UTC()
		ts2 := time.Unix(0, timeNano2).UTC()

		// Property 1: Determinism (seed here is for the underlying Md5Noise)
		checkDeterminism(t, smoothNoise, ts1)
		checkDeterminism(t, smoothNoise, ts2)

		// Property 3: Range [0,1] (assuming underlying Md5Noise is [0,1])
		checkRange(t, smoothNoise, ts1)
		checkRange(t, smoothNoise, ts2)

		// Property 2: Uniqueness (custom logic for SmoothNoise)
		if ts1.Sub(ts2).Abs() > time.Millisecond {
			// If values are too close, they would be interpolated into the same value.
			checkUniqueness(t, smoothNoise, ts1, ts2)
		}
	})
}

func BenchmarkMd5Noise(b *testing.B) {
	noise := Md5Noise{Seed: []byte("benchmark_seed")}
	ts := time.Unix(1678886400, 0).UTC() // A fixed timestamp

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = noise.At(ts)
	}
}

func BenchmarkSmoothNoise(b *testing.B) {
	smoothNoise := SmoothNoise{
		Underlying: Md5Noise{Seed: []byte("benchmark_smooth_seed")},
		Period:     time.Second,
	}
	ts := time.Unix(1678886400, 0).UTC() // A fixed timestamp

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = smoothNoise.At(ts)
	}
}

func TestMd5NoiseConsecutiveDifference(t *testing.T) {
	noise := Md5Noise{Seed: []byte("md5_consecutive_diff_seed")}
	startTs := time.Unix(1234567890, 0).UTC()
	numSamples := 1000
	increment := time.Nanosecond // Smallest possible increment
	expectedMinAvgDiff := 0.3

	var totalAbsDiff float64
	prevVal := noise.At(startTs)

	for i := range numSamples {
		currentTs := startTs.Add(time.Duration(i+1) * increment)
		currentVal := noise.At(currentTs)
		totalAbsDiff += math.Abs(currentVal - prevVal)
		prevVal = currentVal
	}

	avgDiff := totalAbsDiff / float64(numSamples)
	t.Logf("Md5Noise average consecutive difference: %f", avgDiff)
	if avgDiff <= expectedMinAvgDiff {
		t.Errorf("Md5Noise average consecutive difference was %f, expected > %f", avgDiff, expectedMinAvgDiff)
	}
}

func TestSmoothNoiseConsecutiveDifference(t *testing.T) {
	smoothNoise := SmoothNoise{
		Underlying: Md5Noise{Seed: []byte("smooth_consecutive_diff_seed")},
		Period:     time.Second, // A typical period for smoothing
	}
	startTs := time.Unix(987654321, 0).UTC()
	numSamples := 1000
	increment := time.Nanosecond // Very small increment to test smoothness
	expectedMaxAvgDiff := 0.1

	var totalAbsDiff float64
	prevVal := smoothNoise.At(startTs)

	for i := range numSamples {
		currentTs := startTs.Add(time.Duration(i+1) * increment)
		currentVal := smoothNoise.At(currentTs)
		totalAbsDiff += math.Abs(currentVal - prevVal)
		prevVal = currentVal
	}

	avgDiff := totalAbsDiff / float64(numSamples)
	if avgDiff >= expectedMaxAvgDiff {
		t.Errorf("SmoothNoise average consecutive difference was %f, expected < %f (period: %v)", avgDiff, expectedMaxAvgDiff, smoothNoise.Period)
	}
}
