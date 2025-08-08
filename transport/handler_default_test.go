package transport

import (
	"testing"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/require"
)

type capSender struct {
	id     string
	called int
	last   *simsdk.SimMessage
}

func (c *capSender) Send(m *simsdk.SimMessage) error { c.called++; c.last = m; return nil }
func (c *capSender) ComponentID() string             { return c.id }

func TestDefaultPerInstanceStreamHandler(t *testing.T) {
	h := &DefaultPerInstanceStreamHandler{}
	s := &capSender{id: "comp-1"}

	// SetStreamSender stores the sender
	h.SetStreamSender(s)
	require.Equal(t, s, h.sender)

	// SetStreamSender again replaces the sender
	s2 := &capSender{id: "comp-2"}
	h.SetStreamSender(s2)
	require.Equal(t, s2, h.sender)

	// OnInit no-op
	require.NoError(t, h.OnInit(&simsdkrpc.PluginInit{ComponentId: "comp-2"}))

	// OnSimMessage no-op
	resp, err := h.OnSimMessage(&simsdk.SimMessage{MessageID: "m1"})
	require.NoError(t, err)
	require.Nil(t, resp)

	// OnShutdown no panic
	h.OnShutdown("bye")
}
