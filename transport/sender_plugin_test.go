package transport

import (
	"context"
	"errors"
	"testing"

	"github.com/neurosimio/simsdk-go"
	rpc "github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/assert"
)

type mockSender struct {
	startCalled bool
	sendCalled  bool
	stopCalled  bool
	failStart   bool
	failSend    bool
}

func (m *mockSender) Start(ctx context.Context) error {
	m.startCalled = true
	if m.failStart {
		return errors.New("start failed")
	}
	return nil
}

func (m *mockSender) Send(ctx context.Context, payload []byte, messageType string) error {
	m.sendCalled = true
	if m.failSend {
		return errors.New("send failed")
	}
	return nil
}

func (m *mockSender) Close(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

type mockHandler struct{}

func (m *mockHandler) OnSimMessage(msg *simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}

func (m *mockHandler) OnInit(init *rpc.PluginInit) error {
	return nil
}

func (m *mockHandler) OnShutdown(reason string) {}

func TestBaseSenderPlugin(t *testing.T) {
	tests := []struct {
		name        string
		setupPlugin func() *BaseSenderPlugin
		testFunc    func(t *testing.T, plugin *BaseSenderPlugin)
	}{
		{
			name: "CreateComponentInstance success",
			setupPlugin: func() *BaseSenderPlugin {
				mock := &mockSender{}
				return NewBaseSenderPlugin(func(config any) TransportSender {
					return mock
				}, &mockHandler{})
			},
			testFunc: func(t *testing.T, plugin *BaseSenderPlugin) {
				err := plugin.CreateComponentInstance(simsdk.CreateComponentRequest{
					ComponentID: "comp1",
					Parameters: map[string]string{
						"address": "amqp://localhost",
						"queue":   "test",
					},
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "CreateComponentInstance fails on start",
			setupPlugin: func() *BaseSenderPlugin {
				mock := &mockSender{failStart: true}
				return NewBaseSenderPlugin(func(config any) TransportSender {
					return mock
				}, &mockHandler{})
			},
			testFunc: func(t *testing.T, plugin *BaseSenderPlugin) {
				err := plugin.CreateComponentInstance(simsdk.CreateComponentRequest{
					ComponentID: "comp2",
					Parameters: map[string]string{
						"address": "amqp://localhost",
						"queue":   "test",
					},
				})
				assert.Error(t, err)
			},
		},
		{
			name: "HandleMessage success",
			setupPlugin: func() *BaseSenderPlugin {
				mock := &mockSender{}
				p := NewBaseSenderPlugin(func(config any) TransportSender {
					return mock
				}, &mockHandler{})
				_ = p.CreateComponentInstance(simsdk.CreateComponentRequest{
					ComponentID: "comp3",
					Parameters: map[string]string{
						"address": "amqp://localhost",
						"queue":   "test",
					},
				})
				return p
			},
			testFunc: func(t *testing.T, plugin *BaseSenderPlugin) {
				_, err := plugin.HandleMessage(simsdk.SimMessage{
					ComponentID: "comp3",
					Payload:     []byte("hello"),
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "HandleMessage with unknown component",
			setupPlugin: func() *BaseSenderPlugin {
				mock := &mockSender{}
				return NewBaseSenderPlugin(func(config any) TransportSender {
					return mock
				}, &mockHandler{})
			},
			testFunc: func(t *testing.T, plugin *BaseSenderPlugin) {
				_, err := plugin.HandleMessage(simsdk.SimMessage{
					ComponentID: "unknown",
					Payload:     []byte("msg"),
				})
				assert.Error(t, err)
			},
		},
		{
			name: "DestroyComponentInstance stops sender",
			setupPlugin: func() *BaseSenderPlugin {
				mock := &mockSender{}
				p := NewBaseSenderPlugin(func(config any) TransportSender {
					return mock
				}, &mockHandler{})
				_ = p.CreateComponentInstance(simsdk.CreateComponentRequest{
					ComponentID: "comp4",
					Parameters: map[string]string{
						"address": "amqp://localhost",
						"queue":   "test",
					},
				})
				return p
			},
			testFunc: func(t *testing.T, plugin *BaseSenderPlugin) {
				err := plugin.DestroyComponentInstance("comp4")
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := tt.setupPlugin()
			tt.testFunc(t, plugin)
		})
	}
}
