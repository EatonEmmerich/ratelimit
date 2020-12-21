package ratelimit_test

import (
    "fmt"
    "math/rand"
    "sync"
    "testing"
    "time"

    "github.com/corverroos/ratelimit"
    "github.com/stretchr/testify/require"
)

func TestTest(t *testing.T){
    l := ratelimit.NewFixedSizeSlice(time.Hour, 10)
    resource := fmt.Sprintf("%d", rand.Int())

    t.Run(fmt.Sprintf("limiting"),func(t * testing.T) {
        testFixedSizeArray(t, l, resource)
    })

    t.Run(fmt.Sprintf("unlimiting"), func(t *testing.T){
        ratelimit.CreditForTesting(t, l)
        require.True(t, l.Request(resource))
    })
}

func testFixedSizeArray(t *testing.T, l ratelimit.RateLimiter, resource string) {
    var wg sync.WaitGroup
    wg.Add(10)
    var  count  int
    for i := 0; i < 10; i++ {
        go func() {
            if l.Request(resource){
                count ++
            }
            wg.Done()
        }()
    }
    wg.Wait()
    // accuracy not required.
    require.Less(t,count, 11)
    require.Less(t, 8, count)
    require.False(t, l.Request(resource))
}

func TestGoroutineCredits(t *testing.T){
    t.Skip("test contains sleeps that are slow")
    l := ratelimit.NewFixedSizeSlice(time.Second, 10)
    t0 := time.Now().Truncate(time.Second)
    ratelimit.SetNowForTesting(t, func() time.Time{
        return t0
    })
    resource := fmt.Sprintf("%d", rand.Int())
    t.Run(fmt.Sprintf("limiting"),func(t * testing.T) {
        testFixedSizeArray(t, l, resource)
    })

    t.Run(fmt.Sprintf("unlimiting"), func(t *testing.T){
        ratelimit.SetNowForTesting(t, func() time.Time{
            return t0.Add(time.Minute)
        })
        time.Sleep(time.Second*2)
        require.True(t, l.Request(resource))
    })
}

func BenchmarkFixedSizeArray_NeverLimit(b *testing.B) {
    ratelimit.Benchmark(b, func() ratelimit.RateLimiter {
        pl := ratelimit.NewFixedSizeSlice(time.Millisecond, 16384)
        return pl
    })
}

func BenchmarkFixedSizeArray_MostlyLimit(b *testing.B) {
    ratelimit.Benchmark(b, func() ratelimit.RateLimiter {
        pl := ratelimit.NewFixedSizeSlice(time.Millisecond, 2)
        return pl
    })
}
