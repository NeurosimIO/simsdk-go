package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/neurosimio/simsdk-go"
)

// NewSenderPlugin returns a PluginWithHandlers wrapper for a transport sender.
func NewSenderPlugin(
	manifest simsdk.Manifest,
	factory SenderFactory,
	streamHandlerFactory func() simsdk.StreamHandler, // usually DefaultPerInstanceStreamHandler
) simsdk.PluginWithHandlers {
	return &baseSenderPlugin{
		manifest:             manifest,
		factory:              factory,
		streamHandlerFactory: streamHandlerFactory,
		instances:            make(map[string]TransportSender),
	}
}

// baseSenderPlugin is the concrete implementation of PluginWithHandlers for senders.
type baseSenderPlugin struct {
	manifest             simsdk.Manifest
	factory              SenderFactory
	streamHandlerFactory func() simsdk.StreamHandler

	mu        sync.Mutex
	instances map[string]TransportSender
}

func (p *baseSenderPlugin) GetManifest() simsdk.Manifest { return p.manifest }

func (p *baseSenderPlugin) CreateComponentInstance(req simsdk.CreateComponentRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.instances[req.ComponentID]; exists {
		return nil // idempotent
	}
	s := p.factory(req) // if your SenderFactory still takes `any`, adapt with SenderFactoryFromCCR
	if s == nil {
		return fmt.Errorf("sender factory returned nil")
	}
	p.instances[req.ComponentID] = s
	return s.Start(context.Background())
}

func (p *baseSenderPlugin) DestroyComponentInstance(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	s, ok := p.instances[id]
	if !ok {
		return nil
	}
	_ = s.Close(context.Background())
	delete(p.instances, id)
	return nil
}

func (p *baseSenderPlugin) HandleMessage(msg simsdk.SimMessage) ([]simsdk.SimMessage, error) {
	p.mu.Lock()
	s := p.instances[msg.ComponentID]
	p.mu.Unlock()
	if s == nil {
		return nil, fmt.Errorf("unknown component %q", msg.ComponentID)
	}
	if err := s.Send(context.Background(), msg.Payload, msg.MessageType); err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *baseSenderPlugin) GetStreamHandler() simsdk.StreamHandler {
	if p.streamHandlerFactory != nil {
		return p.streamHandlerFactory()
	}
	return &DefaultPerInstanceStreamHandler{}
}
