// Package transport provides interfaces and factories for message transport mechanisms.
package transport

import (
	"context"

	"github.com/neurosimio/simsdk-go"
)

// TransportSender defines the generic interface for sending messages.
// TransportSender sends a complete SimMessage (ID, type, payload, metadata).
type TransportSender interface {
	Start(ctx context.Context) error
	SendSim(ctx context.Context, msg simsdk.SimMessage) error
	Close(ctx context.Context) error
}

// TransportReceiver defines the generic interface for receiving messages.
type TransportReceiver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetInboundChan() <-chan simsdk.SimMessage
}

// SenderFactory creates a TransportSender from a config type.
type SenderFactory func(req simsdk.CreateComponentRequest) TransportSender

// ReceiverFactory creates a TransportReceiver from a config type.
type ReceiverFactory func(req simsdk.CreateComponentRequest) TransportReceiver

// StreamHandlerFactory creates a StreamHandler.
type StreamHandlerFactory func() simsdk.StreamHandler

// FullSender can send the whole message (ID, type, payload, metadata).
// If a TransportSender also implements this, the SDK will prefer it.
type FullSender interface {
	SendSim(ctx context.Context, msg simsdk.SimMessage) error
}
