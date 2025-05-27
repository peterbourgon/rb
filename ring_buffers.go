package rb

import "sync"

// RingBuffers collects ring buffers by string category.
type RingBuffers[T any] struct {
	mtx  sync.Mutex
	sz   int
	bufs map[string]*RingBuffer[T]
}

// NewRingBuffers returns an empty set of ring buffers, each of which will have
// a maximum size of sz, or 1, whichever is greater.
func NewRingBuffers[T any](sz int) *RingBuffers[T] {
	return &RingBuffers[T]{
		sz:   max(1, sz),
		bufs: map[string]*RingBuffer[T]{},
	}
}

// GetOrCreate returns a ring buffer for the given category string. Once a ring
// buffer is created in this way, it will always exist.
func (rbs *RingBuffers[T]) GetOrCreate(category string) *RingBuffer[T] {
	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	rb, ok := rbs.bufs[category]
	if !ok {
		rb = NewRingBuffer[T](rbs.sz)
		rbs.bufs[category] = rb
	}

	return rb
}

// GetAll returns all ring buffers by category.
func (rbs *RingBuffers[T]) GetAll() map[string]*RingBuffer[T] {
	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	all := make(map[string]*RingBuffer[T], len(rbs.bufs))
	for name, rb := range rbs.bufs {
		all[name] = rb
	}

	return all
}

// Resize all of the ring buffers in the set to the new sz, returning all
// dropped values for each ring buffer by category. If sz <= 0 it's ignored and
// the method is a no-op.
func (rbs *RingBuffers[T]) Resize(sz int) (dropped map[string][]T) {
	if sz <= 0 {
		return nil
	}

	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	rbs.sz = sz

	dropped = map[string][]T{}
	for name, rb := range rbs.bufs {
		dropped[name] = append(dropped[name], rb.Resize(sz)...)
	}

	return dropped
}
