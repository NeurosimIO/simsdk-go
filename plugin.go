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
}

// MessageType describes a message that can be used in simulation
// It does not carry the actual payload, just metadata for configuration
type MessageType struct {
	ID          string // e.g., "TrainDeparture"
	DisplayName string // e.g., "Train Departure"
	Description string
	Fields      []FieldSpec
}

// FieldSpec describes a field that must be filled in to configure a message
// This supports code generation, validation, and UI construction
type FieldSpec struct {
	Name        string
	Type        FieldType
	Required    bool
	Description string
}

type FieldType string

const (
	FieldString FieldType = "string"
	FieldInt    FieldType = "int"
	FieldFloat  FieldType = "float"
	FieldBool   FieldType = "bool"
	FieldEnum   FieldType = "enum"
	FieldTime   FieldType = "time"
)

// ControlFunctionType describes a non-message block that alters control flow
// Examples: Delay, Repeat, WaitForAck, etc.
type ControlFunctionType struct {
	ID          string
	DisplayName string
	Description string
	Fields      []FieldSpec
}

// ComponentType describes something that sends or receives messages
type ComponentType struct {
	ID          string
	DisplayName string
	Internal    bool // true if this is simulated internally
	Description string
}
