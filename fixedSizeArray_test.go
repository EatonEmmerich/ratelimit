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
    l := ratelimit.NewFixedSizeSlice(time.Second, 10)
    resource := fmt.Sprintf("%d", rand.Int())
    ratelimit.PrimeForTesting(t, l, resource)

    t.Run(fmt.Sprintf("limiting"),func(t * testing.T) {
        testFixedSizeArray(t, l, resource)
    })

    t.Run(fmt.Sprintf("unlimiting"), func(t *testing.T){
        ratelimit.CreditForTesting(t, l, resource)
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
    ratelimit.PrimeForTesting(t, l, resource)
    t.Run(fmt.Sprintf("limiting"),func(t * testing.T) {
        testFixedSizeArray(t, l, resource)
    })

    t.Run(fmt.Sprintf("unlimiting"), func(t *testing.T){
        ratelimit.SetNowForTesting(t, func() time.Time{
            return t0.Add(time.Minute*2)
        })
        time.Sleep(time.Second*2)
        require.True(t, l.Request(resource))
    })
}

func BenchmarkFixedSizeArray(b *testing.B) {
    ratelimit.Benchmark(b, func() ratelimit.RateLimiter {
        pl := ratelimit.NewFixedSizeSlice(time.Millisecond, 10)
        for i := 0; i < b.N; i ++ {
            ratelimit.PrimeForTesting(b, pl, fmt.Sprint(i))
        }
        println("done priming")
        return pl
    })
}
