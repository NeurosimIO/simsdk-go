package transport

import (
	"testing"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/assert"
)

type mockStreamSenderNoOp struct {
	sent bool
}

func (m *mockStreamSenderNoOp) Send(_ *simsdk.SimMessage) error {
	m.sent = true
	return nil
}
func (m *mockStreamSenderNoOp) ComponentID() string { return "mock" }

func TestNoOpStreamHandler_Methods(t *testing.T) {
	h := &NoOpStreamHandler{}
	sender := &mockStreamSenderNoOp{}

	// SetStreamSender should not panic
	h.SetStreamSender(sender)

	// OnInit should return nil
	err := h.OnInit(&simsdkrpc.PluginInit{ComponentId: "comp1"})
	assert.NoError(t, err)

	// OnSimMessage should return nil values
	msgs, err := h.OnSimMessage(&simsdk.SimMessage{})
	assert.NoError(t, err)
	assert.Nil(t, msgs)

	// OnShutdown should not panic
	h.OnShutdown("test-shutdown")
}
