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

type resourceStruct struct{
    worker int
    limit int
}

type fixedSizeSlice struct {
    limitSlice []*limitStruct
    resourceMap     map[string]resourceStruct
    resourceMapLock multiReadLocker
    period          time.Duration

}

type multiReadLocker struct {
    count int
    writeLockStatus bool
    writeLock       sync.Mutex
    countLock       sync.Mutex
    writeStatusLock sync.Mutex
    readStopLock    sync.Mutex
}

func (mrl *multiReadLocker) Lock(){
    mrl.writeLock.Lock()
    mrl.writeStatusLock.Lock()
    mrl.writeLockStatus = true
    mrl.writeStatusLock.Unlock()
    mrl.readStopLock.Lock()
}

func (mrl *multiReadLocker) Unlock(){
    mrl.writeLock.Unlock()
    mrl.writeStatusLock.Lock()
    defer mrl.writeStatusLock.Unlock()
    mrl.writeLockStatus = false
    mrl.readStopLock.Unlock()
}

func (mrl *multiReadLocker) RLock(resource string){
    mrl.writeStatusLock.Lock()
    if mrl.writeLockStatus{
        mrl.writeStatusLock.Unlock()
        mrl.Lock()
        mrl.Unlock()
    } else {
        mrl.writeStatusLock.Unlock()
    }

    mrl.countLock.Lock()
    if mrl.count == 0 {
        mrl.readStopLock.Lock()
    }
    mrl.count ++
    mrl.countLock.Unlock()
}

func (mrl *multiReadLocker) RUnlock(resource string){
    mrl.countLock.Lock()
    defer mrl.countLock.Unlock()
    mrl.count --
    if mrl.count == 0{
        mrl.readStopLock.Unlock()
    }
}

var maxLimit int

// NewFixedSizeSlice returns a fixed size array limiter.
// mallocs should only take place once whenever the request is first made,
// but still give us access to a scalable array.
func NewFixedSizeSlice(period time.Duration, limit int) *fixedSizeSlice {
    maxLimit = limit -1
    fsslice :=  &fixedSizeSlice{
        resourceMap: make(map[string]resourceStruct),
        period : period,
    }
    initialize(fsslice)
    return fsslice
}

func (m *fixedSizeSlice) Request(resource string) bool {
    ls := getResource(m, resource)

    if  !ls.availableList.available(ls.index) {
        return false
    }

    return  ls.use()
}

func getResource(m *fixedSizeSlice, resource string) *limitStruct {
    m.resourceMapLock.RLock(resource)
    resourceObj, ok := m.resourceMap[resource]
    id := resourceObj.limit
    m.resourceMapLock.RUnlock(resource)
    if !ok {
        m.resourceMapLock.Lock()
        id = len(m.limitSlice)

        m.limitSlice = append(m.limitSlice, newLS())
        resourceObj = resourceStruct{worker: 0, limit: id}
        m.resourceMap[resource] = resourceObj
        m.resourceMapLock.Unlock()
    }
    return m.limitSlice[id]
}

func newLS() *limitStruct{
    ls :=  &limitStruct{
        availableList: make(limitArray, maxLimit+1),
    }
    for l := range ls.availableList {
        ls.availableList[l] = true
    }
    return ls
}

func initialize(m *fixedSizeSlice){
    go removeLastLimitForever(m)
}

func removeLastLimitForever(m *fixedSizeSlice){
    lastRun := now()
    sleeptime := time.Second
    for{
        time.Sleep(sleeptime)
        if now().Before(lastRun.Add(m.period)) {
            continue
        }

        creditCount := int64(1)
        if m.period.Seconds() <= 0 {
            creditCount = (now().Unix() - lastRun.Unix()) / int64(m.period.Seconds())
        }
        for i := int64(0); i < creditCount; i ++{
            credit(m)
        }
        lastRun = now()
    }
}

func credit(m *fixedSizeSlice) {
    for _, ls := range (m.limitSlice) {
        ls.writeLock.Lock()
        defer ls.writeLock.Unlock()
        if ls.index-1 < 0 {
            return
        }
        if ls.availableList[ls.index] {
            ls.index--
        }
        ls.availableList[ls.index] = true
    }
}

func CreditForTesting(t testing.TB, l *fixedSizeSlice){
    credit(l)
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

func SetNowForTesting(t testing.TB, n func() time.Time){
    t.Cleanup(func() {
        now = time.Now
    })
    now = n
}