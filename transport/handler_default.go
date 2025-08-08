// Package transport provides ergonomic, reusable helpers for building
// transport plugins on top of the core simsdk-go transport interfaces.
//
// handler_default.go
//
// This file defines DefaultPerInstanceStreamHandler, a minimal
// implementation of the simsdk.StreamHandlerWithSender interface.
//
// In simsdk-go, each plugin component that participates in a gRPC
// MessageStream must provide a StreamHandlerWithSender, which:
//
//   - Receives an injected StreamSender via SetStreamSender().
//   - Handles initialization events (OnInit).
//   - Handles shutdown events (OnShutdown).
//   - Processes incoming SimMessages from core (OnSimMessage).
//
// Many transport plugins have no need to process messages coming *from*
// core, especially pure-receiver plugins. For these cases, developers
// still need to return a valid handler to satisfy the interface.
//
// DefaultPerInstanceStreamHandler provides a safe no-op handler:
//   - SetStreamSender stores the sender for optional later use.
//   - OnInit does nothing and returns nil.
//   - OnShutdown does nothing.
//   - OnSimMessage returns nil, nil (no responses).
//
// This is useful for quickly scaffolding plugins that don't require
// custom stream handling, while still satisfying the interface contract.
package transport

import (
	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

// DefaultPerInstanceStreamHandler implements StreamHandler + StreamSenderSetter with no-op logic.
type DefaultPerInstanceStreamHandler struct {
	sender simsdk.StreamSender
}

func (h *DefaultPerInstanceStreamHandler) SetStreamSender(sender simsdk.StreamSender) {
	h.sender = sender
}

func (h *DefaultPerInstanceStreamHandler) OnInit(_ *simsdkrpc.PluginInit) error { return nil }
func (h *DefaultPerInstanceStreamHandler) OnShutdown(_ string)                  {}
func (h *DefaultPerInstanceStreamHandler) OnSimMessage(_ *simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	return nil, nil
}
