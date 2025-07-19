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
