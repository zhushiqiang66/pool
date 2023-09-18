package pool

import "time"

// Option
type Option func(p *Pool)

// WithCapacitySize
func WithCapacitySize(size uint32) Option {
	return func(p *Pool) {
		p.capacity = int32(size)
	}
}

// WithWorkerExpiredInSecond
func WithWorkerExpiredInSecond(second uint32) Option {
	return func(p *Pool) {
		p.expired = time.Second * time.Duration(second)
	}
}
