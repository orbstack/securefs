package syncx

import "testing"

func TestOnce(t *testing.T) {
	t.Parallel()

	var once Once[int]
	var count int

	once.Do(func() int {
		count++
		return 1
	})

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	once.Do(func() int {
		count++
		return 2
	})

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestOnceParallel(t *testing.T) {
	t.Parallel()

	var once Once[int]
	var count int

	f := func() int {
		once.Do(func() int {
			count++
			return 1
		})
		return once.Value
	}

	if got, want := f(), 1; got != want {
		t.Errorf("f() = %d, want %d", got, want)
	}

	if got, want := f(), 1; got != want {
		t.Errorf("f() = %d, want %d", got, want)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
