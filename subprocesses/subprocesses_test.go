package subprocesses_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/libocr/subprocesses"
)

func TestSubprocesses_Go(t *testing.T) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var counter int

	s := &subprocesses.Subprocesses{}
	for i := 0; i < 10; i++ {
		s.Go(func() {
			time.Sleep(time.Millisecond)
			mu.Lock()
			counter++
			mu.Unlock()
			wg.Done()
		})
		wg.Add(1)
	}
	wg.Wait()

	assert.Equal(t, 10, counter, "Expected counter to be 10")
}

func TestSubprocesses_Wait(t *testing.T) {
	var wg sync.WaitGroup
	var counter int

	s := &subprocesses.Subprocesses{}
	for i := 0; i < 10; i++ {
		s.Go(func() {
			time.Sleep(time.Millisecond)
			wg.Done()
		})
		wg.Add(1)
	}
	s.Wait()

	assert.Equal(t, 0, counter, "Expected WaitGroupCounter to be 0")
}

func TestSubprocesses_BlockForAtMost(t *testing.T) {
	s := &subprocesses.Subprocesses{}

	// Test with timeout shorter than function duration
	start := time.Now()
	ok := s.BlockForAtMost(context.Background(), time.Millisecond*500, func(ctx context.Context) {
		select {
		case <-time.After(time.Second):
			// Wait for longer than the timeout
		case <-ctx.Done():
			// Return when the context is cancelled
		}
	})
	elapsed := time.Since(start)

	assert.False(t, ok, "Expected BlockForAtMost to return false")
	assert.LessOrEqual(t, elapsed, time.Millisecond*700, "Expected elapsed time to be less than 700ms")

	// Test with timeout longer than function duration
	start = time.Now()
	ok = s.BlockForAtMost(context.Background(), time.Second, func(ctx context.Context) {
		select {
		case <-time.After(time.Millisecond * 500):
			// Wait for shorter than the timeout
		case <-ctx.Done():
			// Return when the context is cancelled
		}
	})
	elapsed = time.Since(start)

	assert.True(t, ok, "Expected BlockForAtMost to return true")
	assert.GreaterOrEqual(t, elapsed, time.Millisecond*500, "Expected elapsed time to be at least 500ms")
}

func TestSubprocesses_BlockForAtMostMany(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	fs := []func(context.Context){
		func(ctx context.Context) {
			select {
			case <-time.After(time.Second):
				// Wait for longer than the timeout
			case <-ctx.Done():
				// Return when the context is cancelled
			}
		},
		func(ctx context.Context) {
			select {
			case <-time.After(time.Millisecond * 500):
				// Wait for shorter than the timeout
			case <-ctx.Done():
				// Return when the context is cancelled
			}
		},
		func(ctx context.Context) {
			select {
			case <-time.After(time.Millisecond * 750):
				// Wait for longer than the timeout
			case <-ctx.Done():
				// Return when the context is cancelled
			}
		},
	}

	// Test with timeout shorter than longest function duration
	start := time.Now()
	ok, oks := s.BlockForAtMostMany(context.Background(), time.Millisecond*700, fs...)
	elapsed := time.Since(start)

	assert.False(t, ok, "Expected BlockForAtMostMany to return false")
	assert.Equal(t, []bool{false, true, false}, oks, "Expected oks to be [false, true, false]")
	assert.LessOrEqual(t, elapsed, time.Millisecond*1000, "Expected elapsed time to be less than 1000ms")

	// Test with timeout longer than longest function duration
	start = time.Now()
	ok, oks = s.BlockForAtMostMany(context.Background(), time.Second, fs...)
	elapsed = time.Since(start)

	assert.True(t, ok, "Expected BlockForAtMostMany to return true")
	assert.Equal(t, []bool{false, true, true}, oks, "Expected oks to be [false, true, true]")
	assert.GreaterOrEqual(t, elapsed, time.Millisecond*750, "Expected elapsed time to be at least 750ms")
}

func TestSubprocesses_RepeatWithCancel(t *testing.T) {
	var mu sync.Mutex
	var counter int

	s := &subprocesses.Subprocesses{}
	ctx, cancel := context.WithCancel(context.Background())
	s.RepeatWithCancel("test", time.Millisecond*500, ctx, func() {
		mu.Lock()
		counter++
		mu.Unlock()
	})

	// Wait for a short time to allow some repetitions
	time.Sleep(time.Millisecond * 1500)
	cancel()

	assert.GreaterOrEqual(t, counter, 2, "Expected counter to be at least 2")
}

func TestSubprocesses_BlockForAtMost_EdgeCases(t *testing.T) {
	s := &subprocesses.Subprocesses{}

	// Test with zero duration
	ok := s.BlockForAtMost(context.Background(), 0, func(ctx context.Context) {})
	assert.True(t, ok, "Expected BlockForAtMost with 0 duration to return true")

	// Test with negative duration
	ok = s.BlockForAtMost(context.Background(), -1*time.Second, func(ctx context.Context) {})
	assert.True(t, ok, "Expected BlockForAtMost with negative duration to return true")

	// Test with nil function
	ok = s.BlockForAtMost(context.Background(), time.Second, nil)
	assert.False(t, ok, "Expected BlockForAtMost with nil function to return false")
}

func TestSubprocesses_Go_Concurrency(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	var mu sync.Mutex
	var counter int

	// Start 100 goroutines that each increment a shared variable
	for i := 0; i < 100; i++ {
		s.Go(func() {
			time.Sleep(time.Millisecond)
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	// Wait for all goroutines to finish
	s.Wait()

	assert.Equal(t, 100, counter, "Expected counter to be 100")
}

func TestSubprocesses_BlockForAtMostMany_Concurrency(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	var mu sync.Mutex
	var counter int
	fs := []func(context.Context){
		func(ctx context.Context) {
			time.Sleep(time.Millisecond)
			mu.Lock()
			counter++
			mu.Unlock()
		},
		func(ctx context.Context) {
			time.Sleep(time.Millisecond * 50)
			mu.Lock()
			counter++
			mu.Unlock()
		},
		func(ctx context.Context) {
			time.Sleep(time.Millisecond * 100)
			mu.Lock()
			counter++
			mu.Unlock()
		},
	}

	// Start 100 goroutines that each call BlockForAtMostMany with a slice of functions
	for i := 0; i < 100; i++ {
		s.Go(func() {
			ok, _ := s.BlockForAtMostMany(context.Background(), time.Millisecond*500, fs...)
			if ok {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		})
	}

	// Wait for all goroutines to finish
	s.Wait()

	assert.GreaterOrEqual(t, counter, 100, "Expected counter to be at least 100")
}

func TestSubprocesses_RepeatWithCancel_Cancel(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	var mu sync.Mutex
	var counter int

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Start a goroutine that calls RepeatWithCancel with the cancelled context
	s.Go(func() {
		s.RepeatWithCancel("test", time.Millisecond*500, ctx, func() {
			mu.Lock()
			counter++
			mu.Unlock()
		})
	})

	// Wait for the goroutine to finish
	s.Wait()

	assert.Equal(t, 0, counter, "Expected counter to be 0")
}

func TestSubprocesses_BlockForAtMostMany_Cancel(t *testing.T) {
	s := &subprocesses.Subprocesses{}
	var mu sync.Mutex
	var counter int

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Start a goroutine that calls BlockForAtMostMany with the cancelled context and a slice of functions
	s.Go(func() {
		ok, _ := s.BlockForAtMostMany(ctx, time.Millisecond*500, func(ctx context.Context) {
			mu.Lock()
			counter++
			mu.Unlock()
		})
		assert.False(t, ok, "Expected BlockForAtMostMany with cancelled context to return false")
	})

	// Wait for the goroutine to finish
	s.Wait()

	assert.Equal(t, 0, counter, "Expected counter to be 0")
}
