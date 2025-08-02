package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockAllocator returns fixed ports in sequence for testing.
type mockAllocator struct {
	nextPort int
	registry *PortRegistry
}

func (m *mockAllocator) Allocate(plugin string) int {
	port := m.nextPort
	m.nextPort++
	if m.registry != nil {
		m.registry.Set(plugin, port)
	}
	return port
}

func TestAllocatorServer_handleAllocate(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectPort     bool
	}{
		{
			name:           "valid plugin",
			query:          "?plugin=test-plugin",
			expectedStatus: http.StatusOK,
			expectPort:     true,
		},
		{
			name:           "missing plugin",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectPort:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewPortRegistry()
			mock := &mockAllocator{nextPort: 9200, registry: registry}
			logger := testLogger()
			server := NewAllocatorServer(0, mock, registry, logger)

			req := httptest.NewRequest(http.MethodGet, "/allocate"+tt.query, nil)
			w := httptest.NewRecorder()

			server.Mux.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("unexpected status code: got %d, want %d", res.StatusCode, tt.expectedStatus)
			}

			if tt.expectPort {
				var body AllocateResponse
				err := json.NewDecoder(res.Body).Decode(&body)
				if err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if body.Port != 9200 {
					t.Errorf("unexpected port: got %d, want %d", body.Port, 9200)
				}
			} else {
				buf := new(strings.Builder)
				_, _ = io.Copy(buf, res.Body)
				if !strings.Contains(buf.String(), "plugin") {
					t.Errorf("expected error mentioning 'plugin' but got: %s", buf.String())
				}
			}
		})
	}
}

func TestAllocatorServer_handleAllocatedPorts(t *testing.T) {
	registry := NewPortRegistry()
	registry.Set("plugin-one", 9100)
	registry.Set("plugin-two", 9101)

	mock := &mockAllocator{registry: registry}
	logger := testLogger()
	server := NewAllocatorServer(0, mock, registry, logger)

	req := httptest.NewRequest(http.MethodGet, "/allocated_ports", nil)
	w := httptest.NewRecorder()

	server.Mux.ServeHTTP(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", res.StatusCode)
	}

	var result map[string]int
	err := json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if result["plugin-one"] != 9100 || result["plugin-two"] != 9101 {
		t.Errorf("unexpected result map: %+v", result)
	}
}

func testLogger() *log.Logger {
	return log.New(&strings.Builder{}, "", 0) // discard output
}
