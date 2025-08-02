package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// AllocateResponse defines the JSON response for port allocations.
type AllocateResponse struct {
	Port int `json:"port"`
}

// RegisterRequest represents a plugin registration request.
type RegisterRequest struct {
	ComponentID   string            `json:"component_id"`
	ComponentType string            `json:"component_type"`
	Parameters    map[string]string `json:"parameters"`
}

// PortRegistry tracks plugin-to-port assignments in memory.
type PortRegistry struct {
	mu    sync.RWMutex
	ports map[string]int
}

func NewPortRegistry() *PortRegistry {
	return &PortRegistry{
		ports: make(map[string]int),
	}
}

func (r *PortRegistry) Set(plugin string, port int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ports[plugin] = port
}

func (r *PortRegistry) GetAll() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copy := make(map[string]int, len(r.ports))
	for k, v := range r.ports {
		copy[k] = v
	}
	return copy
}

// RegisteredPluginRegistry tracks plugin registration metadata.
type RegisteredPluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]RegisterRequest
}

func NewRegisteredPluginRegistry() *RegisteredPluginRegistry {
	return &RegisteredPluginRegistry{
		plugins: make(map[string]RegisterRequest),
	}
}

func (r *RegisteredPluginRegistry) Add(id string, req RegisterRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[id] = req
}

func (r *RegisteredPluginRegistry) GetAll() map[string]RegisterRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copy := make(map[string]RegisterRequest, len(r.plugins))
	for k, v := range r.plugins {
		copy[k] = v
	}
	return copy
}

// Allocator handles dynamic port assignments.
type Allocator interface {
	Allocate(plugin string) int
}

// InMemoryAllocator assigns sequential ports and stores them in a registry.
type InMemoryAllocator struct {
	mu       sync.Mutex
	current  int
	registry *PortRegistry
}

func NewInMemoryAllocator(basePort int, registry *PortRegistry) *InMemoryAllocator {
	return &InMemoryAllocator{
		current:  basePort,
		registry: registry,
	}
}

func (a *InMemoryAllocator) Allocate(plugin string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	port := a.current
	a.current++
	a.registry.Set(plugin, port)
	return port
}

// AllocatorServer wraps the HTTP server and dependencies.
type AllocatorServer struct {
	Port      int
	Allocator Allocator
	Logger    *log.Logger
	Mux       *http.ServeMux
	Registry  *PortRegistry
	PluginReg *RegisteredPluginRegistry
}

// NewAllocatorServer constructs a new server with handlers wired.
func NewAllocatorServer(port int, allocator Allocator, registry *PortRegistry, logger *log.Logger) *AllocatorServer {
	mux := http.NewServeMux()
	pluginReg := NewRegisteredPluginRegistry()

	server := &AllocatorServer{
		Port:      port,
		Allocator: allocator,
		Logger:    logger,
		Mux:       mux,
		Registry:  registry,
		PluginReg: pluginReg,
	}
	mux.HandleFunc("/allocate", server.handleAllocate)
	mux.HandleFunc("/allocated_ports", server.handleAllocatedPorts)
	mux.HandleFunc("/register", server.handleRegister)
	mux.HandleFunc("/registered_plugins", server.handleRegisteredPlugins)

	return server
}

func (s *AllocatorServer) handleAllocate(w http.ResponseWriter, r *http.Request) {
	plugin := r.URL.Query().Get("plugin")
	if plugin == "" {
		http.Error(w, "missing required query parameter: plugin", http.StatusBadRequest)
		return
	}

	port := s.Allocator.Allocate(plugin)
	s.Logger.Printf("🧩 Allocated port %d for plugin %s", port, plugin)

	resp := AllocateResponse{Port: port}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *AllocatorServer) handleAllocatedPorts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.Registry.GetAll())
}

func (s *AllocatorServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.ComponentID) == "" || strings.TrimSpace(req.ComponentType) == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	s.PluginReg.Add(req.ComponentID, req)
	s.Logger.Printf("📝 Registered plugin %s of type %s", req.ComponentID, req.ComponentType)
	w.WriteHeader(http.StatusOK)
}

func (s *AllocatorServer) handleRegisteredPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.PluginReg.GetAll())
}

// Run starts the allocator HTTP server.
func (s *AllocatorServer) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.Port)
	s.Logger.Printf("🚀 Plugin allocator listening on %s", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Mux,
	}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	return srv.ListenAndServe()
}

func main() {
	port := mustEnvInt("ALLOCATOR_PORT", 8080)
	base := mustEnvInt("ALLOCATOR_BASE_PORT", 9100)

	logger := log.New(os.Stdout, "", log.LstdFlags)
	registry := NewPortRegistry()
	allocator := NewInMemoryAllocator(base, registry)
	server := NewAllocatorServer(port, allocator, registry, logger)

	ctx := context.Background()
	if err := server.Run(ctx); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("❌ Failed to start allocator: %v", err)
	}
}

func mustEnvInt(name string, def int) int {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("❌ Invalid value for %s: %v", name, err)
	}
	return i
}
