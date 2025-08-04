package simsdk

// MessageType describes a message that can be used in simulation.
// It does not carry the actual payload, just metadata for configuration.
type MessageType struct {
	ID          string      `json:"id" yaml:"id" xml:"id" protobuf:"bytes,1,opt,name=id" mapstructure:"id"`
	DisplayName string      `json:"displayName" yaml:"displayName" xml:"displayName" protobuf:"bytes,2,opt,name=displayName" mapstructure:"displayName"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty" protobuf:"bytes,3,opt,name=description" mapstructure:"description"`
	Fields      []FieldSpec `json:"fields" yaml:"fields" xml:"fields" protobuf:"bytes,4,rep,name=fields" mapstructure:"fields"`
}

// FieldSpec describes a field that must be filled in to configure a message.
type FieldSpec struct {
	Name         string      `json:"name" yaml:"name" xml:"name" protobuf:"bytes,1,opt,name=name" mapstructure:"name"`
	Type         FieldType   `json:"type" yaml:"type" xml:"type" protobuf:"bytes,2,opt,name=type" mapstructure:"type"`
	Required     bool        `json:"required" yaml:"required" xml:"required" protobuf:"varint,3,opt,name=required" mapstructure:"required"`
	EnumValues   []string    `json:"enumValues,omitempty" yaml:"enumValues,omitempty" xml:"enumValues,omitempty" protobuf:"bytes,4,rep,name=enumValues" mapstructure:"enumValues"`
	Repeated     bool        `json:"repeated,omitempty" yaml:"repeated,omitempty" xml:"repeated,omitempty" protobuf:"varint,5,opt,name=repeated" mapstructure:"repeated"`
	Description  string      `json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty" protobuf:"bytes,6,opt,name=description" mapstructure:"description"`
	Subtype      *FieldType  `json:"subtype,omitempty" yaml:"subtype,omitempty" xml:"subtype,omitempty" protobuf:"bytes,7,opt,name=subtype" mapstructure:"subtype"`
	ObjectFields []FieldSpec `json:"objectFields,omitempty" yaml:"objectFields,omitempty" xml:"objectFields,omitempty" protobuf:"bytes,8,rep,name=objectFields" mapstructure:"objectFields"`
}

// ControlFunctionType describes a non-message block that alters control flow.
type ControlFunctionType struct {
	ID          string      `json:"id" yaml:"id" xml:"id" protobuf:"bytes,1,opt,name=id" mapstructure:"id"`
	DisplayName string      `json:"displayName" yaml:"displayName" xml:"displayName" protobuf:"bytes,2,opt,name=displayName" mapstructure:"displayName"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty" protobuf:"bytes,3,opt,name=description" mapstructure:"description"`
	Fields      []FieldSpec `json:"fields,omitempty" yaml:"fields,omitempty" xml:"fields,omitempty" protobuf:"bytes,4,rep,name=fields" mapstructure:"fields"`
}

// ComponentType describes something that sends or receives messages.
type ComponentType struct {
	ID                        string `json:"id" yaml:"id" xml:"id" protobuf:"bytes,1,opt,name=id" mapstructure:"id"`
	DisplayName               string `json:"displayName" yaml:"displayName" xml:"displayName" protobuf:"bytes,2,opt,name=displayName" mapstructure:"displayName"`
	Internal                  bool   `json:"internal,omitempty" yaml:"internal,omitempty" xml:"internal,omitempty" protobuf:"varint,3,opt,name=internal" mapstructure:"internal"`
	Description               string `json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty" protobuf:"bytes,4,opt,name=description" mapstructure:"description"`
	SupportsMultipleInstances bool   `json:"supportsMultipleInstances,omitempty" yaml:"supportsMultipleInstances,omitempty" xml:"supportsMultipleInstances,omitempty" protobuf:"varint,5,opt,name=supportsMultipleInstances" mapstructure:"supportsMultipleInstances"`
}

// TransportType describes a transport mechanism (e.g., AMQP).
type TransportType struct {
	ID          string `json:"id" yaml:"id" xml:"id" protobuf:"bytes,1,opt,name=id" mapstructure:"id"`
	DisplayName string `json:"displayName" yaml:"displayName" xml:"displayName" protobuf:"bytes,2,opt,name=displayName" mapstructure:"displayName"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty" protobuf:"bytes,3,opt,name=description" mapstructure:"description"`
	Internal    bool   `json:"internal,omitempty" yaml:"internal,omitempty" xml:"internal,omitempty" protobuf:"varint,4,opt,name=internal" mapstructure:"internal"`
}
