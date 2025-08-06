package transport

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

// StreamHandlerFactory creates a new handler that supports both StreamHandler and StreamSenderSetter.
type StreamHandlerFactory func() StreamHandlerWithSender

// BaseReceiverPlugin is a reusable base for implementing a receiver transport plugin.
// It manages lifecycle, stream forwarding, and handler wiring without being tied to a specific transport.
type BaseReceiverPlugin struct {
	receivers      map[string]TransportReceiver
	factory        ReceiverFactory
	handlerFactory StreamHandlerFactory
	handlers       map[string]StreamHandlerWithSender
	streamOnce     map[string]*sync.Once
	coreStreams    map[string]simsdk.StreamSender
	mu             sync.RWMutex
}

// NewBaseReceiverPlugin creates a new BaseReceiverPlugin.
func NewBaseReceiverPlugin(factory ReceiverFactory, handlerFactory StreamHandlerFactory) *BaseReceiverPlugin {
	return &BaseReceiverPlugin{
		receivers:      make(map[string]TransportReceiver),
		factory:        factory,
		handlerFactory: handlerFactory,
		handlers:       make(map[string]StreamHandlerWithSender),
		streamOnce:     make(map[string]*sync.Once),
		coreStreams:    make(map[string]simsdk.StreamSender),
	}
}

// --- simsdk.Plugin interface ---

func (p *BaseReceiverPlugin) GetManifest() simsdk.Manifest {
	// Transport-specific plugin should override this to provide correct metadata
	return simsdk.Manifest{
		Name: "base-receiver",
		ComponentTypes: []simsdk.ComponentType{
			{
				ID:                        "base-receiver",
				DisplayName:               "Base Receiver",
				Description:               "Receives messages from a transport and forwards to core.",
				SupportsMultipleInstances: true,
			},
		},
	}
}

// CreateComponentInstance creates and starts a new receiver instance.
func (p *BaseReceiverPlugin) CreateComponentInstance(req simsdk.CreateComponentRequest) error {
	receiver := p.factory(req)
	if receiver == nil {
		return fmt.Errorf("receiver factory returned nil for component %s", req.ComponentID)
	}

	if err := receiver.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start receiver for component %s: %w", req.ComponentID, err)
	}

	p.mu.Lock()
	p.receivers[req.ComponentID] = receiver
	p.mu.Unlock()
	return nil
}

// DestroyComponentInstance stops and removes a receiver instance.
func (p *BaseReceiverPlugin) DestroyComponentInstance(componentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	r, ok := p.receivers[componentID]
	if !ok {
		return nil
	}
	delete(p.receivers, componentID)
	return r.Stop(context.Background())
}

func (p *BaseReceiverPlugin) HandleMessage(msg simsdk.SimMessage) ([]simsdk.SimMessage, error) {
	p.mu.RLock()
	_, ok := p.receivers[msg.ComponentID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown component ID: %s", msg.ComponentID)
	}

	// Receivers typically don't handle direct inbound messages from core
	return nil, nil
}

// --- simsdk.StreamHandler interface ---

func (p *BaseReceiverPlugin) OnInit(init *simsdkrpc.PluginInit) error {
	id := init.ComponentId

	p.mu.Lock()

	// Ensure a receiver exists for this component
	if _, ok := p.receivers[id]; !ok {
		r := p.factory(simsdk.CreateComponentRequest{ComponentID: id})
		if r == nil {
			p.mu.Unlock()
			return fmt.Errorf("receiver factory returned nil for %s", id)
		}
		if err := r.Start(context.Background()); err != nil {
			p.mu.Unlock()
			return fmt.Errorf("failed to start receiver for %s: %w", id, err)
		}
		p.receivers[id] = r
	}

	// Create streamOnce for readiness if not already created
	if _, ok := p.streamOnce[id]; !ok {
		p.streamOnce[id] = &sync.Once{}
	}
	once := p.streamOnce[id]

	receiver := p.receivers[id]

	// Apply cached StreamSender if available
	if sender, ok := p.coreStreams[id]; ok {
		if handler, ok := p.handlers[id]; ok {
			handler.SetStreamSender(sender)
			log.Printf("✅ [%s] SetStreamSender successfully installed on handler", id)
		}
	}

	p.mu.Unlock()

	// Start inbound message forwarder exactly once
	once.Do(func() {
		go func(compID string, r TransportReceiver) {
			log.Printf("🚦 [%s] Starting message forward loop", compID)
			for msg := range r.GetInboundChan() {
				log.Printf("📨 [%s] Received message from transport: %s", compID, msg.MessageID)

				p.mu.RLock()
				stream := p.coreStreams[compID]
				p.mu.RUnlock()

				if stream == nil {
					log.Printf("❗ [%s] No stream found; cannot send message %s", compID, msg.MessageID)
					continue
				}

				log.Printf("📤 [%s] Sending message %s to core via stream", compID, msg.MessageID)

				if err := stream.Send(&msg); err != nil {
					log.Printf("❌ [%s] failed to stream to core: %v", compID, err)
				}
			}
			log.Printf("⛔ [%s] Inbound channel closed", compID)
		}(id, receiver)
	})

	return nil
}

func (p *BaseReceiverPlugin) OnSimMessage(msg *simsdk.SimMessage) ([]*simsdk.SimMessage, error) {
	// Receivers don't handle direct messages from core
	return nil, nil
}

func (p *BaseReceiverPlugin) OnShutdown(reason string) {
	log.Printf("🔻 Receiver plugin shutting down: %s", reason)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, r := range p.receivers {
		_ = r.Stop(context.Background())
	}
}

// SetStreamSender caches or applies a stream sender for a component.
func (p *BaseReceiverPlugin) SetStreamSender(sender simsdk.StreamSender) {
	compID := sender.ComponentID()

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.coreStreams[compID]; !ok {
		p.coreStreams[compID] = sender
		log.Printf("✅ [%s] coreStreams[%s] set in SetStreamSender", compID, compID)
	} else {
		log.Printf("⚠️ [%s] coreStreams[%s] already set", compID, compID)
	}

	if handler, ok := p.handlers[compID]; ok {
		handler.SetStreamSender(sender)
		log.Printf("✅ [%s] SetStreamSender successfully installed on handler", compID)
	} else {
		log.Printf("⚠️ No handler found for %s during SetStreamSender — caching for later", compID)
	}
}

// GetStreamHandler returns the plugin itself as a stream handler.
func (p *BaseReceiverPlugin) GetStreamHandler() simsdk.StreamHandler {
	return p
}

// GetHandler retrieves or creates a StreamHandlerWithSender for a component ID.
func (p *BaseReceiverPlugin) GetHandler(componentID string) StreamHandlerWithSender {
	p.mu.Lock()
	defer p.mu.Unlock()

	if handler, ok := p.handlers[componentID]; ok {
		return handler
	}

	handler := p.handlerFactory()
	p.handlers[componentID] = handler
	return handler
}

var (
	_ simsdk.PluginWithHandlers = (*BaseReceiverPlugin)(nil)
	_ simsdk.StreamSenderSetter = (*BaseReceiverPlugin)(nil)
)
