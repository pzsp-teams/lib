package util

import (
	"sync/atomic"
	"testing"
)

func TestSyncRunner_RunAndWait(t *testing.T) {
	var called int32
	r := &SyncRunner{}

	r.Run(func() {
		atomic.AddInt32(&called, 1)
	})
	r.Wait() 

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected called=1, got %d", called)
	}
}

func TestAsyncRunner_RunAndWait(t *testing.T) {
	var called int32
	r := &AsyncRunner{}

	const n = 10
	for i := 0; i < n; i++ {
		r.Run(func() {
			atomic.AddInt32(&called, 1)
		})
	}

	r.Wait()

	if atomic.LoadInt32(&called) != n {
		t.Fatalf("expected %d calls, got %d", n, called)
	}
}

func TestAsyncRunner_WaitWithNoTasks(t *testing.T) {
	var r AsyncRunner
	// Should not deadlock
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
		t.Fatalf("Run should not execute function synchronously")
	default:
	}

	close(block)
	r.Wait()

	select {
	case <-done:
		// success
	default:
		t.Fatalf("expected function to finish after Wait()")
	}
}



