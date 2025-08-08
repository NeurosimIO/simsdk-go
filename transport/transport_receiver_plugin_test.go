package transport

import (
	"context"
	"errors"
	"testing"

	"github.com/neurosimio/simsdk-go"
	"github.com/stretchr/testify/require"
)

type mockReceiver struct {
	started  bool
	stopped  bool
	startErr error
	stopErr  error
	ch       chan simsdk.SimMessage
}

func (m *mockReceiver) Start(context.Context) error { m.started = true; return m.startErr }
func (m *mockReceiver) Stop(context.Context) error  { m.stopped = true; return m.stopErr }
func (m *mockReceiver) GetInboundChan() <-chan simsdk.SimMessage {
	if m.ch == nil {
		m.ch = make(chan simsdk.SimMessage)
	}
	return m.ch
}

func TestReceiverPlugin_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		factory         ReceiverFactory
		expectCreateErr bool
	}{
		{
			name: "happy path: start & stop",
			factory: func(req simsdk.CreateComponentRequest) TransportReceiver {
				return &mockReceiver{}
			},
		},
		{
			name: "factory returns nil -> create error",
			factory: func(req simsdk.CreateComponentRequest) TransportReceiver {
				return nil
			},
			expectCreateErr: true,
		},
		{
			name: "receiver start fails -> create error",
			factory: func(req simsdk.CreateComponentRequest) TransportReceiver {
				return &mockReceiver{startErr: errors.New("nope")}
			},
			expectCreateErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewReceiverPlugin(
				simsdk.Manifest{TransportTypes: []simsdk.TransportType{{ID: "x"}}},
				tt.factory,
				func() simsdk.StreamHandler { return &DefaultPerInstanceStreamHandler{} },
			)

			ccr := simsdk.CreateComponentRequest{ComponentID: "r1"}
			err := plugin.CreateComponentInstance(ccr)
			if tt.expectCreateErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// receiver plugins ignore HandleMessage; it should be no-op
			out, err := plugin.HandleMessage(simsdk.SimMessage{})
			require.NoError(t, err)
			require.Nil(t, out)

			// destroy should call Stop()
			err = plugin.DestroyComponentInstance("r1")
			require.NoError(t, err)

			if r, ok := getReceiverForTest(plugin, "r1"); ok {
				mr := r.(*mockReceiver)
				require.True(t, mr.started)
				require.True(t, mr.stopped)
			}
		})
	}
}

// getReceiverForTest mirrors getSenderForTest for the receiver concrete type.
func getReceiverForTest(p simsdk.PluginWithHandlers, id string) (TransportReceiver, bool) {
	b, ok := p.(*baseReceiverPlugin)
	if !ok {
		return nil, false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	r, ok := b.instances[id]
	return r, ok
}

func TestReceiverPlugin_StreamHandlerFactory_DefaultVsCustom(t *testing.T) {
	// default
	p1 := NewReceiverPlugin(
		simsdk.Manifest{},
		func(simsdk.CreateComponentRequest) TransportReceiver { return &mockReceiver{} },
		nil,
	)
	h1 := p1.GetStreamHandler()
	_, isDefault := h1.(*DefaultPerInstanceStreamHandler)
	require.True(t, isDefault)

	// custom
	p2 := NewReceiverPlugin(
		simsdk.Manifest{},
		func(simsdk.CreateComponentRequest) TransportReceiver { return &mockReceiver{} },
		func() simsdk.StreamHandler { return &mockStreamHandler{} },
	)
	h2 := p2.GetStreamHandler()
	_, isCustom := h2.(*mockStreamHandler)
	require.True(t, isCustom)
}
