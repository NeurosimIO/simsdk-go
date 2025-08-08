package transport

import (
	"context"
	"errors"
	"testing"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/require"
)

type mockSender struct {
	started   bool
	closed    bool
	sentType  string
	sentBytes []byte
	startErr  error
	sendErr   error
	closeErr  error
}

func (m *mockSender) Start(context.Context) error { m.started = true; return m.startErr }
func (m *mockSender) Send(_ context.Context, payload []byte, messageType string) error {
	m.sentType = messageType
	m.sentBytes = payload
	return m.sendErr
}
func (m *mockSender) Close(context.Context) error { m.closed = true; return m.closeErr }

type mockStreamHandler struct {
	inited     bool
	shutReason string
}

func (h *mockStreamHandler) OnInit(_ *simsdkrpc.PluginInit) error { h.inited = true; return nil }
func (h *mockStreamHandler) OnSimMessage(*simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}
func (h *mockStreamHandler) OnShutdown(reason string) { h.shutReason = reason }

func TestSenderPlugin_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		factory         SenderFactory
		expectCreateErr bool
		doHandle        bool
		handleMsg       simsdk.SimMessage
		expectSendType  string
		expectSendBytes []byte
	}{
		{
			name: "happy path: start, handle, close",
			factory: func(req simsdk.CreateComponentRequest) TransportSender {
				return &mockSender{}
			},
			doHandle:        true,
			handleMsg:       simsdk.SimMessage{ComponentID: "c1", MessageType: "t1", Payload: []byte("x")},
			expectSendType:  "t1",
			expectSendBytes: []byte("x"),
		},
		{
			name: "factory returns nil -> create error",
			factory: func(req simsdk.CreateComponentRequest) TransportSender {
				return nil
			},
			expectCreateErr: true,
		},
		{
			name: "sender.Start fails -> create error",
			factory: func(req simsdk.CreateComponentRequest) TransportSender {
				return &mockSender{startErr: errors.New("boom")}
			},
			expectCreateErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewSenderPlugin(
				simsdk.Manifest{TransportTypes: []simsdk.TransportType{{ID: "x"}}},
				tt.factory,
				func() simsdk.StreamHandler { return &DefaultPerInstanceStreamHandler{} },
			)

			// create
			ccr := simsdk.CreateComponentRequest{ComponentID: "c1"}
			err := plugin.CreateComponentInstance(ccr)
			if tt.expectCreateErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// handle
			if tt.doHandle {
				_, err := plugin.HandleMessage(tt.handleMsg)
				require.NoError(t, err)
			}

			// destroy
			err = plugin.DestroyComponentInstance("c1")
			require.NoError(t, err)

			// Inspect underlying mock if applicable
			if s, ok := getSenderForTest(plugin, "c1"); ok {
				ms := s.(*mockSender)
				// started true if create succeeded
				require.True(t, ms.started)
				// send assertions
				if tt.doHandle {
					require.Equal(t, tt.expectSendType, ms.sentType)
					require.Equal(t, tt.expectSendBytes, ms.sentBytes)
				}
				// closed after destroy
				require.True(t, ms.closed)
			}
		})
	}
}

// getSenderForTest reaches into the concrete plugin to fetch the instance.
// This relies on the concrete type being baseSenderPlugin; adjust if renamed.
func getSenderForTest(p simsdk.PluginWithHandlers, id string) (TransportSender, bool) {
	b, ok := p.(*baseSenderPlugin)
	if !ok {
		return nil, false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	s, ok := b.instances[id]
	return s, ok
}

func TestSenderPlugin_StreamHandlerFactory_DefaultVsCustom(t *testing.T) {
	// default
	p1 := NewSenderPlugin(
		simsdk.Manifest{},
		func(simsdk.CreateComponentRequest) TransportSender { return &mockSender{} },
		nil, // default handler
	)
	h1 := p1.GetStreamHandler()
	_, isDefault := h1.(*DefaultPerInstanceStreamHandler)
	require.True(t, isDefault)

	// custom
	p2 := NewSenderPlugin(
		simsdk.Manifest{},
		func(simsdk.CreateComponentRequest) TransportSender { return &mockSender{} },
		func() simsdk.StreamHandler { return &mockStreamHandler{} },
	)
	h2 := p2.GetStreamHandler()
	_, isCustom := h2.(*mockStreamHandler)
	require.True(t, isCustom)
}
