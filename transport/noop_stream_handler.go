package transport

import (
	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

// NoOpStreamHandler implements StreamHandler but does nothing.
// Useful for senders that don't need to handle inbound streams.
type NoOpStreamHandler struct{}

func (h *NoOpStreamHandler) OnInit(*simsdkrpc.PluginInit) error {
	return nil
}

func (h *NoOpStreamHandler) OnSimMessage(*simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}

func (h *NoOpStreamHandler) OnShutdown(reason string) {}

var _ simsdk.StreamHandler = (*NoOpStreamHandler)(nil)
