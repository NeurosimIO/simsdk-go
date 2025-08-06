package transport

import (
	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

// NoOpStreamHandler implements StreamHandler but does nothing.
// Useful for senders that don't need to handle inbound streams.
type NoOpStreamHandler struct{}

// SetStreamSender implements simsdk.StreamSenderSetter but does nothing.
func (h *NoOpStreamHandler) SetStreamSender(sender simsdk.StreamSender) {}

// OnInit implements simsdk.StreamHandler but does nothing.
func (h *NoOpStreamHandler) OnInit(_ *simsdkrpc.PluginInit) error { return nil }

// OnSimMessage implements simsdk.StreamHandler but does nothing.
func (h *NoOpStreamHandler) OnSimMessage(_ *simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}

// OnShutdown implements simsdk.StreamHandler but does nothing.
func (h *NoOpStreamHandler) OnShutdown(_ string) {}

var _ simsdk.StreamHandler = (*NoOpStreamHandler)(nil)
