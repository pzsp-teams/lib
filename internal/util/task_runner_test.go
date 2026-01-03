package util

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskRunners_RunAndWait(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		runner TaskRunner
		n      int
	}{
		{"SyncRunner", &SyncRunner{}, 5},
		{"AsyncRunner", &AsyncRunner{}, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var called int32

			for i := 0; i < tt.n; i++ {
				tt.runner.Run(func() {
					atomic.AddInt32(&called, 1)
				})
			}

			tt.runner.Wait()
			require.Equal(t, int32(tt.n), atomic.LoadInt32(&called))
		})
	}
}

func TestAsyncRunner_WaitWithNoTasks(t *testing.T) {
	t.Parallel()

	var r AsyncRunner
	r.Wait()
}

func TestAsyncRunner_RunIsNonBlocking(t *testing.T) {
	t.Parallel()

	r := &AsyncRunner{}

	block := make(chan struct{})
	done := make(chan struct{})

	r.Run(func() {
		<-block
		close(done)
	})

	require.Never(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 5*time.Millisecond, "Run should not execute function synchronously")

	close(block)
	r.Wait()

	require.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond, "expected function to finish after Wait()")
}

