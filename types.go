package simsdk

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
	EnumValues  []string // only for FieldEnum
	Repeated    bool     // true if this field can appear multiple times
	Description string
	Subtype     *FieldType `json:"subtype,omitempty"`
}

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

type TransportType struct {
	ID          string // e.g., "amqp", "kafka", "mqtt"
	DisplayName string // e.g., "AMQP 1.0", "Kafka Stream"
	Description string
	Internal    bool // true = core-managed transport, false = plugin-provided
}
