package consul

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

type mockAgent struct {
	registerCalls   []string
	deregisterCalls []string
	registerErr     error
	deregisterErr   error
	registerFunc    func(service *api.AgentServiceRegistration) error
}

func (m *mockAgent) Register(service *api.AgentServiceRegistration) error {
	m.registerCalls = append(m.registerCalls, service.ID)
	if m.registerFunc != nil {
		return m.registerFunc(service)
	}
	return m.registerErr
}

func (m *mockAgent) Deregister(serviceID string) error {
	m.deregisterCalls = append(m.deregisterCalls, serviceID)
	return m.deregisterErr
}

func TestRegistrar_Register(t *testing.T) {
	tests := []struct {
		name          string
		config        RegistrationConfig
		registerErr   error
		deregisterErr error
		registerDelay time.Duration
		expectLog     string
	}{
		{
			name: "successful registration and deregistration",
			config: RegistrationConfig{
				Name:          "test-service",
				Address:       "127.0.0.1",
				Port:          9999,
				CheckPath:     "/health",
				CheckInterval: 5 * time.Second,
			},
			expectLog: "✅ Registered plugin with Consul",
		},
		{
			name: "missing address",
			config: RegistrationConfig{
				Name:          "test-service",
				Address:       "",
				Port:          9999,
				CheckPath:     "/health",
				CheckInterval: 5 * time.Second,
			},
			expectLog: "❌ Plugin registration failed: Address must be provided",
		},
		{
			name: "registration fails",
			config: RegistrationConfig{
				Name:          "test-service",
				Address:       "127.0.0.1",
				Port:          9999,
				CheckPath:     "/health",
				CheckInterval: 5 * time.Second,
			},
			registerErr: errors.New("consul down"),
			expectLog:   "⚠️ Failed to register with Consul",
		},
		{
			name: "registration times out",
			config: RegistrationConfig{
				Name:          "slow-service",
				Address:       "127.0.0.1",
				Port:          8888,
				CheckPath:     "/health",
				CheckInterval: 5 * time.Second,
			},
			registerDelay: 3 * time.Second, // longer than timeout
			expectLog:     "⏱️ Plugin registration timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf := &bytes.Buffer{}
			logger := log.New(logBuf, "", 0)

			agent := &mockAgent{
				registerErr:   tt.registerErr,
				deregisterErr: tt.deregisterErr,
			}
			if tt.registerDelay > 0 {
				agent.registerFunc = func(service *api.AgentServiceRegistration) error {
					time.Sleep(tt.registerDelay)
					return nil
				}
			}

			registrar := &Registrar{
				Log:      logger,
				Agent:    agent,
				Shutdown: make(chan os.Signal, 1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			registrar.Register(ctx, tt.config)

			time.Sleep(100 * time.Millisecond) // allow logger to flush

			if out := logBuf.String(); !strings.Contains(out, tt.expectLog) {
				t.Errorf("expected log to contain %q, got %q", tt.expectLog, out)
			}
		})
	}
}
