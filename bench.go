package ratelimit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Benchmark(b *testing.B, provider func(m int) RateLimiter) {
	// Run the benchmark for increasing number of concurrent goroutines.
	for _, m := range []int{1, 4, 16, 64, 256, 1024, 4096, 16384} {
		b.Run(fmt.Sprintf("Concurrency_%d", m), benchmark(provider, m))
	}
}

func benchmark(provider func(m int) RateLimiter, m int) func(b *testing.B) {
	return func(b *testing.B) {
		limiter := provider(m)
		var wg sync.WaitGroup
		wg.Add(m)
		var total int64
		for i := 0; i < m; i++ {
			go func(n int) {
				t0 := time.Now()
				for i := 0; i < n; i++ {
					limiter.Request(fmt.Sprint(i))
				}
				atomic.AddInt64(&total, time.Since(t0).Nanoseconds())
				wg.Done()
			}(b.N)
		}
		wg.Wait()
		b.ReportMetric(float64(atomic.LoadInt64(&total))/float64(m*b.N), "ns/op")
	}
}
