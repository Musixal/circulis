# circulis

**circulis** is a thread-safe, high-performance circular buffer (ring buffer) for Go.  
It supports multiple concurrent readers and writers, with optional blocking or non-blocking behavior.

---

## Features

- ✅ Fixed-size ring buffer with power-of-two capacity
- ✅ Thread-safe for concurrent `Read` and `Write` calls
- ✅ Supports both blocking and non-blocking modes
- ✅ Efficient, minimal allocation design
- ✅ Graceful shutdown with `Close()`

---

## Installation

```bash
go get github.com/Musixal/circulis
```

## Usage
package main

```go
import (
    "fmt"
    "log"
    "time"

    "github.com/Musixal/circulis"
)

func main() {
    buf := circulis.New(1024)


    go func() {
        data := make([]byte, 512)
        for {
            n, err := buf.Read(data)
            if err == circulis.ErrClosed {
                fmt.Println("Buffer closed")
                return
            } else if err == circulis.ErrEmpty {
                time.Sleep(time.Millisecond)
                continue
            }
            fmt.Println("Read:", string(data[:n]))
        }
    }()


    buf.SetBlocking(true)
    for i := 0; i < 5; i++ {
        msg := []byte(fmt.Sprintf("message %d", i))
        if _, err := buf.Write(msg); err != nil {
            log.Println("Write error:", err)
        }
    }

    buf.Close()
}
```
## Blocking Modes

By default, `circulis` operates in **non-blocking** mode:

- `Write()` returns `ErrFull` if the buffer is full.
- `Read()` returns `ErrEmpty` if there is no data.

To enable **blocking mode**, call:

```go
buf.SetBlocking(true)
```

## Errors

* ```circulis.ErrFull``` – Write failed: buffer is full (non-blocking mode)

* ```circulis.ErrEmpty``` – Read failed: buffer is empty (non-blocking mode)

* ```circulis.ErrClosed``` – Operation failed: buffer is closed

## Benchmark

Run:
```go test -bench=. -benchmem```

Musixal/circulis:
```
goos: darwin
goarch: arm64
pkg: github.com/Musixal/circulis
cpu: Apple M1
BenchmarkRingBuffer_Sync-8                      30080002                39.63 ns/op            0 B/op          0 allocs/op
BenchmarkRingBuffer_AsyncRead-8                 22569682                83.43 ns/op            0 B/op          0 allocs/op
BenchmarkRingBuffer_AsyncReadBlocking-8          8458837               146.0 ns/op             0 B/op          0 allocs/op
BenchmarkRingBuffer_AsyncWrite-8                18759770                61.38 ns/op            0 B/op          0 allocs/op
BenchmarkRingBuffer_AsyncWriteBlocking-8        10000000               161.9 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/Musixal/circulis     8.540s
```

smallnest/ringbuffer:
```
goarch: arm64
pkg: github.com/smallnest/ringbuffer
cpu: Apple M1
BenchmarkRingBuffer_Sync-8                 	21460840	        55.72 ns/op	       0 B/op	       0 allocs/op
BenchmarkRingBuffer_AsyncRead-8            	13819650	       109.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkRingBuffer_AsyncReadBlocking-8    	 5173066	       215.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkRingBuffer_AsyncWrite-8           	16859881	       132.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkRingBuffer_AsyncWriteBlocking-8   	 6520000	       284.9 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/smallnest/ringbuffer	25.029s
```

### 📊 Performance Comparison (Apple M1, `go test -bench`)

| Benchmark                    | circulis (ns/op) | ringbuffer (ns/op) | Improvement       |
|-----------------------------|------------------|---------------------|-------------------|
| **Sync**                    | 39.63            | 55.72               | ✅ **28.9% faster** |
| **AsyncRead**               | 83.43            | 109.6               | ✅ **23.9% faster** |
| **AsyncReadBlocking**       | 146.0            | 215.7               | ✅ **32.3% faster** |
| **AsyncWrite**              | 61.38            | 132.5               | ✅ **53.7% faster** |
| **AsyncWriteBlocking**      | 161.9            | 284.9               | ✅ **43.2% faster** |

> All tests: 0 B allocations per op · Architecture: `arm64` · CPU: Apple M1

---

### ✅ Summary

- **Musixal/circulis** outperforms `smallnest/ringbuffer` in every benchmarked mode.
- Particularly efficient in **write-heavy** and **blocking** use cases.
- Ideal for **low-latency pipelines**, **streaming systems**, and **concurrent buffers**.


## 🚀 Future Improvements
While circulis already delivers strong performance, especially under contention, there is still room for optimization—particularly around:

* Reducing mutex contention
* Improving fairness in blocking mode
* Potential lock-free variants for SPSC/MPMC cases

Pull requests, performance ideas, and contributions are very welcome! 🙌
Let’s make circulis even faster and more robust together.