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

// StreamHandlerWithSender is a StreamHandler that can also accept a StreamSender.
type StreamHandlerWithSender interface {
	simsdk.StreamHandler
	simsdk.StreamSenderSetter
}

// SenderFactory creates a TransportSender from a config type.
type SenderFactory func(config any) TransportSender

// ReceiverFactory creates a TransportReceiver from a config type.
type ReceiverFactory func(cfg simsdk.CreateComponentRequest) TransportReceiver
