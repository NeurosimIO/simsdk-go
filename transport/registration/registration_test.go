package registration

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

// --- Mocks ---

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

type mockRegistrar struct {
	called     bool
	registered []RegistrationConfig
	err        error
	wg         *sync.WaitGroup
}

func (m *mockRegistrar) Register(ctx context.Context, cfg RegistrationConfig) {
	m.called = true
	m.registered = append(m.registered, cfg)
	if m.wg != nil {
		m.wg.Done()
	}
}

// --- Tests ---

func TestTransportInitializer_Init(t *testing.T) {
	tests := []struct {
		name                  string
		config                RegistrationConfig
		httpClientFunc        func(req *http.Request) (*http.Response, error)
		expectPort            int
		expectErr             bool
		expectRegistrarCalled bool
	}{
		{
			name: "successful registration with registrar",
			config: RegistrationConfig{
				PluginName:     "sample-plugin",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   true,
				ServiceAddress: "10.1.1.10",
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				body := io.NopCloser(bytes.NewBufferString(`{"port": 7777, "ip": "10.1.1.10"}`))
				return &http.Response{StatusCode: 200, Body: body}, nil
			},
			expectPort:            7777,
			expectErr:             false,
			expectRegistrarCalled: true,
		},
		{
			name: "successful init without registrar",
			config: RegistrationConfig{
				PluginName:     "standalone-plugin",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   false,
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				body := io.NopCloser(bytes.NewBufferString(`{"port": 8181, "ip": "10.0.0.8"}`))
				return &http.Response{StatusCode: 200, Body: body}, nil
			},
			expectPort:            8181,
			expectErr:             false,
			expectRegistrarCalled: false,
		},
		{
			name: "core returns error",
			config: RegistrationConfig{
				PluginName:     "broken-plugin",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   true,
				ServiceAddress: "10.0.0.2",
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString("fail"))}, nil
			},
			expectErr: true,
		},
		{
			name: "invalid JSON response",
			config: RegistrationConfig{
				PluginName:     "bad-json",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   true,
				ServiceAddress: "10.0.0.2",
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`not-a-json`)),
				}, nil
			},
			expectErr: true,
		},
		{
			name: "network error from core",
			config: RegistrationConfig{
				PluginName:     "offline-core",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   true,
				ServiceAddress: "10.0.0.2",
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("unreachable")
			},
			expectErr: true,
		},
		{
			name: "missing ServiceAddress when consul is enabled",
			config: RegistrationConfig{
				PluginName:     "bad-config",
				CoreAPIBaseURL: "http://core",
				EnableConsul:   true,
				ServiceAddress: "",
			},
			httpClientFunc: func(req *http.Request) (*http.Response, error) {
				body := io.NopCloser(bytes.NewBufferString(`{"port": 7070, "ip": "10.0.0.1"}`))
				return &http.Response{StatusCode: 200, Body: body}, nil
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf := &bytes.Buffer{}
			logger := log.New(logBuf, "", 0)

			mockClient := &mockHTTPClient{doFunc: tt.httpClientFunc}
			var wg sync.WaitGroup

			mockReg := &mockRegistrar{}
			if tt.expectRegistrarCalled {
				wg.Add(1)
				mockReg.wg = &wg
			}

			ti := &TransportInitializer{
				Config:     tt.config,
				HTTPClient: mockClient,
				Logger:     logger,
			}
			if tt.config.EnableConsul {
				ti.Registrar = mockReg
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			port, err := ti.Init(ctx)

			if (err != nil) != tt.expectErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.expectErr)
			}
			if port != tt.expectPort && !tt.expectErr {
				t.Errorf("Init() port = %d, want %d", port, tt.expectPort)
			}

			if tt.expectRegistrarCalled {
				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()
				select {
				case <-done:
					// success
				case <-time.After(100 * time.Millisecond):
					t.Errorf("expected registrar to be called, but it was not (timeout)")
				}
			}
		})
	}
}
