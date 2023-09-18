package pool

import (
	"sync"
	"testing"
	"time"
)

const (
	RunTimes           = 1e6
	PoolCap            = 5e4
	BenchParam         = 10
	DefaultExpiredTime = 10 // second
)

func Demo() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				Demo()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkChannel(b *testing.B) {
	var wg sync.WaitGroup
	seam := make(chan struct{}, PoolCap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			seam <- struct{}{}
			go func() {
				Demo()
				<-seam
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	p := New(WithCapacitySize(PoolCap), WithWorkerExpiredInSecond(uint32(DefaultExpiredTime)))
	defer p.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			p.Submit(func() {
				Demo()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkGoroutinesThroughput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go Demo()
		}
	}
}

func BenchmarkSemaphoreThroughput(b *testing.B) {
	seam := make(chan struct{}, PoolCap)
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			seam <- struct{}{}
			go func() {
				Demo()
				<-seam
			}()
		}
	}
}

func BenchmarkPoolThroughput(b *testing.B) {
	p := New(WithCapacitySize(PoolCap), WithWorkerExpiredInSecond(uint32(DefaultExpiredTime)))
	defer p.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			p.Submit(Demo)
		}
	}
}
