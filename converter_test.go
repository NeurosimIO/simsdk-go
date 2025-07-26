package simsdk

import (
	"testing"

	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

func TestToProtoFieldType(t *testing.T) {
	tests := []struct {
		name     string
		input    FieldType
		expected simsdkrpc.FieldType
	}{
		{"String", FieldString, simsdkrpc.FieldType_STRING},
		{"Int", FieldInt, simsdkrpc.FieldType_INT},
		{"Uint", FieldUint, simsdkrpc.FieldType_UINT},
		{"Float", FieldFloat, simsdkrpc.FieldType_FLOAT},
		{"Bool", FieldBool, simsdkrpc.FieldType_BOOL},
		{"Enum", FieldEnum, simsdkrpc.FieldType_ENUM},
		{"Timestamp", FieldTimestamp, simsdkrpc.FieldType_TIMESTAMP},
		{"Repeated", FieldRepeated, simsdkrpc.FieldType_REPEATED},
		{"Object", FieldObject, simsdkrpc.FieldType_OBJECT},
		{"Invalid", FieldType("INVALID"), simsdkrpc.FieldType_FIELD_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toProtoFieldType(tt.input)
			if got != tt.expected {
				t.Errorf("toProtoFieldType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFieldType_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    FieldType
		expected FieldType
	}{
		{"String", FieldString, FieldString},
		{"Int", FieldInt, FieldInt},
		{"Uint", FieldUint, FieldUint},
		{"Float", FieldFloat, FieldFloat},
		{"Bool", FieldBool, FieldBool},
		{"Enum", FieldEnum, FieldEnum},
		{"Timestamp", FieldTimestamp, FieldTimestamp},
		{"Repeated", FieldRepeated, FieldRepeated},
		{"Object", FieldObject, FieldObject},
		{"Unspecified", FieldType("INVALID"), FieldType("FIELD_TYPE_UNSPECIFIED")},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_RoundTrip", func(t *testing.T) {
			proto := toProtoFieldType(tt.input)
			got := fromProtoFieldType(proto)
			if got != tt.expected {
				t.Errorf("roundtrip failed: got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestToProtoManifest_RoundTrip(t *testing.T) {
	original := Manifest{
		Name:    "TestPlugin",
		Version: "1.0",
		MessageTypes: []MessageType{{
			ID:          "msg1",
			DisplayName: "Message One",
			Description: "A test message",
			Fields: []FieldSpec{
				{Name: "field1", Type: FieldString, Required: true},
				{Name: "field2", Type: FieldEnum, EnumValues: []string{"A", "B"}, Repeated: true},
			},
		}},
		ControlFunctions: []ControlFunctionType{{
			ID:          "cf1",
			DisplayName: "Control One",
			Description: "A control function",
			Fields:      []FieldSpec{},
		}},
		ComponentTypes: []ComponentType{{
			ID:                        "cmp1",
			DisplayName:               "Component A",
			Description:               "A test component",
			Internal:                  false,
			SupportsMultipleInstances: true,
		}},
		TransportTypes: []TransportType{{
			ID:          "tcp",
			DisplayName: "TCP",
			Description: "Transport over TCP",
			Internal:    false,
		}},
	}

	got := FromProtoManifest(ToProtoManifest(original))

	if got.Name != original.Name || got.Version != original.Version {
		t.Errorf("basic fields mismatch: got %+v, want %+v", got, original)
	}
	if len(got.MessageTypes) != len(original.MessageTypes) {
		t.Errorf("MessageTypes length mismatch: got %d, want %d", len(got.MessageTypes), len(original.MessageTypes))
	}
	if len(got.ComponentTypes) != len(original.ComponentTypes) {
		t.Errorf("ComponentTypes length mismatch: got %d, want %d", len(got.ComponentTypes), len(original.ComponentTypes))
	}
	if len(got.TransportTypes) != len(original.TransportTypes) {
		t.Errorf("TransportTypes length mismatch: got %d, want %d", len(got.TransportTypes), len(original.TransportTypes))
	}
}

func TestToProtoFieldSpec_ObjectFields(t *testing.T) {
	tests := []struct {
		name     string
		input    FieldSpec
		expected int
	}{
		{
			name: "NestedObject",
			input: FieldSpec{
				Name:         "parent",
				Type:         FieldObject,
				ObjectFields: []FieldSpec{{Name: "child", Type: FieldString}},
			},
			expected: 1,
		},
		{
			name: "EmptyObjectFields",
			input: FieldSpec{
				Name:         "empty",
				Type:         FieldObject,
				ObjectFields: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toProtoFieldSpec(tt.input)
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if got.Type != simsdkrpc.FieldType_OBJECT {
				t.Errorf("expected FieldType_OBJECT, got %v", got.Type)
			}
			if len(got.ObjectFields) != tt.expected {
				t.Errorf("expected %d ObjectFields, got %d", tt.expected, len(got.ObjectFields))
			}
		})
	}
}

func TestToProtoFieldSpecs_Empty(t *testing.T) {
	var input []FieldSpec
	got := toProtoFieldSpecs(input)
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(got))
	}
}

func TestToProtoFieldSpec_RepeatedSubtype(t *testing.T) {
	sub := FieldString
	spec := FieldSpec{
		Name:    "listOfStrings",
		Type:    FieldRepeated,
		Subtype: &sub,
	}
	result := toProtoFieldSpec(spec)

	if result.Type != simsdkrpc.FieldType_REPEATED {
		t.Errorf("expected REPEATED, got %v", result.Type)
	}
	if result.Subtype != simsdkrpc.FieldType_STRING {
		t.Errorf("expected Subtype STRING, got %v", result.Subtype)
	}
}

func TestToProtoControlFunction(t *testing.T) {
	input := ControlFunctionType{
		ID:          "cf-id",
		DisplayName: "Control Function",
		Description: "For testing",
		Fields:      []FieldSpec{{Name: "enabled", Type: FieldBool}},
	}
	got := toProtoControlFunction(input)
	if got.Id != input.ID || got.DisplayName != input.DisplayName {
		t.Errorf("ControlFunction metadata mismatch")
	}
	if len(got.Fields) != 1 || got.Fields[0].Name != "enabled" {
		t.Errorf("Fields not mapped correctly")
	}
}

func TestToProtoMessageType(t *testing.T) {
	input := MessageType{
		ID:          "msg-type",
		DisplayName: "Message Type",
		Description: "test description",
		Fields:      []FieldSpec{{Name: "fieldA", Type: FieldString}},
	}
	got := toProtoMessageType(input)
	if got.Id != input.ID || got.DisplayName != input.DisplayName {
		t.Errorf("MessageType metadata mismatch")
	}
	if got.Fields[0].Name != "fieldA" {
		t.Errorf("Field not mapped correctly")
	}
}

func TestFromProtoComponentTypes(t *testing.T) {
	input := []*simsdkrpc.ComponentType{{
		Id:                        "csx-core",
		DisplayName:               "CSX Core",
		Description:               "Core system component",
		Internal:                  false,
		SupportsMultipleInstances: false,
	}}
	got := fromProtoComponentTypes(input)
	if len(got) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(got))
	}
	if got[0].ID != "csx-core" || got[0].DisplayName != "CSX Core" {
		t.Errorf("ComponentType fields not correctly converted")
	}
}

func TestToProtoTransportTypes(t *testing.T) {
	input := []TransportType{{
		ID:          "amqp",
		DisplayName: "AMQP 1.0",
		Description: "AMQP transport",
		Internal:    false,
	}}
	got := toProtoTransportTypes(input)
	if len(got) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(got))
	}
	if got[0].Id != "amqp" || got[0].DisplayName != "AMQP 1.0" {
		t.Errorf("TransportType fields not correctly converted to proto")
	}
}

func TestFromProtoTransportTypes(t *testing.T) {
	input := []*simsdkrpc.TransportType{{
		Id:          "tcp",
		DisplayName: "TCP",
		Description: "Raw TCP transport",
		Internal:    false,
	}}
	got := fromProtoTransportTypes(input)
	if len(got) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(got))
	}
	if got[0].ID != "tcp" || got[0].DisplayName != "TCP" {
		t.Errorf("TransportType fields not correctly converted")
	}
}

func TestFromProtoTransportTypes_Multiple(t *testing.T) {
	input := []*simsdkrpc.TransportType{
		{Id: "amqp", DisplayName: "AMQP", Description: "AMQP Desc", Internal: false},
		{Id: "kafka", DisplayName: "Kafka", Description: "Kafka Desc", Internal: true},
	}
	got := fromProtoTransportTypes(input)
	if len(got) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(got))
	}
	if got[1].ID != "kafka" || !got[1].Internal {
		t.Errorf("Second transport not correctly converted")
	}
}

func FuzzToProtoFieldType(f *testing.F) {
	f.Add(int32(0))
	f.Add(int32(1))
	f.Fuzz(func(t *testing.T, input int32) {
		_ = toProtoFieldType(FieldType(input)) // Ensure no panic
	})
}

func TestToProtoSimMessageAndBack(t *testing.T) {
	original := &SimMessage{
		MessageType: "core.status.update",
		MessageID:   "msg-001",
		ComponentID: "core-system",
		Payload:     []byte(`{"status":"running"}`),
		Metadata: map[string]string{
			"traceId":  "abc123",
			"priority": "high",
		},
	}

	proto := ToProtoSimMessage(original)
	if proto == nil {
		t.Fatal("ToProtoSimMessage returned nil")
	}
	if proto.MessageType != original.MessageType || proto.ComponentId != original.ComponentID {
		t.Errorf("Proto fields mismatch: %+v", proto)
	}
	if string(proto.Payload) != string(original.Payload) {
		t.Errorf("Payload mismatch: got %s, want %s", proto.Payload, original.Payload)
	}
	if len(proto.Metadata) != len(original.Metadata) {
		t.Errorf("Metadata length mismatch")
	}

	roundTrip := FromProtoSimMessage(proto)
	if roundTrip.MessageID != original.MessageID || roundTrip.ComponentID != original.ComponentID {
		t.Errorf("Round-trip field mismatch: got %+v, want %+v", roundTrip, original)
	}
	if string(roundTrip.Payload) != string(original.Payload) {
		t.Errorf("Round-trip payload mismatch")
	}
	for k, v := range original.Metadata {
		if roundTrip.Metadata[k] != v {
			t.Errorf("Metadata value for %s mismatch: got %s, want %s", k, roundTrip.Metadata[k], v)
		}
	}
}
