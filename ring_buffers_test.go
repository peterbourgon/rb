package rb_test

import (
	"testing"

	"github.com/peterbourgon/rb"
)

func TestRingBuffersBasics(t *testing.T) {
	t.Parallel()

	rbs := rb.NewRingBuffers[int](100)

	// Add some stuff to "foo".
	foo1 := rbs.GetOrCreate("foo")
	foo1.Add(123)
	foo1.Add(456)

	// Get it again, it's the same ring buffer.
	foo2 := rbs.GetOrCreate("foo")
	foo2.Add(789)

	// They should have the same entries.
	var have1, have2 []int
	foo1.Walk(func(i int) error { have1 = append(have1, i); return nil })
	foo2.Walk(func(i int) error { have2 = append(have2, i); return nil })
	assertEqual(t, have1, have2)

	// Create a new ring buffer "bar", and add some elements.
	bar := rbs.GetOrCreate("bar")
	bar.Add(1)
	bar.Add(2)
	bar.Add(3)
	bar.Add(4)
	bar.Add(5)
	bar.Add(6)

	// Resize all of the ring buffers down to 2 elements.
	dropped := rbs.Resize(2)
	assertEqual(t, 2, len(dropped))
	assertEqual(t, []int{123}, dropped["foo"])
	assertEqual(t, []int{4, 3, 2, 1}, dropped["bar"])
}

func TestRingBuffersClear(t *testing.T) {
	t.Parallel()

	rbs := rb.NewRingBuffers[int](5)

	foo := rbs.GetOrCreate("foo")
	foo.Add(1)
	foo.Add(2)
	foo.Add(3)

	bar := rbs.GetOrCreate("bar")
	bar.Add(10)
	bar.Add(20)

	dropped := rbs.Clear()
	assertEqual(t, 2, len(dropped))
	assertEqual(t, []int{3, 2, 1}, dropped["foo"])
	assertEqual(t, []int{20, 10}, dropped["bar"])

	// Buffers should be empty after clear.
	var have []int
	foo.Walk(func(i int) error { have = append(have, i); return nil })
	assertEqual(t, ([]int)(nil), have)
}
