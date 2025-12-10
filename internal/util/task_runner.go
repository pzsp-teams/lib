package util

import "sync"

type TaskRunner interface {
	Run(fn func())
	Wait()
}

type SyncRunner struct{}

func (*SyncRunner) Run(fn func()) {
	fn()
}

func (*SyncRunner) Wait() {}

type AsyncRunner struct {
	wg sync.WaitGroup
}

func (r *AsyncRunner) Run(fn func()) {
	r.wg.Go(func() {
		fn()
	})
}

func (r *AsyncRunner) Wait() {
	r.wg.Wait()
}
