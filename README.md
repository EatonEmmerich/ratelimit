# ratelimit

Package `ratelimit` provides a simple "resource based" rate limit API and multiple implementations.

The aim is to find a performant implementation to use as generic logging rate limiter for Luno microservices. We want to limit specific logs from spamming the infrastructure, identified either by unique message or source. Performance is more important than correctness in this use case since we expect high level of concurrency and slight over (or under) limiting is not an issue.

```
// RateLimiter is the interface implemented by a resource rate limiter.
type RateLimiter interface {
	// Request returns true if the request of the resource conforms
	// to the rate limit; it contributes to the rate. It returns false
	// if the request would exceed the rate limit; it doesn't
	// contributes to the rate.
	Request(resource string) bool
}
```

The package also provides performance benchmarking; number of operations per number of concurrent goroutines. The following are results of `ratelimit.Benchmark` for different implementations with config aimed at `limit=100, period=10ms` executed on MacBook Pro Quad Core 16GB RAM: 

|Implementation | 1 | 4 | 16 | 64 | 256 | 1024 | 4096 | 16384 |
|---|--|--|--|--|--|--|--|--|
|github.com/corver/ratelimit.NaiveWindow | 341 | 1_581 | 6_520 | 25_084 | 108_735 | 389_116 | 1_410_022 | 4_813_929 |
|github.com/corver/ratelimit.SyncMapWindow | 609 | 2_056 | 14_259 | 53_523 | 222_216 | 756_072 | 1_969_951 | 3_686_225 |
|github.com/eatonemmerich/ratelimit.FixedSizeSlice| 866 | 2_317 | 6_597 | 18_224 | 81_816 | 233_635 | 870_058 | 2_401_010 |

Please DM me link to your implementation for me to add your benchmark score.
