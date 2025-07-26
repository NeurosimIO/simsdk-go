package simsdk

import (
	"io"
	"testing"

	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestRegisterAndRetrieveManifest(t *testing.T) {
	// Clear global state before the test
	registeredManifests = nil

	m := Manifest{
		Name:    "TestPlugin",
		Version: "1.0",
	}
	RegisterManifest(m)

	all := GetAllRegisteredManifests()
	if len(all) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(all))
	}
	if all[0].Name != "TestPlugin" || all[0].Version != "1.0" {
		t.Errorf("unexpected manifest content: %+v", all[0])
	}
}

func TestMultipleManifestRegistrations(t *testing.T) {
	registeredManifests = nil // reset global state

	RegisterManifest(Manifest{Name: "PluginA", Version: "1.0"})
	RegisterManifest(Manifest{Name: "PluginB", Version: "2.0"})

	all := GetAllRegisteredManifests()
	if len(all) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(all))
	}
	if all[0].Name != "PluginA" || all[1].Name != "PluginB" {
		t.Errorf("unexpected manifest ordering or content: %+v", all)
	}
}

func TestRegisterEmptyManifest(t *testing.T) {
	registeredManifests = nil

	RegisterManifest(Manifest{})
	all := GetAllRegisteredManifests()

	if len(all) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(all))
	}
	if all[0].Name != "" || all[0].Version != "" {
		t.Errorf("expected empty Name/Version, got %+v", all[0])
	}
}

func TestSimMessageStruct(t *testing.T) {
	msg := SimMessage{
		MessageType: "msg.test",
		MessageID:   "123",
		ComponentID: "comp01",
		Payload:     []byte(`{"key":"value"}`),
		Metadata:    map[string]string{"trace": "abc123"},
	}

	if msg.MessageType != "msg.test" || msg.ComponentID != "comp01" {
		t.Errorf("unexpected message content: %+v", msg)
	}
	if len(msg.Metadata) != 1 || msg.Metadata["trace"] != "abc123" {
		t.Errorf("metadata mismatch: %+v", msg.Metadata)
	}
}

type mockStreamHandler struct {
	receivedInit    *simsdkrpc.PluginInit
	receivedMessage *SimMessage
	shutdownReason  string
}

func (m *mockStreamHandler) OnInit(init *simsdkrpc.PluginInit) error {
	m.receivedInit = init
	return nil
}

func (m *mockStreamHandler) OnSimMessage(msg *SimMessage) ([]*SimMessage, error) {
	m.receivedMessage = msg
	return []*SimMessage{
		{
			MessageType: "echo.response",
			MessageID:   "resp-123",
			ComponentID: "mock",
			Payload:     []byte(`{"reply":"ok"}`),
		},
	}, nil
}

func (m *mockStreamHandler) OnShutdown(reason string) {
	m.shutdownReason = reason
}

// mockStream implements simsdkrpc.PluginService_MessageStreamServer
type mockStream struct {
	grpc.ServerStream
	incoming []*simsdkrpc.PluginMessageEnvelope
	sent     []*simsdkrpc.PluginMessageEnvelope
	index    int
}

func (m *mockStream) Recv() (*simsdkrpc.PluginMessageEnvelope, error) {
	if m.index >= len(m.incoming) {
		return nil, io.EOF
	}
	msg := m.incoming[m.index]
	m.index++
	return msg, nil
}

func (m *mockStream) Send(resp *simsdkrpc.PluginMessageEnvelope) error {
	m.sent = append(m.sent, resp)
	return nil
}

func TestServeStream_HandlesInitSimMessageShutdown(t *testing.T) {
	handler := &mockStreamHandler{}
	stream := &mockStream{
		incoming: []*simsdkrpc.PluginMessageEnvelope{
			{
				Content: &simsdkrpc.PluginMessageEnvelope_Init{
					Init: &simsdkrpc.PluginInit{
						ComponentId: "test-component",
					},
				},
			},
			{
				Content: &simsdkrpc.PluginMessageEnvelope_SimMessage{
					SimMessage: &simsdkrpc.SimMessage{
						MessageType: "echo.request",
						MessageId:   "msg-123",
						ComponentId: "test-component",
						Payload:     []byte(`{"test":true}`),
					},
				},
			},
			{
				Content: &simsdkrpc.PluginMessageEnvelope_Shutdown{
					Shutdown: &simsdkrpc.PluginShutdown{Reason: "test-complete"},
				},
			},
		},
	}

	err := ServeStream(stream, handler)
	require.NoError(t, err)

	assert.Equal(t, "test-component", handler.receivedInit.ComponentId)
	assert.NotNil(t, handler.receivedMessage)
	assert.Equal(t, "msg-123", handler.receivedMessage.MessageID)
	assert.Equal(t, "test-complete", handler.shutdownReason)

	// Validate streamed SimMessage response and Ack
	var foundResponse, foundAck bool
	for _, msg := range stream.sent {
		switch m := msg.Content.(type) {
		case *simsdkrpc.PluginMessageEnvelope_SimMessage:
			foundResponse = true
			assert.Equal(t, "echo.response", m.SimMessage.MessageType)
		case *simsdkrpc.PluginMessageEnvelope_Ack:
			foundAck = true
			assert.Equal(t, "msg-123", m.Ack.MessageId)
		}
	}
	assert.True(t, foundResponse, "expected response SimMessage")
	assert.True(t, foundAck, "expected Ack message")
}
