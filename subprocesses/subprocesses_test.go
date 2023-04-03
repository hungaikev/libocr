package subprocesses_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/libocr/subprocesses"
)

func TestGoAndWait(t *testing.T) {
	var counter int32
	s := &subprocesses.Subprocesses{}

	for i := 0; i < 10; i++ {
		s.Go(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(50 * time.Millisecond)
		})
	}

	s.Wait()
	assert.Equal(t, int32(10), counter)
}

func TestBlockForAtMost(t *testing.T) {
	s := &subprocesses.Subprocesses{}

	f := func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-time.After(50 * time.Millisecond):
		}
	}

	ctx := context.Background()

	// Test with a function that finishes before the specified duration
	result := s.BlockForAtMost(ctx, 100*time.Millisecond, f)
	assert.True(t, result, "Function should have finished before the specified duration")

	// Test with a function that takes longer than the specified duration
	result = s.BlockForAtMost(ctx, 25*time.Millisecond, f)
	assert.False(t, result, "Function should not have finished before the specified duration")
}

func TestBlockForAtMostMany(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	ctx := context.Background()

	fastFunc := func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-time.After(50 * time.Millisecond):
		}
	}

	slowFunc := func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-time.After(200 * time.Millisecond):
		}
	}

	// Test with multiple functions that finish before the specified duration
	ok, results := s.BlockForAtMostMany(ctx, 100*time.Millisecond, fastFunc, fastFunc, fastFunc)
	assert.True(t, ok, "All functions should have finished before the specified duration")
	assert.Equal(t, []bool{true, true, true}, results)

	// Test with a mix of functions, some finishing before the specified duration and others taking longer
	ok, results = s.BlockForAtMostMany(ctx, 100*time.Millisecond, fastFunc, slowFunc, fastFunc)
	assert.False(t, ok, "Not all functions should have finished before the specified duration")
	assert.Equal(t, []bool{true, false, true}, results)
}

func TestRepeatWithCancel(t *testing.T) {
	var counter int32
	s := &subprocesses.Subprocesses{}
	ctx, cancel := context.WithCancel(context.Background())

	s.RepeatWithCancel("test", 20*time.Millisecond, ctx, func() {
		atomic.AddInt32(&counter, 1)
	})

	time.Sleep(110 * time.Millisecond)
	cancel()
	s.Wait()

	assert.GreaterOrEqual(t, counter, int32(5))
	assert.LessOrEqual(t, counter, int32(6))
}
