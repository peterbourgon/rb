package rb

import "sync"

// RingBuffers collects individual ring buffers by string key.
type RingBuffers[T any] struct {
	mtx  sync.Mutex
	cap  int
	bufs map[string]*RingBuffer[T]
}

// NewRingBuffers returns an empty set of ring buffers, each of which will have
// a maximum capacity of the given cap.
func NewRingBuffers[T any](cap int) *RingBuffers[T] {
	return &RingBuffers[T]{
		cap:  cap,
		bufs: map[string]*RingBuffer[T]{},
	}
}

// GetOrCreate returns a ring buffer corresponding to the given category string.
// Once a ring buffer is created in this way, it will always exist.
func (rbs *RingBuffers[T]) GetOrCreate(category string) *RingBuffer[T] {
	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	rb, ok := rbs.bufs[category]
	if !ok {
		rb = NewRingBuffer[T](rbs.cap)
		rbs.bufs[category] = rb
	}

	return rb
}

// GetAll returns all of the ring buffers in the set, grouped by category.
func (rbs *RingBuffers[T]) GetAll() map[string]*RingBuffer[T] {
	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	all := make(map[string]*RingBuffer[T], len(rbs.bufs))
	for name, rb := range rbs.bufs {
		all[name] = rb
	}

	return all
}

// Resize all of the ring buffers in the set to the new capacity.
func (rbs *RingBuffers[T]) Resize(cap int) (dropped map[string][]T) {
	if cap <= 0 {
		return
	}

	rbs.mtx.Lock()
	defer rbs.mtx.Unlock()

	rbs.cap = cap

	dropped = map[string][]T{}
	for name, rb := range rbs.bufs {
		dropped[name] = append(dropped[name], rb.Resize(cap)...)
	}

	return dropped
}
