package genpool

import (
	"context"
)

type Pool[T any] struct {
	ch       chan T
	ready    chan struct{}
	waiting  chan struct{}
	resetter func(T) error
}

func NewPool[T any](poolSize int, seeder func() (T, error), resetter func(T) error) (*Pool[T], error) {
	p := &Pool[T]{
		ch:       make(chan T, poolSize),
		ready:    make(chan struct{}),
		waiting:  make(chan struct{}, poolSize),
		resetter: resetter,
	}

	for i := 0; i < poolSize; i++ {
		seed, err := seeder()
		if err != nil {
			return nil, err
		}

		p.ch <- seed
	}

	return p, nil
}

func (p *Pool[T]) Take(ctx context.Context) (T, error) {
	if len(p.ch) == 0 {
		p.waiting <- struct{}{}

	outer:
		for {
			select {
			case <-p.ready:
				break outer
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			}
		}
	}

	return <-p.ch, nil
}

func (p *Pool[T]) Release(id T) error {
	if p.resetter != nil {
		if err := p.resetter(id); err != nil {
			return err
		}
	}

	p.ch <- id

	if len(p.waiting) > 0 {
		<-p.waiting
		p.ready <- struct{}{}
	}

	return nil
}
