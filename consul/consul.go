// Package consul allows plugins to register with Consul using a DI-friendly design.
package consul

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

// RegistrationConfig holds service metadata for Consul registration.
type RegistrationConfig struct {
	Name          string
	Address       string // Required: IP or hostname as seen by core
	Port          int
	CheckPath     string        // e.g., "/health"
	CheckInterval time.Duration // e.g., 10 * time.Second
}

// AgentRegistrar abstracts Consul's Agent API for testability.
type AgentRegistrar interface {
	Register(service *api.AgentServiceRegistration) error
	Deregister(serviceID string) error
}

// DefaultAgentRegistrar wraps the real Consul client agent.
type DefaultAgentRegistrar struct {
	agent *api.Agent
}

func (d *DefaultAgentRegistrar) Register(service *api.AgentServiceRegistration) error {
	return d.agent.ServiceRegister(service)
}
func (d *DefaultAgentRegistrar) Deregister(serviceID string) error {
	return d.agent.ServiceDeregister(serviceID)
}

// Registrar manages Consul registration with injected dependencies.
type Registrar struct {
	Log      *log.Logger
	Agent    AgentRegistrar
	Shutdown chan os.Signal
}

// NewRegistrar constructs a new Registrar with DI-compatible fields.
func NewRegistrar(logger *log.Logger, client *api.Client) *Registrar {
	return &Registrar{
		Log:      logger,
		Agent:    &DefaultAgentRegistrar{agent: client.Agent()},
		Shutdown: make(chan os.Signal, 1),
	}
}

// Register registers the plugin and manages graceful deregistration.
// Register registers the plugin and manages graceful deregistration.
func (r *Registrar) Register(ctx context.Context, cfg RegistrationConfig) {
	if cfg.Address == "" {
		r.Log.Printf("❌ Plugin registration failed: Address must be provided")
		return
	}

	serviceID := fmt.Sprintf("%s-%d", cfg.Name, cfg.Port)
	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Name,
		Address: cfg.Address,
		Port:    cfg.Port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d%s", cfg.Address, cfg.Port, cfg.CheckPath),
			Interval: cfg.CheckInterval.String(),
			Timeout:  "2s",
		},
	}

	// Add timeout to registration process
	regCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		err := r.Agent.Register(reg)
		done <- err
	}()

	select {
	case <-regCtx.Done():
		r.Log.Printf("⏱️ Plugin registration timed out: %v", regCtx.Err())
		return
	case err := <-done:
		if err != nil {
			r.Log.Printf("⚠️ Failed to register with Consul: %v", err)
			return
		}
	}

	r.Log.Printf("✅ Registered plugin with Consul as %s on %s:%d", cfg.Name, cfg.Address, cfg.Port)

	// Handle graceful shutdown
	go func() {
		signal.Notify(r.Shutdown, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-r.Shutdown:
		case <-ctx.Done():
			r.Log.Printf("⏹️ Context canceled before shutdown")
			return
		}

		r.Log.Printf("📦 Deregistering plugin %s from Consul...", serviceID)
		if err := r.Agent.Deregister(serviceID); err != nil {
			r.Log.Printf("❌ Failed to deregister plugin: %v", err)
		} else {
			r.Log.Printf("🧹 Plugin %s deregistered from Consul", serviceID)
		}
		os.Exit(0)
	}()
}
