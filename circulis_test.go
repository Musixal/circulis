package circulis

import (
	"sync"
	"testing"
)

// TestConcurrentReadWrite spins up multiple writers and readers to
// exercise the ring buffer under contention and then runs with -race.
func TestConcurrentReadWrite(t *testing.T) {
	const (
		writers    = 8
		readers    = 8
		iterations = 10000
		bufSize    = 128
	)
	c := New(1024)

	// prepare a pattern
	pattern := make([]byte, bufSize)
	for i := range pattern {
		pattern[i] = byte(i)
	}

	var writeWG sync.WaitGroup
	var readWG sync.WaitGroup
	writeWG.Add(writers)
	readWG.Add(readers)

	// writer goroutines
	for w := 0; w < writers; w++ {
		go func() {
			defer writeWG.Done()
			for i := 0; i < iterations; i++ {
				if n, err := c.Write(pattern); err != nil {
					t.Errorf("Write error: %v (wrote %d bytes)", err, n)
				}
			}
		}()
	}

	// reader goroutines
	for r := 0; r < readers; r++ {
		go func() {
			defer readWG.Done()
			buf := make([]byte, bufSize)
			for {
				n, err := c.Read(buf)
				if err == ErrClosed {
					return
				}
				if err != nil {
					t.Errorf("Read error: %v", err)
					return
				}
				if n != len(pattern) {
					t.Errorf("Read size %d, want %d", n, len(pattern))
				}
				// verify data integrity
				for i := 0; i < n; i++ {
					if buf[i] != byte(i) {
						t.Errorf("Data mismatch at %d: got %d", i, buf[i])
						break
					}
				}
			}
		}()
	}

	// wait for all writers then close the buffer
	writeWG.Wait()
	c.Close()
	// wait for readers to drain
	readWG.Wait()
}
