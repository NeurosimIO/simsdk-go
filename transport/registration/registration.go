// Package registration provides shared utilities for transport plugin startup,
// including dynamic port allocation, optional Consul registration, and core registration.
package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Registrar is implemented by anything that can register a service (e.g., Consul registrar plugin).
type Registrar interface {
	Register(ctx context.Context, cfg RegistrationConfig)
}

// RegistrationConfig holds options for plugin registration.
type RegistrationConfig struct {
	PluginName     string // Logical name of the plugin (used for Consul service)
	PluginType     string // Optional: categorization or logging
	CoreAPIBaseURL string // e.g., http://core:8080
	EnableConsul   bool   // Whether to attempt Consul registration
	ServiceAddress string // IP or hostname the plugin should advertise to core
}

// PortAssignment represents the response from the core /allocate endpoint.
type PortAssignment struct {
	Port int    `json:"port"`
	IP   string `json:"ip"`
}

// HTTPClient is an interface for mocking HTTP client behavior.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TransportInitializer sets up plugin startup, including dynamic port allocation and optional Consul/core registration.
type TransportInitializer struct {
	Config     RegistrationConfig
	HTTPClient HTTPClient
	Logger     *log.Logger
	Registrar  Registrar // Optional: for Consul or other service registration
}

// Init retrieves a port and IP from core, optionally registers with Consul, and informs core of the final plugin address.
func (t *TransportInitializer) Init(ctx context.Context) (int, error) {
	if t.Config.EnableConsul && t.Registrar != nil && t.Config.ServiceAddress == "" {
		return 0, fmt.Errorf("registration: ServiceAddress is required for Consul registration")
	}

	assignment, err := t.fetchPortFromCore(ctx)
	if err != nil {
		return 0, err
	}

	if t.Config.EnableConsul && t.Registrar != nil {
		go t.Registrar.Register(ctx, RegistrationConfig{
			PluginName:     t.Config.PluginName,
			ServiceAddress: t.Config.ServiceAddress,
			PluginType:     t.Config.PluginType,
			CoreAPIBaseURL: t.Config.CoreAPIBaseURL,
		})
		t.Logger.Printf("[registration] 🧭 Consul registration started for %s at %s:%d",
			t.Config.PluginName, t.Config.ServiceAddress, assignment.Port)
	} else {
		t.Logger.Printf("[registration] ℹ️ Consul registration disabled for %s", t.Config.PluginName)
	}

	ipToUse := t.Config.ServiceAddress
	if ipToUse == "" {
		ipToUse = assignment.IP
	}

	if err := t.registerWithCore(ctx, ipToUse, assignment.Port); err != nil {
		t.Logger.Printf("[registration] ❌ Failed to register with core: %v", err)
	}

	return assignment.Port, nil
}

func (t *TransportInitializer) fetchPortFromCore(ctx context.Context) (PortAssignment, error) {
	url := fmt.Sprintf("%s/allocate?plugin=%s", t.Config.CoreAPIBaseURL, t.Config.PluginName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PortAssignment{}, fmt.Errorf("failed to create request to core: %w", err)
	}

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return PortAssignment{}, fmt.Errorf("failed to contact core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PortAssignment{}, fmt.Errorf("unexpected response from core: %s", resp.Status)
	}

	var result PortAssignment
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return PortAssignment{}, fmt.Errorf("failed to decode response from core: %w", err)
	}

	t.Logger.Printf("[registration] 📦 Received port %d and IP %s from core for plugin %s",
		result.Port, result.IP, t.Config.PluginName)
	return result, nil
}

func (t *TransportInitializer) registerWithCore(ctx context.Context, ip string, port int) error {
	url := fmt.Sprintf("%s/register", t.Config.CoreAPIBaseURL)

	payload := map[string]any{
		"plugin": t.Config.PluginName,
		"type":   t.Config.PluginType,
		"ip":     ip,
		"port":   port,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal registration body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register plugin with core: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from core register: %s", resp.Status)
	}

	t.Logger.Printf("[registration] 📡 Plugin %s registered with core at %s:%d",
		t.Config.PluginName, ip, port)
	return nil
}
