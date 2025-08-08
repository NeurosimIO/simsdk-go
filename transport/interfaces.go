// Package transport provides interfaces and factories for message transport mechanisms.
package transport

import (
	"context"

	"github.com/neurosimio/simsdk-go"
)

// TransportSender defines the generic interface for sending messages.
type TransportSender interface {
	Start(ctx context.Context) error
	Send(ctx context.Context, payload []byte, messageType string) error
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
