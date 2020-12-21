package ratelimit_test

import (
	"sync"
	"testing"
	"time"

	"github.com/corverroos/ratelimit"
	"github.com/stretchr/testify/require"
)

func TestSyncMapWindow(t *testing.T) {
	l := ratelimit.NewSyncMapWindow(time.Hour, 10)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			require.True(t, l.Request(""))
			wg.Done()
		}()
	}
	wg.Wait()
	// Since SyncMapWindow is not 100%, we cannot do: require.False(t, l.Request(""))
}

func BenchmarkSyncMapWindow(b *testing.B) {
	ratelimit.Benchmark(b, func(m int) ratelimit.RateLimiter {
		return ratelimit.NewSyncMapWindow(time.Millisecond, 10)
	})
}
