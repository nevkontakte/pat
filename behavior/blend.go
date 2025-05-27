package behavior

import "time"

type Period[T any] struct {
	Duration time.Duration
	Value    T
}

type Schedule[T any] []Period[T]

type Blender[T any] struct {
	Schedule Schedule[T]
	Seed     uint64
}

func (b Blender[T]) Fixed(fraction float64) Schedule[T] {
	return nil
}
