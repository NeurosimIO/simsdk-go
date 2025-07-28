package simsdk

// MessageType describes a message that can be used in simulation
// It does not carry the actual payload, just metadata for configuration
type MessageType struct {
	ID          string      `json:"id"`          // e.g., "TrainDeparture"
	DisplayName string      `json:"displayName"` // e.g., "Train Departure"
	Description string      `json:"description"`
	Fields      []FieldSpec `json:"fields"`
}

// FieldSpec describes a field that must be filled in to configure a message
// This supports code generation, validation, and UI construction
type FieldSpec struct {
	Name         string      `json:"name"`
	Type         FieldType   `json:"type"`
	Required     bool        `json:"required"`
	EnumValues   []string    `json:"enumValues,omitempty"` // only for FieldEnum
	Repeated     bool        `json:"repeated,omitempty"`   // true if this field can appear multiple times
	Description  string      `json:"description,omitempty"`
	Subtype      *FieldType  `json:"subtype,omitempty"`
	ObjectFields []FieldSpec `json:"objectFields,omitempty"`
}

// ControlFunctionType describes a non-message block that alters control flow
// Examples: Delay, Repeat, WaitForAck, etc.
type ControlFunctionType struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"displayName"`
	Description string      `json:"description,omitempty"`
	Fields      []FieldSpec `json:"fields,omitempty"`
}

// ComponentType describes something that sends or receives messages
type ComponentType struct {
	ID                        string `json:"id"`
	DisplayName               string `json:"displayName"`
	Internal                  bool   `json:"internal,omitempty"` // true if this is simulated internally
	Description               string `json:"description,omitempty"`
	SupportsMultipleInstances bool   `json:"supportsMultipleInstances,omitempty"`
}

type TransportType struct {
	ID          string `json:"id"`          // e.g., "amqp"
	DisplayName string `json:"displayName"` // e.g., "AMQP 1.0"
	Description string `json:"description,omitempty"`
	Internal    bool   `json:"internal,omitempty"` // core-managed or plugin-provided
}
