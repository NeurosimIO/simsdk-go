// Package transport provides abstractions and plugins for sending messages over various transport mechanisms.
package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/neurosimio/simsdk-go"
)

// BaseSenderPlugin implements simsdk.PluginWithHandlers for managing generic transport senders.
type BaseSenderPlugin struct {
	mu      sync.RWMutex
	senders map[string]TransportSender
	factory SenderFactory
	handler simsdk.StreamHandler
}

// NewBaseSenderPlugin creates a new plugin with the given factory and stream handler.
func NewBaseSenderPlugin(factory SenderFactory, handler simsdk.StreamHandler) *BaseSenderPlugin {
	return &BaseSenderPlugin{
		senders: make(map[string]TransportSender),
		factory: factory,
		handler: handler,
	}
}

// CreateComponentInstance creates and starts a new transport sender instance.
func (p *BaseSenderPlugin) CreateComponentInstance(req simsdk.CreateComponentRequest) error {
	sender := p.factory(req)
	if sender == nil {
		return fmt.Errorf("failed to create sender for component %s", req.ComponentID)
	}

	if err := sender.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start sender: %w", err)
	}

	p.mu.Lock()
	p.senders[req.ComponentID] = sender
	p.mu.Unlock()
	return nil
}

// DestroyComponentInstance stops and removes a sender instance.
func (p *BaseSenderPlugin) DestroyComponentInstance(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sender, ok := p.senders[id]
	if !ok {
		return fmt.Errorf("component not found: %s", id)
	}

	if err := sender.Stop(context.Background()); err != nil {
		return fmt.Errorf("failed to stop sender: %w", err)
	}

	delete(p.senders, id)
	return nil
}

// HandleMessage sends the payload through the registered sender.
func (p *BaseSenderPlugin) HandleMessage(msg simsdk.SimMessage) ([]simsdk.SimMessage, error) {
	p.mu.RLock()
	sender, ok := p.senders[msg.ComponentID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown component ID: %s", msg.ComponentID)
	}

	if err := sender.Send(context.Background(), msg.Payload, msg.MessageType); err != nil {
		return nil, fmt.Errorf("send failed: %w", err)
	}

	return nil, nil
}

// GetStreamHandler returns the stream handler for this plugin.
func (p *BaseSenderPlugin) GetStreamHandler() simsdk.StreamHandler {
	return p.handler
}

// GetManifest returns metadata about this sender plugin.
func (p *BaseSenderPlugin) GetManifest() simsdk.Manifest {
	return simsdk.Manifest{
		Name: "base-sender",
		ComponentTypes: []simsdk.ComponentType{
			{
				ID:                        "generic-sender",
				DisplayName:               "Generic Sender",
				Description:               "Sends messages to a transport endpoint.",
				SupportsMultipleInstances: true,
			},
		},
	}
}

var _ simsdk.PluginWithHandlers = (*BaseSenderPlugin)(nil)
