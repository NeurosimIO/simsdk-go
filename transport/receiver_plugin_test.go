package transport

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/assert"
)

// --- Mocks for testing ---

type mockReceiver struct {
	Inbound chan simsdk.SimMessage
}

func (m *mockReceiver) Start(_ context.Context) error { return nil }
func (m *mockReceiver) Stop(_ context.Context) error  { return nil }
func (m *mockReceiver) GetInboundChan() <-chan simsdk.SimMessage {
	return m.Inbound
}

type mockStreamSender struct {
	ID   string
	Sent []simsdk.SimMessage
	MU   sync.Mutex
	WG   *sync.WaitGroup
}

func (m *mockStreamSender) Send(msg *simsdk.SimMessage) error {
	m.MU.Lock()
	defer m.MU.Unlock()
	m.Sent = append(m.Sent, *msg)
	if m.WG != nil {
		m.WG.Done()
	}
	return nil
}
func (m *mockStreamSender) ComponentID() string { return m.ID }

type mockPerInstanceHandler struct {
	stream simsdk.StreamSender
}

func (h *mockPerInstanceHandler) OnSimMessage(msg *simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}
func (h *mockPerInstanceHandler) OnInit(_ *simsdkrpc.PluginInit) error { return nil }
func (h *mockPerInstanceHandler) OnShutdown(_ string)                  {}
func (h *mockPerInstanceHandler) SetStreamSender(sender simsdk.StreamSender) {
	h.stream = sender
}

// --- Tests ---

func TestBaseReceiverPlugin(t *testing.T) {
	componentID := "receiver-forward"

	receiver := &mockReceiver{Inbound: make(chan simsdk.SimMessage, 1)}
	mockFactory := func(cfg simsdk.CreateComponentRequest) TransportReceiver { return receiver }

	var wg sync.WaitGroup
	wg.Add(1)

	mockSender := &mockStreamSender{
		ID: componentID,
		WG: &wg,
	}

	plugin := NewBaseReceiverPlugin(mockFactory, func() StreamHandlerWithSender {
		return &mockPerInstanceHandler{}
	})

	// Create component before OnInit so receiver is not nil
	err := plugin.CreateComponentInstance(simsdk.CreateComponentRequest{ComponentID: componentID})
	assert.NoError(t, err)

	// Simulate SetStreamSender before init
	plugin.SetStreamSender(mockSender)

	// Call OnInit
	err = plugin.OnInit(&simsdkrpc.PluginInit{ComponentId: componentID})
	assert.NoError(t, err)

	// Send a message to receiver's inbound channel
	receiver.Inbound <- simsdk.SimMessage{
		MessageID:   "msg-123",
		ComponentID: componentID,
		MessageType: "test.forward",
	}

	// Wait for message to be sent to core
	waitAndAssertSent(t, &wg, mockSender, "msg-123")
}

func TestBaseReceiverPlugin_StreamSenderBeforeInit(t *testing.T) {
	componentID := "receiver-late-init"

	receiver := &mockReceiver{Inbound: make(chan simsdk.SimMessage, 1)}
	mockFactory := func(cfg simsdk.CreateComponentRequest) TransportReceiver { return receiver }

	var wg sync.WaitGroup
	wg.Add(1)

	mockSender := &mockStreamSender{
		ID: componentID,
		WG: &wg,
	}

	plugin := NewBaseReceiverPlugin(mockFactory, func() StreamHandlerWithSender {
		return &mockPerInstanceHandler{}
	})

	// Create component before OnInit
	err := plugin.CreateComponentInstance(simsdk.CreateComponentRequest{ComponentID: componentID})
	assert.NoError(t, err)

	// SetStreamSender first
	plugin.SetStreamSender(mockSender)

	// Then init
	err = plugin.OnInit(&simsdkrpc.PluginInit{ComponentId: componentID})
	assert.NoError(t, err)

	// Send message
	receiver.Inbound <- simsdk.SimMessage{
		MessageID:   "msg-early",
		ComponentID: componentID,
		MessageType: "test.forward",
	}

	waitAndAssertSent(t, &wg, mockSender, "msg-early")
}

func waitAndAssertSent(t *testing.T, wg *sync.WaitGroup, sender *mockStreamSender, expectedID string) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		sender.MU.Lock()
		defer sender.MU.Unlock()
		assert.Len(t, sender.Sent, 1)
		assert.Equal(t, expectedID, sender.Sent[0].MessageID)
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for message to be sent to core stream (%s)", expectedID)
	}
}
