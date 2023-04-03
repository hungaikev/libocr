package ragep2p

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStreamID(id int) streamID {
	var sid streamID
	sid[0] = byte(id)
	return sid
}
func TestNewDemuxer(t *testing.T) {
	d := newDemuxer()
	assert.NotNil(t, d)
	assert.Len(t, d.streams, 0)
}

func TestAddStream(t *testing.T) {
	// Test data
	sid := newTestStreamID(1)
	incomingBufferSize := 10
	maxMessageSize := 1024
	messagesLimit := TokenBucketParams{Rate: 10, Capacity: 5}
	bytesLimit := TokenBucketParams{Rate: 10000, Capacity: 5000}

	d := newDemuxer()
	result := d.AddStream(sid, incomingBufferSize, maxMessageSize, messagesLimit, bytesLimit)

	assert.True(t, result)
	require.Contains(t, d.streams, sid)

	s := d.streams[sid]
	assert.NotNil(t, s)
	assert.Equal(t, maxMessageSize, s.maxMessageSize)
}

func TestAddDuplicateStream(t *testing.T) {
	sid := newTestStreamID(1)

	d := newDemuxer()
	d.AddStream(sid, 10, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})
	result := d.AddStream(sid, 20, 512, TokenBucketParams{Rate: 20, Capacity: 10}, TokenBucketParams{Rate: 20000, Capacity: 10000})

	assert.False(t, result)
}

func TestRemoveStream(t *testing.T) {
	sid := newTestStreamID(1)

	d := newDemuxer()
	d.AddStream(sid, 10, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})
	d.RemoveStream(sid)

	assert.NotContains(t, d.streams, sid)
}

func TestRemoveNonexistentStream(t *testing.T) {
	sid := newTestStreamID(1)

	d := newDemuxer()
	d.RemoveStream(sid)

	assert.Len(t, d.streams, 0)
}

func TestShouldPush(t *testing.T) {
	sid := newTestStreamID(1)
	d := newDemuxer()
	d.AddStream(sid, 10, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})

	// Valid size
	result := d.ShouldPush(sid, 100)
	assert.Equal(t, shouldPushResultYes, result)

	// Message too big
	result = d.ShouldPush(sid, 2000)
	assert.Equal(t, shouldPushResultMessageTooBig, result)

	// Message limit exceeded
	for i := 0; i < 5; i++ {
		d.streams[sid].messagesLimiter.RemoveTokens(1)
	}
	result = d.ShouldPush(sid, 100)
	assert.Equal(t, shouldPushResultMessagesLimitExceeded, result)

	// Unknown stream
	result = d.ShouldPush(newTestStreamID(2), 100)
	assert.Equal(t, shouldPushResultUnknownStream, result)
}

func TestPushMessage(t *testing.T) {
	sid := newTestStreamID(1)
	d := newDemuxer()
	d.AddStream(sid, 2, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})

	// Valid push
	msg := []byte("test message")
	result := d.PushMessage(sid, msg)
	assert.Equal(t, pushResultSuccess, result)

	// Buffer capacity exceeded
	msg2 := []byte("test message 2")
	d.PushMessage(sid, msg2)
	result = d.PushMessage(sid, []byte("test message 3"))
	assert.Equal(t, pushResultDropped, result)

	// Unknown stream
	result = d.PushMessage(newTestStreamID(2), msg)
	assert.Equal(t, pushResultUnknownStream, result)
}

func TestPopMessage(t *testing.T) {
	sid := newTestStreamID(1)
	d := newDemuxer()
	d.AddStream(sid, 2, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})

	msg := []byte("test message")
	d.PushMessage(sid, msg)
	poppedMsg := d.PopMessage(sid)

	assert.Equal(t, msg, poppedMsg)

	// Unknown stream
	poppedMsg = d.PopMessage(newTestStreamID(2))
	assert.Nil(t, poppedMsg)
}

func TestSignalPending(t *testing.T) {
	sid := newTestStreamID(1)
	d := newDemuxer()
	d.AddStream(sid, 2, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})

	signal := d.SignalPending(sid)
	assert.NotNil(t, signal)

	// Test with pending message
	msg := []byte("test message")
	d.PushMessage(sid, msg)

	select {
	case <-signal:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "SignalPending did not receive a signal")
	}

	// Test without pending message
	d.PopMessage(sid)
	select {
	case <-signal:
		assert.Fail(t, "SignalPending should not have received a signal")
	case <-time.After(1 * time.Second):
		// no signal received, as expected
	}

	// Test with unknown stream
	signal = d.SignalPending(newTestStreamID(2))
	assert.Nil(t, signal)
}

func TestConcurrency(t *testing.T) {
	sid := newTestStreamID(1)
	d := newDemuxer()
	d.AddStream(sid, 10, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})

	var wg sync.WaitGroup

	// Concurrently push messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			d.PushMessage(sid, []byte("test message"))
		}
	}()

	// Concurrently pop messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			d.PopMessage(sid)
		}
	}()

	// Concurrently add and remove streams
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			d.AddStream(newTestStreamID(i+2), 10, 1024, TokenBucketParams{Rate: 10, Capacity: 5}, TokenBucketParams{Rate: 10000, Capacity: 5000})
			d.RemoveStream(newTestStreamID(i + 2))
		}
	}()

	wg.Wait()
}
