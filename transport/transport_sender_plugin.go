package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neurosimio/simsdk-go"
)

// NewSenderPlugin returns a PluginWithHandlers wrapper for a transport sender.
func NewSenderPlugin(
	manifest simsdk.Manifest,
	factory SenderFactory,
	streamHandlerFactory func() simsdk.StreamHandler,
) simsdk.PluginWithHandlers {
	if streamHandlerFactory == nil {
		streamHandlerFactory = func() simsdk.StreamHandler { return &DefaultPerInstanceStreamHandler{} }
	}
	return &baseSenderPlugin{
		manifest:             manifest,
		factory:              factory,
		streamHandlerFactory: streamHandlerFactory,
		instances:            make(map[string]TransportSender),
		cancels:              make(map[string]context.CancelFunc), // NEW
	}
}

// baseSenderPlugin is the concrete implementation of PluginWithHandlers for senders.
type baseSenderPlugin struct {
	manifest             simsdk.Manifest
	factory              SenderFactory
	streamHandlerFactory func() simsdk.StreamHandler

	mu        sync.RWMutex
	instances map[string]TransportSender

	// NEW
	cancels map[string]context.CancelFunc
}

func (p *baseSenderPlugin) GetManifest() simsdk.Manifest { return p.manifest }

func (p *baseSenderPlugin) CreateComponentInstance(req simsdk.CreateComponentRequest) error {
	s := p.factory(req)
	if s == nil {
		return fmt.Errorf("sender factory returned nil")
	}

	// Create a per-instance context so we can cancel on destroy.
	ctx, cancel := context.WithCancel(context.Background())

	if err := s.Start(ctx); err != nil {
		cancel() // avoid leak
		return err
	}

	p.mu.Lock()
	p.instances[req.ComponentID] = s
	p.cancels[req.ComponentID] = cancel
	p.mu.Unlock()
	return nil
}

func (p *baseSenderPlugin) DestroyComponentInstance(componentID string) error {
	p.mu.Lock()
	s, ok := p.instances[componentID]
	cancel, hasCancel := p.cancels[componentID]
	if ok {
		delete(p.instances, componentID)
	}
	if hasCancel {
		delete(p.cancels, componentID)
	}
	p.mu.Unlock()

	if !ok {
		return nil
	}
	if hasCancel {
		cancel()
	}
	return s.Close(context.Background())
}

// HandleMessage finds the sender instance and forwards the SimMessage.
// Creates a per-call context (with optional timeout) and does not hold locks
// while invoking external code.
func (p *baseSenderPlugin) HandleMessage(msg simsdk.SimMessage) ([]simsdk.SimMessage, error) {
	// read under RLock
	p.mu.RLock()
	s := p.instances[msg.ComponentID]
	p.mu.RUnlock()

	if s == nil {
		return nil, fmt.Errorf("no sender instance for %q", msg.ComponentID)
	}

	// per-call context (adjust timeout to taste or make configurable)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.SendSim(ctx, msg); err != nil {
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
