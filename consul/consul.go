// Package consul package allows plugins to register with Consul sidecar
package consul

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

type RegistrationConfig struct {
	Name          string
	Address       string // Usually the IP of the container/pod
	Port          int
	CheckPath     string // e.g., "/health"
	CheckInterval time.Duration
}

// RegisterWithConsul registers the plugin as a service in Consul,
// adds a health check, and deregisters on shutdown.
func RegisterWithConsul(cfg RegistrationConfig) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("⚠️ Consul client creation failed: %v", err)
		return // fail gracefully
	}

	serviceID := fmt.Sprintf("%s-%d", cfg.Name, cfg.Port)
	address := cfg.Address
	if address == "" {
		address = getLocalIP()
	}

	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Name,
		Address: address,
		Port:    cfg.Port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d%s", address, cfg.Port, cfg.CheckPath),
			Interval: cfg.CheckInterval.String(),
			Timeout:  "2s",
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		log.Printf("⚠️ Failed to register with Consul: %v", err)
		return
	}
	log.Printf("✅ Registered plugin with Consul as %s on %s:%d", cfg.Name, address, cfg.Port)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Printf("📦 Deregistering plugin %s from Consul...", serviceID)
		if err := client.Agent().ServiceDeregister(serviceID); err != nil {
			log.Printf("❌ Failed to deregister plugin: %v", err)
		} else {
			log.Printf("🧹 Plugin %s deregistered from Consul", serviceID)
		}
		os.Exit(0)
	}()
}

func getLocalIP() string {
	// Try to resolve local IP for registration
	conn, err := net.Dial("udp", "1.1.1.1:53")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
