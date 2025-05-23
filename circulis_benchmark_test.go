// Benchmark tests adapted from https://github.com/smallnest/ringbuffer
// Original author: smallnest
// Used under MIT license for comparison and validation purposes.

package circulis

import (
	"strings"
	"testing"
	"time"
)

func BenchmarkRingBuffer_Sync(b *testing.B) {
	rb := New(1024)
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(data)
		rb.Read(buf)
	}
}

func BenchmarkRingBuffer_AsyncRead(b *testing.B) {
	rb := New(1024)
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	go func() {
		for {
			_, err := rb.Read(buf)
			if err == ErrEmpty {
				time.Sleep(1 * time.Nanosecond)
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(data)
	}
}

func BenchmarkRingBuffer_AsyncReadBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := New(sz * buffers)
	rb.SetBlocking(true)

	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			rb.Read(buf)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(data)
	}
}

func BenchmarkRingBuffer_AsyncWrite(b *testing.B) {
	rb := New(1024)
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	go func() {
		for {
			rb.Write(data)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Read(buf)
	}
}

func BenchmarkRingBuffer_AsyncWriteBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10

	rb := New(sz * buffers)
	rb.SetBlocking(true)

	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			rb.Write(data)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Read(buf)
	}
}
