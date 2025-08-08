package transport

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/neurosimio/simsdk-go"
	"github.com/stretchr/testify/require"
)

type chanReceiver struct{ ch chan simsdk.SimMessage }

func (c *chanReceiver) Start(context.Context) error              { return nil }
func (c *chanReceiver) Stop(context.Context) error               { close(c.ch); return nil }
func (c *chanReceiver) GetInboundChan() <-chan simsdk.SimMessage { return c.ch }

type sinkSender struct {
	mu  sync.Mutex
	log []*simsdk.SimMessage
}

func (s *sinkSender) Send(m *simsdk.SimMessage) error {
	s.mu.Lock()
	s.log = append(s.log, m)
	s.mu.Unlock()
	return nil
}
func (s *sinkSender) ComponentID() string { return "sink" }

func TestForwarder_BasicAndTransformAndCancel(t *testing.T) {
	t.Parallel()

	// --- phase 1: no transform ---
	ctx1, cancel1 := context.WithCancel(context.Background())
	rcv1 := &chanReceiver{ch: make(chan simsdk.SimMessage, 4)}
	snd := &sinkSender{}
	fwd := &Forwarder{}

	fwd.Start(ctx1, rcv1, snd, nil)
	rcv1.ch <- simsdk.SimMessage{MessageID: "a", MessageType: "one"}
	rcv1.ch <- simsdk.SimMessage{MessageID: "b", MessageType: "two"}

	// wait for messages to be delivered
	require.Eventually(t, func() bool {
		snd.mu.Lock()
		defer snd.mu.Unlock()
		return len(snd.log) == 2
	}, 200*time.Millisecond, 10*time.Millisecond)

	snd.mu.Lock()
	require.Equal(t, "a", snd.log[0].MessageID)
	require.Equal(t, "b", snd.log[1].MessageID)
	snd.mu.Unlock()

	// stop phase 1 forwarder cleanly
	cancel1()
	close(rcv1.ch) // ensures no further reads are possible
	time.Sleep(50 * time.Millisecond)

	// still exactly two messages
	snd.mu.Lock()
	require.Len(t, snd.log, 2)
	snd.log = nil // reset for next phase
	snd.mu.Unlock()

	// --- phase 2: with transform ---
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	rcv2 := &chanReceiver{ch: make(chan simsdk.SimMessage, 4)} // fresh channel
	fwd2 := &Forwarder{}

	transform := func(m *simsdk.SimMessage) *simsdk.SimMessage {
		if m.Metadata == nil {
			m.Metadata = map[string]string{}
		}
		m.Metadata["x"] = "y"
		return m
	}

	fwd2.Start(ctx2, rcv2, snd, transform)
	rcv2.ch <- simsdk.SimMessage{MessageID: "c", MessageType: "three"}

	// wait for transformed message
	require.Eventually(t, func() bool {
		snd.mu.Lock()
		defer snd.mu.Unlock()
		return len(snd.log) == 1
	}, 200*time.Millisecond, 10*time.Millisecond)

	snd.mu.Lock()
	require.Equal(t, "c", snd.log[0].MessageID)
	require.Equal(t, "y", snd.log[0].Metadata["x"])
	snd.mu.Unlock()
}
