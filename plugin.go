// Package simsdk provides core interfaces and types for building simulation plugins
package simsdk

import "github.com/neurosimio/simsdk-go/rpc/simsdkrpc"

// Plugin is the main interface all simulation plugins must implement
type Plugin interface {
	GetManifest() Manifest
}

// Manifest describes what this plugin provides
type Manifest struct {
	Name             string
	Version          string
	MessageTypes     []MessageType
	ControlFunctions []ControlFunctionType
	ComponentTypes   []ComponentType
	TransportTypes   []TransportType
}

var registeredManifests []Manifest

// RegisterManifest is called by each plugin to register itself.
func RegisterManifest(m Manifest) {
	registeredManifests = append(registeredManifests, m)
}

// GetAllRegisteredManifests returns all plugin manifests registered so far.
func GetAllRegisteredManifests() []Manifest {
	return registeredManifests
}

type SimMessage struct {
	MessageType string
	MessageID   string
	ComponentID string
	Payload     []byte
	Metadata    map[string]string
}
type PluginWithHandlers interface {
	Plugin
	CreateComponentInstance(req CreateComponentRequest) error
	DestroyComponentInstance(componentID string) error
	HandleMessage(msg SimMessage) ([]SimMessage, error)
}
type CreateComponentRequest struct {
	ComponentType string            // The declared ComponentType.ID from the manifest
	ComponentID   string            // Unique instance ID (e.g., "locomotive-001")
	Parameters    map[string]string // Optional plugin-specific initialization parameters
}

func (m Manifest) ToProto() *simsdkrpc.Manifest {
	return ToProtoManifest(m)
}
