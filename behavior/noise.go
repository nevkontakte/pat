package behavior

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"time"

	"golang.org/x/exp/constraints"
)

// TemporalNoise provides pseudo-random, time-dependent noise.
//
// It can be used to introduces time-dependent randomness into various
// behaviors, generate deterministic but random events, etc. All implementations
// must be thread-safe.
//
// Underlying implementations should be safe for use for any time between
// January 1, 1970 UTC and year 2262.
//
// Returned numbers are not cryptographically secure.
type TemporalNoise interface {
	// At returns a pseudo-random, deterministic value between [0, 1].
	At(t time.Time) float64
}

// Md5Noise returns hash-based, pseudo-random, time-dependent noise.
//
// SeededHash is the underlying hash algorithm, which may be pre-seeded with
// arbitrary data to generate unique sequences. Calls to At() are guaranteed
// to not mutate SeededHash state.
type Md5Noise struct {
	Seed []byte
}

func (hn Md5Noise) At(t time.Time) float64 {
	bytes, _ := t.UTC().MarshalBinary()
	h := md5.New()
	h.Write(hn.Seed)
	h.Write(bytes)
	s := h.Sum(nil)
	ui64 := binary.BigEndian.Uint64(s)
	f64 := float64(ui64)
	return f64 / math.MaxUint64
}

// SmoothNoise provides pseudo-random, smooth noise dependent on the time.
//
// It uses smoothstep interpolation to smoothen the underlying TemporalNoise
// source, which is sampled at the grid defined by Period, referenced to zero
// time.
//
// The random numbers are not cryptographically secure.
type SmoothNoise struct {
	Underlying TemporalNoise
	Period     time.Duration
}

// At returns a temporal noise value corresponding to the time point `t`. The
// value is guaranteed to be in the [0, 1.0] range.
func (tn SmoothNoise) At(t time.Time) float64 {
	t0 := t.Truncate(tn.Period)
	t1 := t0.Add(tn.Period)

	v0 := tn.Underlying.At(t0)
	v1 := tn.Underlying.At(t1)
	return v0 + smootherStep(t0, t1, t)*(v1-v0)
}

// clamp01 clamps v to the [0, 1] interval.
func clamp01(v float64) float64 {
	return min(max(v, 0), 1)
}

// Provides a smooth interpolation function between [0, 1] during the [before,
// after] period.
func smootherStep(before, after, now time.Time) float64 {
	edge0 := float64(before.UnixNano())
	edge1 := float64(after.UnixNano())
	x := float64(now.UnixNano())
	x = clamp01((x - edge0) / (edge1 - edge0))
	return x * x * x * (x*(6*x-15) + 10)
}

// Spread converts simple [0,1] float64 noise into a range of numeric values.
func Spread[T constraints.Integer | constraints.Float](min, max T, noise float64) T {
	return min + T(float64(max-min)*noise)
}
