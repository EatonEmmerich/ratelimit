package ratelimit_test

import (
	"sync"
	"testing"
	"time"

	"github.com/corverroos/ratelimit"
	"github.com/stretchr/testify/require"
)

func TestNaiveWindow(t *testing.T) {
	l := ratelimit.NewNaiveWindow(time.Hour, 10)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			require.True(t, l.Request(""))
			wg.Done()
		}()
	}
	wg.Wait()
	require.False(t, l.Request(""))
}

func BenchmarkNaiveWindow(b *testing.B) {
	ratelimit.Benchmark(b, func(m int) ratelimit.RateLimiter {
		return ratelimit.NewNaiveWindow(time.Millisecond, 10)
	})
}
