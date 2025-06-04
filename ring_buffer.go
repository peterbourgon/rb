package rb

import (
	"io"
	"sync"
)

// RingBuffer is a fixed-size collection of recent values.
//
// It's safe for concurrent use by multiple goroutines.
type RingBuffer[T any] struct {
	mtx sync.Mutex // explicitly not RWMutex, to avoid starving writers (Add)
	buf []T        // fully allocated at construction
	cur int        // index for next write, walk backwards to read
	len int        // count of actual values
}

// NewRingBuffer returns an empty ring buffer of values of type T, with a
// pre-allocated and fixed size as defined by sz.
func NewRingBuffer[T any](sz int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buf: make([]T, sz),
	}
}

// Resize the ring buffer to the given size. If the new size is smaller than the
// existing size, resize will drop the oldest values as necessary, and return
// those dropped values. If sz <= 0 it's ignored and the method is a no-op.
func (rb *RingBuffer[T]) Resize(sz int) (dropped []T) {
	// Safety first.
	if sz <= 0 {
		return nil
	}

	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	// Calculate how many values to fill from the old buffer to the new one.
	fill := rb.len
	if fill > sz {
		fill = sz
	}

	// Calculate the read cursor for the old buffer.
	rdcur := rb.cur - 1
	if rdcur < 0 {
		rdcur += rb.len
	}

	// Construct the new buffer with the given size. As fill is guaranteed to be
	// less than or equal to sz, we calculate the write cursor as simply fill,
	// and will copy values by walking both cursors backwards.
	buf := make([]T, sz)
	wrcur := fill - 1

	// Copy recent values from the old buffer to the new buffer.
	for wrcur >= 0 {
		buf[wrcur] = rb.buf[rdcur]

		rdcur -= 1
		if rdcur < 0 {
			rdcur += len(rb.buf)
		}

		wrcur -= 1 // no need to do the wraparound math
	}

	// If we resized smaller, and the old buffer has more values than the new
	// size, then capture the values from the old buffer which are dropped.
	for i := sz; i < rb.len; i++ {
		dropped = append(dropped, rb.buf[rdcur])

		rdcur -= 1
		if rdcur < 0 {
			rdcur += len(rb.buf)
		}
	}

	// Calculate the next write cursor for the new buffer. If we resized
	// smaller, then fill will equal sz, and we need to wrap around.
	cur := fill
	if cur >= sz {
		cur -= sz
	}

	// Modify all of the buffer fields to their new values.
	rb.buf = buf
	rb.cur = cur
	rb.len = fill

	// Done.
	return dropped
}

// Add the value to the ring buffer. If the ring buffer was full, and the oldest
// value was overwritten by this add, return that oldest/dropped value and true;
// otherwise, return a zero value and false.
func (rb *RingBuffer[T]) Add(val T) (dropped T, ok bool) {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	// Safety first.
	if cap(rb.buf) <= 0 {
		var zero T
		return zero, false
	}

	// Capture any overwritten value so it can be returned.
	if rb.len >= len(rb.buf) {
		dropped, ok = rb.buf[rb.cur], true
	}

	// Write the value at the write cursor.
	rb.buf[rb.cur] = val

	// Update the ring buffer size.
	if rb.len < len(rb.buf) {
		rb.len += 1
	}

	// Advance the write cursor.
	rb.cur += 1
	if rb.cur >= len(rb.buf) {
		rb.cur -= len(rb.buf)
	}

	// Done.
	return dropped, ok
}

// Walk calls the given function for each value in the ring buffer, starting
// with the most recent value, and ending with the oldest value. Walk takes an
// exclusive lock on the ring buffer, which blocks other calls, including Add.
func (rb *RingBuffer[T]) Walk(fn func(T) error) error {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	// Read up to rb.len values.
	for i := range rb.len {
		// Reads go backwards from one before the write cursor.
		cur := rb.cur - 1 - i

		// Wrap around when necessary.
		if cur < 0 {
			cur += len(rb.buf)
		}

		// If the function returns an error, we're done.
		if err := fn(rb.buf[cur]); err != nil {
			return err
		}
	}

	return nil
}

// Overview returns the newest and oldest values in the ring buffer, as well as
// the total number of values stored in the ring buffer.
func (rb *RingBuffer[T]) Overview() (newest, oldest T, count int) {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	// The cursor math assumes a non-empty buffer.
	if rb.len == 0 {
		var zero T
		return zero, zero, 0
	}

	// The read head is the value just before the write cursor.
	headidx := rb.cur - 1
	if headidx < 0 {
		headidx += len(rb.buf)
	}

	// The read tail is len-1 values back from the read head.
	// If the buffer is full, this is the write cursor.
	tailidx := headidx - (rb.len - 1)
	if tailidx < 0 {
		tailidx += len(rb.buf)
	}

	return rb.buf[headidx], rb.buf[tailidx], rb.len
}

// Copy the most recent values from the ring buffer into dst, newest first.
// Returns the number of values copied into dst.
func (rb *RingBuffer[T]) Copy(dst []T) (int, error) {
	var index int
	rb.Walk(func(val T) error {
		if index >= len(dst) {
			return io.EOF
		}
		dst[index] = val
		index += 1
		return nil
	})
	return index, nil
}

// Take copies up to the n most recent values from the ring buffer into a newly
// allocated slice, newest-to-oldest, and returns that slice. The ring buffer
// isn't modified.
func (rb *RingBuffer[T]) Take(n int) ([]T, error) {
	dst := make([]T, n)
	n, err := rb.Copy(dst)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}
