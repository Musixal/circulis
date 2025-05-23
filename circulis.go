package circulis

import (
	"errors"
	"sync"
)

// Predefined errors
var (
	ErrClosed = errors.New("circulis: buffer closed")
	ErrFull   = errors.New("circulis: buffer full")
	ErrEmpty  = errors.New("circulis: buffer empty")
)

// Circulis is a fixed-size, power-of-two-capacity ring buffer.
// It is safe for concurrent use by multiple readers/writers.
type Circulis struct {
	buf      []byte   // underlying storage
	mask     uint64   // = cap(buf)-1, for wrapping
	head     uint64   // next read index (monotonic)
	_        [56]byte // pad out the cache line
	tail     uint64   // next write index (monotonic)
	_        [56]byte
	closed   bool // set when Close() is called
	blocking bool // if true: Read/Write block on empty/full

	mu       sync.Mutex
	notEmpty *sync.Cond // signaled when data is written
	notFull  *sync.Cond // signaled when data is read
}

// New creates a *Circulis with at least the requested capacity,
// rounded up to the next power of two. By default it is in blocking mode.
func New(capacity int) *Circulis {
	if capacity < 1 {
		panic("circulis: capacity must be > 0")
	}
	cap2 := nextPowerOfTwo(uint64(capacity))
	c := &Circulis{
		buf:      make([]byte, cap2),
		mask:     cap2 - 1,
		blocking: false,
	}
	c.notEmpty = sync.NewCond(&c.mu)
	c.notFull = sync.NewCond(&c.mu)
	return c
}

// SetBlocking enables or disables blocking behavior for both Read and Write.
// When blocking=false, Write returns ErrFull immediately if no space,
// Read returns ErrEmpty immediately if no data.
func (c *Circulis) SetBlocking(blocking bool) {
	c.mu.Lock()
	c.blocking = blocking
	// wake all waiters so they re-check their conditions
	c.notEmpty.Broadcast()
	c.notFull.Broadcast()
	c.mu.Unlock()
}

// Write writes up to len(p) bytes into the buffer.
// It returns the number of bytes written and, if applicable, ErrFull or ErrClosed.
//
// If blocking=true, it will block until it can write all bytes or the buffer is closed.
// If blocking=false, it writes as much as fits (possibly zero) and returns ErrFull
// if not all bytes were written.
func (c *Circulis) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If closed, no more writes allowed.
	if c.closed {
		return 0, ErrClosed
	}

	total := len(p)
	for n < total {
		free := int(uint64(len(c.buf)) - (c.tail - c.head))
		if free == 0 {
			// buffer full
			if !c.blocking {
				if n == 0 {
					return 0, ErrFull
				}
				return n, ErrFull
			}
			// wait for readers to consume
			c.notFull.Wait()
			if c.closed {
				return n, ErrClosed
			}
			continue
		}

		// how many we can write now
		toWrite := total - n
		if toWrite > free {
			toWrite = free
		}
		// write in up to two segments
		start := c.tail & c.mask
		first := toWrite
		// if segment wraps past end of buffer
		endSpace := int(uint64(len(c.buf)) - start)
		if first > endSpace {
			first = endSpace
		}
		// copy first chunk
		copy(c.buf[start:start+uint64(first)], p[n:n+first])
		// copy second chunk if needed
		second := toWrite - first
		if second > 0 {
			copy(c.buf[0:second], p[n+first:n+first+second])
		}

		c.tail += uint64(toWrite)
		n += toWrite
		// wake one reader
		c.notEmpty.Signal()
	}
	return n, nil
}

// Read reads up to len(p) bytes from the buffer.
// It returns the number of bytes read and, if applicable, ErrEmpty or ErrClosed.
//
// If blocking=true, it will block until at least 1 byte is available or the buffer is closed.
// If blocking=false, it returns ErrEmpty immediately if no data.
func (c *Circulis) Read(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		available := int(c.tail - c.head)
		if available == 0 {
			// buffer empty
			if c.closed {
				return 0, ErrClosed
			}
			if !c.blocking {
				return 0, ErrEmpty
			}
			c.notEmpty.Wait()
			continue
		}
		// we have at least one byte
		toRead := len(p)
		if toRead > available {
			toRead = available
		}
		// read in up to two segments
		start := c.head & c.mask
		first := toRead
		endSpace := int(uint64(len(c.buf)) - start)
		if first > endSpace {
			first = endSpace
		}
		copy(p[0:first], c.buf[start:start+uint64(first)])
		second := toRead - first
		if second > 0 {
			copy(p[first:first+second], c.buf[0:second])
		}

		c.head += uint64(toRead)
		n = toRead
		// wake one writer
		c.notFull.Signal()
		return n, nil
	}
}

// Close marks the buffer as closed. Further Write calls return ErrClosed.
// Any goroutines blocked in Read or Write are awakened and will see ErrClosed once drained.
func (c *Circulis) Close() {
	c.mu.Lock()
	c.closed = true
	c.notEmpty.Broadcast()
	c.notFull.Broadcast()
	c.mu.Unlock()
}

// nextPowerOfTwo returns the smallest power of two >= v.
func nextPowerOfTwo(v uint64) uint64 {
	if v == 0 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	return v + 1
}
