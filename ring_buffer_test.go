package rb_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/peterbourgon/rb"
)

func assertEqual[T any](t *testing.T, have, want T) {
	t.Helper()
	if !cmp.Equal(have, want) {
		t.Fatal(cmp.Diff(have, want))
	}
}

func TestRingBuffer(t *testing.T) {
	t.Parallel()

	rb := rb.NewRingBuffer[int](3)

	top := func(k int) []int {
		res := []int{}
		rb.Walk(func(i int) error {
			if k >= 0 && len(res) >= k {
				return errors.New("done")
			}
			res = append(res, int(i))
			return nil
		})
		return res
	}

	assertEqual(t, top(-1), []int{})
	assertEqual(t, top(0), []int{})
	assertEqual(t, top(99), []int{})

	rb.Add(1)

	assertEqual(t, top(-1), []int{1})
	assertEqual(t, top(0), []int{})
	assertEqual(t, top(1), []int{1})
	assertEqual(t, top(2), []int{1})
	assertEqual(t, top(3), []int{1})
	assertEqual(t, top(4), []int{1})

	rb.Add(2)

	assertEqual(t, top(-1), []int{2, 1})
	assertEqual(t, top(0), []int{})
	assertEqual(t, top(1), []int{2})
	assertEqual(t, top(2), []int{2, 1})
	assertEqual(t, top(3), []int{2, 1})
	assertEqual(t, top(4), []int{2, 1})

	rb.Add(3)

	assertEqual(t, top(-1), []int{3, 2, 1})
	assertEqual(t, top(0), []int{})
	assertEqual(t, top(1), []int{3})
	assertEqual(t, top(2), []int{3, 2})
	assertEqual(t, top(3), []int{3, 2, 1})
	assertEqual(t, top(4), []int{3, 2, 1})

	removed, did := rb.Add(4)

	assertEqual(t, did, true)
	assertEqual(t, removed, 1)
	assertEqual(t, top(-1), []int{4, 3, 2})
	assertEqual(t, top(0), []int{})
	assertEqual(t, top(1), []int{4})
	assertEqual(t, top(2), []int{4, 3})
	assertEqual(t, top(3), []int{4, 3, 2})
	assertEqual(t, top(4), []int{4, 3, 2})

	rb.Add(5)
	rb.Add(6)

	assertEqual(t, top(-1), []int{6, 5, 4})
	assertEqual(t, top(99), []int{6, 5, 4})
}

func TestRingBufferCopyTake(t *testing.T) {
	rb := rb.NewRingBuffer[int](32)
	rb.Add(1)
	rb.Add(2)
	rb.Add(3)
	rb.Add(4)
	rb.Add(5)

	{
		var dst []int
		n, err := rb.Copy(dst)
		assertEqual(t, error(nil), err)
		assertEqual(t, 0, n)
		assertEqual(t, ([]int)(nil), dst)
	}

	{
		dst := make([]int, 0)
		n, err := rb.Copy(dst)
		assertEqual(t, error(nil), err)
		assertEqual(t, 0, n)
		assertEqual(t, []int{}, dst)
	}

	{
		dst := make([]int, 1)
		n, err := rb.Copy(dst)
		assertEqual(t, error(nil), err)
		assertEqual(t, 1, n)
		assertEqual(t, []int{5}, dst)
	}

	{
		dst := make([]int, 3)
		n, err := rb.Copy(dst)
		assertEqual(t, error(nil), err)
		assertEqual(t, 3, n)
		assertEqual(t, []int{5, 4, 3}, dst)
	}

	{
		dst := make([]int, 10)
		n, err := rb.Copy(dst)
		assertEqual(t, error(nil), err)
		assertEqual(t, 5, n)
		assertEqual(t, []int{5, 4, 3, 2, 1, 0, 0, 0, 0, 0}, dst)
		assertEqual(t, []int{5, 4, 3, 2, 1}, dst[:n])
	}

	{
		vals, err := rb.Take(0)
		assertEqual(t, error(nil), err)
		assertEqual(t, []int{}, vals)
	}

	{
		vals, err := rb.Take(1)
		assertEqual(t, error(nil), err)
		assertEqual(t, []int{5}, vals)
	}

	{
		vals, err := rb.Take(5)
		assertEqual(t, error(nil), err)
		assertEqual(t, []int{5, 4, 3, 2, 1}, vals)
	}

	{
		vals, err := rb.Take(6)
		assertEqual(t, error(nil), err)
		assertEqual(t, []int{5, 4, 3, 2, 1}, vals)
	}
}

func TestRingBufferOverview(t *testing.T) {
	t.Parallel()

	firstLast := func(rb *rb.RingBuffer[int]) (int, int) {
		var count, first, last int
		rb.Walk(func(i int) error {
			if count == 0 {
				first = i
			}
			last = i
			count++
			return nil
		})
		return first, last
	}

	{
		rb := rb.NewRingBuffer[int](0)
		var zeroint int

		newest, oldest, n := rb.Overview()
		assertEqual(t, newest, zeroint)
		assertEqual(t, oldest, zeroint)
		assertEqual(t, n, 0)

		rb.Add(1)
		rb.Add(2)

		newest, oldest, n = rb.Overview()
		first, last := firstLast(rb)
		assertEqual(t, newest, first)
		assertEqual(t, oldest, last)
		assertEqual(t, n, 0)
	}

	{
		rb := rb.NewRingBuffer[int](10)

		rb.Add(1)
		rb.Add(2)
		rb.Add(3)

		newest, oldest, n := rb.Overview()
		assertEqual(t, newest, 3)
		assertEqual(t, oldest, 1)
		assertEqual(t, n, 3)

		first, last := firstLast(rb)
		assertEqual(t, newest, first)
		assertEqual(t, oldest, last)
	}

	{
		rb := rb.NewRingBuffer[int](123)

		for i := 42; i < 951; i++ {
			rb.Add(int(i))
		}

		newest, oldest, n := rb.Overview()
		first, last := firstLast(rb)
		assertEqual(t, newest, first)
		assertEqual(t, oldest, last)
		assertEqual(t, n, 123)
	}
}

func TestRingBufferResize(t *testing.T) {
	t.Parallel()

	rb := rb.NewRingBuffer[int](3)

	top := func(k int) []int {
		res := []int{}
		rb.Walk(func(i int) error {
			if k >= 0 && len(res) >= k {
				return errors.New("done")
			}
			res = append(res, int(i))
			return nil
		})
		return res
	}

	rb.Add(1)
	rb.Add(2)
	rb.Add(3)

	assertEqual(t, top(3), []int{3, 2, 1})

	removed := rb.Resize(2)

	assertEqual(t, removed, []int{1})
	assertEqual(t, top(3), []int{3, 2})

	removed = rb.Resize(4)

	assertEqual(t, removed, nil)
	assertEqual(t, top(3), []int{3, 2})

	rb.Add(4)
	rb.Add(5)
	rb.Add(6)
	rb.Add(7)

	assertEqual(t, top(3), []int{7, 6, 5})
	assertEqual(t, top(10), []int{7, 6, 5, 4})
}

func BenchmarkRingBuffer(b *testing.B) {
	for _, cap := range []int{100, 1_000, 10_000, 100_000, 1_000_000} {
		b.Run(fmt.Sprintf("cap=%d", cap), func(b *testing.B) {
			rb := rb.NewRingBuffer[int](cap)
			for i := range cap {
				rb.Add(int(i))
			}

			var captured int
			_ = captured

			walkOnlyFn := func(int) error {
				return nil
			}

			walkReadFn := func(i int) error {
				captured = i
				return nil
			}

			b.ReportAllocs()

			b.Run("Add", func(b *testing.B) {
				for i := range b.N {
					rb.Add(int(i))
				}
			})

			b.Run("Walk", func(b *testing.B) {
				for b.Loop() {
					rb.Walk(walkOnlyFn)
				}
			})

			b.Run("Walk+Read", func(b *testing.B) {
				for b.Loop() {
					rb.Walk(walkReadFn)
				}
			})

			b.Run("Add+Walk", func(b *testing.B) {
				for i := range b.N {
					rb.Add(int(i))
					rb.Walk(walkOnlyFn)
				}
			})

			b.Run("Add+Walk+Read", func(b *testing.B) {
				for i := range b.N {
					rb.Add(int(i))
					rb.Walk(walkReadFn)
				}
			})
		})
	}
}

func BenchmarkRingBufferParallel(b *testing.B) {
	walkFn := func(int) error { return nil }
	_ = walkFn

	for _, cap := range []int{100, 1000, 10000} {
		for _, par := range []int{10, 100, 1000} {
			b.Run(fmt.Sprintf("cap=%d/par=%d", cap, par), func(b *testing.B) {
				rb := rb.NewRingBuffer[int](cap)
				b.SetParallelism(par)
				b.RunParallel(func(p *testing.PB) {
					for p.Next() {
						rb.Add(123)
						rb.Walk(walkFn)
					}
				})
			})
		}
	}
}

func BenchmarkCopyTake(b *testing.B) {
	for _, tc := range []struct {
		cap   int
		sizes []int
	}{
		{
			cap:   1_000,
			sizes: []int{1, 10, 100, 1000, 10000},
		},
		{
			cap:   100_000,
			sizes: []int{1, 10, 100, 1000, 10000},
		},
	} {
		b.Run(fmt.Sprintf("cap=%d", tc.cap), func(b *testing.B) {
			rb := rb.NewRingBuffer[int](tc.cap)
			for i := range tc.cap {
				rb.Add(int(i))
			}

			for _, sz := range tc.sizes {
				b.Run(fmt.Sprintf("sz=%d", sz), func(b *testing.B) {

					b.Run("Copy", func(b *testing.B) {
						dst := make([]int, sz)
						b.ResetTimer()
						b.ReportAllocs()
						for b.Loop() {
							rb.Copy(dst)
						}
					})

					b.Run("Take", func(b *testing.B) {
						b.ResetTimer()
						b.ReportAllocs()
						for b.Loop() {
							rb.Take(sz)
						}
					})

				})
			}
		})
	}
}
