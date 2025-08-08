package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/neurosimio/simsdk-go"
)

// NewReceiverPlugin returns a PluginWithHandlers wrapper for a transport receiver.
func NewReceiverPlugin(
	manifest simsdk.Manifest,
	factory ReceiverFactory,
	streamHandlerFactory func() simsdk.StreamHandler, // typically DefaultPerInstanceStreamHandler
) simsdk.PluginWithHandlers {
	return &baseReceiverPlugin{
		manifest:             manifest,
		factory:              factory,
		streamHandlerFactory: streamHandlerFactory,
		instances:            make(map[string]TransportReceiver),
	}
}

type baseReceiverPlugin struct {
	manifest             simsdk.Manifest
	factory              ReceiverFactory
	streamHandlerFactory func() simsdk.StreamHandler

	mu        sync.Mutex
	instances map[string]TransportReceiver
}

func (p *baseReceiverPlugin) GetManifest() simsdk.Manifest { return p.manifest }

func (p *baseReceiverPlugin) CreateComponentInstance(req simsdk.CreateComponentRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.instances[req.ComponentID]; exists {
		return nil
	}
	r := p.factory(req)
	if r == nil {
		return fmt.Errorf("receiver factory returned nil")
	}
	p.instances[req.ComponentID] = r
	return r.Start(context.Background())
}

func (p *baseReceiverPlugin) DestroyComponentInstance(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	r, ok := p.instances[id]
	if !ok {
		return nil
	}
	_ = r.Stop(context.Background())
	delete(p.instances, id)
	return nil
}

func (p *baseReceiverPlugin) HandleMessage(_ simsdk.SimMessage) ([]simsdk.SimMessage, error) {
	// Receivers don’t handle direct HandleMessage calls
	return nil, nil
}

func (p *baseReceiverPlugin) GetStreamHandler() simsdk.StreamHandler {
	if p.streamHandlerFactory != nil {
		return p.streamHandlerFactory()
	}
	return &DefaultPerInstanceStreamHandler{}
}
