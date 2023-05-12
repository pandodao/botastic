package starter

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Starter interface {
	Start(ctx context.Context) error
}

var _ Starter = new(multiStarter)

type multiStarter struct {
	fs []func(context.Context) error
}

func Multi(starters ...Starter) Starter {
	m := &multiStarter{}
	for _, s := range starters {
		switch v := s.(type) {
		case *multiStarter:
			m.fs = append(m.fs, v.fs...)
		default:
			m.fs = append(m.fs, s.Start)
		}
	}

	return m
}

func MultiFunc(fs ...func(ctx context.Context) error) Starter {
	return &multiStarter{
		fs: fs,
	}
}

func (ms *multiStarter) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(len(ms.fs))

	var g *errgroup.Group
	g, ctx = errgroup.WithContext(ctx)

	for _, f1 := range ms.fs {
		f := f1
		g.Go(func() error {
			return f(ctx)
		})
	}

	return g.Wait()
}
