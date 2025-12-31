package util

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncRunner_RunAndWait(t *testing.T) {
	var called int32
	r := &SyncRunner{}

	r.Run(func() {
		atomic.AddInt32(&called, 1)
	})
	r.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestAsyncRunner_RunAndWait(t *testing.T) {
	var called int32
	r := &AsyncRunner{}

	const n = 10
	for range n {
		r.Run(func() {
			atomic.AddInt32(&called, 1)
		})
	}

	r.Wait()

	assert.Equal(t, int32(n), atomic.LoadInt32(&called))
}

func TestAsyncRunner_WaitWithNoTasks(_ *testing.T) {
	var r AsyncRunner
	r.Wait()
}

func TestAsyncRunner_RunIsNonBlocking(t *testing.T) {
	r := &AsyncRunner{}

	block := make(chan struct{})
	done := make(chan struct{})

	r.Run(func() {
		<-block
		close(done)
	})

	// Check that Run does not block
	select {
	case <-done:
		assert.Fail(t, "Run should not execute function synchronously")
	default:
	}

	close(block)
	r.Wait()

	select {
	case <-done:
		// success
	default:
		assert.Fail(t, "expected function to finish after Wait()")
	}
}
