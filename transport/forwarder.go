// Package transport provides ergonomic, reusable helpers for building
// transport plugins on top of the core simsdk-go transport interfaces.
//
// forwarder.go
//
// This file defines the Forwarder type, a lightweight utility for
// relaying inbound messages from a TransportReceiver to the simulator
// core via a StreamSender.
//
// In most transport plugins, receivers expose an inbound channel of
// simsdk.SimMessage values via GetInboundChan(). These messages must be
// sent upstream to core over a gRPC stream. The Forwarder encapsulates
// this loop:
//
//   - Listens for new messages from the receiver's channel.
//   - Optionally applies a transformation function to each message.
//   - Sends the message to the core via the provided StreamSender.
//   - Runs in its own goroutine until the context is cancelled or
//     the channel is closed.
//
// By centralizing this behavior, all transport plugins can share the
// same reliable, idiomatic forwarding logic, avoiding subtle variations
// or missed edge cases (e.g., context cancellation, channel closure).
//
// Typical usage:
//
//	fwd := &transport.Forwarder{Log: logger}
//	fwd.Start(ctx, receiver, streamSender, nil)
//
// This will forward all inbound messages without transformation.
package transport

import (
	"context"
	"log"

	"github.com/neurosimio/simsdk-go"
)

// Forwarder relays messages from a TransportReceiver to core via a StreamSender.
type Forwarder struct {
	Log *log.Logger
}

// Start launches a goroutine that forwards messages until ctx is cancelled or the channel closes.
func (f *Forwarder) Start(
	ctx context.Context,
	rcv TransportReceiver,
	snd simsdk.StreamSender,
	transform func(*simsdk.SimMessage) *simsdk.SimMessage,
) {
	go func() {
		ch := rcv.GetInboundChan()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-ch:
				if !ok {
					return
				}

				// Work with a pointer
				msg := &m

				// Apply transform if provided
				if transform != nil {
					if tm := transform(msg); tm != nil {
						msg = tm
					}
				}

				// Send the final message
				_ = snd.Send(msg)
			}
		}
	}()
}
