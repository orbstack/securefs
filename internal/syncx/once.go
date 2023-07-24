package syncx

import "sync"

type Once[T any] struct {
	sync.Once
	Value T
}

func (o *Once[T]) Do(f func() T) T {
	o.Once.Do(func() {
		o.Value = f()
	})
	return o.Value
}
