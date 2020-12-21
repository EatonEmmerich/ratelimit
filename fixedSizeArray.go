package ratelimit

import (
    "sync"
    "testing"
    "time"
)

type limitArray []bool

var now = time.Now

type limitStruct struct {
    availableList limitArray
    index int
    writeLock sync.Mutex
}

type fixedSizeSlice struct {
    limitSlice []*limitStruct
    resourceMap map[string]int
    rwLock sync.RWMutex
     period time.Duration
}

var maxLimit int

// NewFixedSizeSlice returns a fixed size array limiter.
// mallocs should only take place once whenever the request is first made,
// but still give us access to a scalable array.
func NewFixedSizeSlice(period time.Duration, limit int) *fixedSizeSlice {
    maxLimit = limit -1
    fsslice :=  &fixedSizeSlice{
        resourceMap: make(map[string]int),
        period : period,
    }

    return fsslice
}

func (m *fixedSizeSlice) Request(resource string) bool {
    m.rwLock.RLock()
    id, ok := m.resourceMap[resource]
    m.rwLock.RUnlock()
    if !ok {
        la := initialize(m.period)
        m.rwLock.Lock()
        id = len(m.limitSlice)
        m.limitSlice = append(m.limitSlice, la)
        m.resourceMap[resource] = id
        m.rwLock.Unlock()
    }
    ls := m.limitSlice[id]

    if  !ls.availableList.available(ls.index) {
        return false
    }

    return  ls.use()
}

func initialize(period time.Duration) *limitStruct{
    ls :=  &limitStruct{
        availableList: make(limitArray, maxLimit+1),
    }
    for l := range ls.availableList {
        ls.availableList[l] = true
    }
    go removeLastLimitForever(ls, period)
    return ls
}

func removeLastLimitForever(ls *limitStruct, period time.Duration){
    lastRun := now()
    sleeptime := time.Second
    for{
        time.Sleep(sleeptime)
        if now().Before(lastRun.Add(period)) {
            continue
        }
        creditCount := (now().Unix() - lastRun.Unix())/int64(period.Seconds())
        for i := int64(0); i < creditCount; i ++{
            credit(ls)
        }
        lastRun = now()
    }
}

func credit(ls *limitStruct){
    ls.writeLock.Lock()
    defer ls.writeLock.Unlock()
    if ls.index -1 < 0{
        return
    }
    if ls.availableList[ls.index] {
        ls.index --
    }
    ls.availableList[ls.index] = true
}

func CreditForTesting(t *testing.T, l *fixedSizeSlice, resource string){
    id := l.resourceMap[resource]
    credit(l.limitSlice[id])
}

func (ls limitArray) available(index int) bool {
    return ls[index]
}

func (ls *limitStruct) use() bool {
    ls.writeLock.Lock()
    defer     ls.writeLock.Unlock()
    if ls.index >= maxLimit{
        return false
    }
    ls.availableList[ls.index] = false
    ls.index ++
    return true
}

var _ RateLimiter = (*fixedSizeSlice)(nil)

// not concurrent safe
func PrimeForTesting(t testing.TB, l *fixedSizeSlice, resource string) {
    _, ok := l.resourceMap[resource]
    if ok {
        return
    }
    l.resourceMap[resource] = len(l.limitSlice)
    l.limitSlice = append(l.limitSlice, initialize(time.Second))
}

func SetNowForTesting(t testing.TB, n func() time.Time){
    t.Cleanup(func() {
        now = time.Now
    })
    now = n
}