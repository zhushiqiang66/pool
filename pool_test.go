package pool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	N               = 10000
	CapacitySize    = 1024
	ExpiredInSecond = 60
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)

var curMem uint64

func TestPool(t *testing.T) {
	p := New(WithCapacitySize(CapacitySize), WithWorkerExpiredInSecond(ExpiredInSecond))
	defer p.Close()

	wg := sync.WaitGroup{}
	for i := 0; i < N; i++ {
		wg.Add(1)
		p.Submit(func() {
			defer wg.Done()
			time.Sleep(time.Second)
		})
	}

	wg.Wait()
	t.Logf("pool, capacity:%d", p.capacity)
	t.Logf("pool, running workers number:%d", p.Running())
	t.Logf("pool, free workers number:%d", p.Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPoolGetWorkerFromCache(t *testing.T) {
	p := New(WithCapacitySize(CapacitySize))
	defer p.Close()

	for i := 0; i < N; i++ {
		p.Submit(func() { time.Sleep(time.Second) })
	}
	time.Sleep(2)
	p.Submit(func() { time.Sleep(time.Second) })
	t.Logf("pool, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}
