// Package simsdk provides core interfaces and types for building simulation plugins
package simsdk

// Plugin is the main interface all simulation plugins must implement
type Plugin interface {
	Manifest() Manifest
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
