package simsdk

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type mockPlugin struct {
	manifest     Manifest
	failCreate   bool
	failDestroy  bool
	failHandle   bool
	lastCreateID string
}

func (m *mockPlugin) Manifest() Manifest {
	return m.manifest
}

func (m *mockPlugin) CreateComponentInstance(req CreateComponentRequest) error {
	if m.failCreate {
		return errors.New("create failed")
	}
	m.lastCreateID = req.ComponentID
	return nil
}

func (m *mockPlugin) DestroyComponentInstance(componentID string) error {
	if m.failDestroy {
		return errors.New("destroy failed")
	}
	m.lastCreateID = ""
	return nil
}

func (m *mockPlugin) HandleMessage(msg SimMessage) ([]SimMessage, error) {
	if m.failHandle || msg.MessageType == "fail" {
		return nil, errors.New("handle failed")
	}
	return []SimMessage{
		{
			MessageType: "ACK",
			MessageID:   "ack-" + msg.MessageID,
			ComponentID: msg.ComponentID,
			Payload:     []byte("received"),
		},
	}, nil
}

func TestGRPCAdapter_GetManifest(t *testing.T) {
	tests := []struct {
		name     string
		plugin   PluginWithHandlers
		expected *simsdkrpc.Manifest
	}{
		{
			name: "basic manifest",
			plugin: &mockPlugin{
				manifest: Manifest{
					Name:    "example",
					Version: "1.0",
				},
			},
			expected: &simsdkrpc.Manifest{
				Name:    "example",
				Version: "1.0",
			},
		},
		{
			name: "manifest with messages",
			plugin: &mockPlugin{
				manifest: Manifest{
					Name:    "extended",
					Version: "2.0",
					MessageTypes: []MessageType{
						{ID: "A", DisplayName: "MsgA"},
						{ID: "B", DisplayName: "MsgB"},
					},
				},
			},
			expected: &simsdkrpc.Manifest{
				Name:    "extended",
				Version: "2.0",
				MessageTypes: []*simsdkrpc.MessageType{
					{Id: "A", DisplayName: "MsgA"},
					{Id: "B", DisplayName: "MsgB"},
				},
			},
		},
	}

	for _, tt := range tests {
		adapter := NewGRPCAdapter(tt.plugin)
		got, err := adapter.GetManifest(context.Background(), &simsdkrpc.ManifestRequest{})
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}
		if got.Manifest.Name != tt.expected.Name || got.Manifest.Version != tt.expected.Version {
			t.Errorf("%s: got %v, want %v", tt.name, got.Manifest, tt.expected)
		}
	}
}

func TestGRPCAdapter_CreateComponentInstance(t *testing.T) {
	tests := []struct {
		name        string
		request     *simsdkrpc.CreateComponentRequest
		failCreate  bool
		expectError bool
	}{
		{
			name: "valid",
			request: &simsdkrpc.CreateComponentRequest{
				ComponentType: "locomotive",
				ComponentId:   "loc-123",
			},
		},
		{
			name: "failure",
			request: &simsdkrpc.CreateComponentRequest{
				ComponentType: "bad",
				ComponentId:   "fail-1",
			},
			failCreate:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		plugin := &mockPlugin{failCreate: tt.failCreate}
		adapter := NewGRPCAdapter(plugin)
		_, err := adapter.CreateComponentInstance(context.Background(), tt.request)
		if tt.expectError && err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
		if !tt.expectError && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}
	}
}

func TestGRPCAdapter_DestroyComponentInstance(t *testing.T) {
	tests := []struct {
		name        string
		input       *wrapperspb.StringValue
		failDestroy bool
		expectError bool
	}{
		{
			name:  "valid",
			input: wrapperspb.String("delete-123"),
		},
		{
			name:        "fail destroy",
			input:       wrapperspb.String("bad-id"),
			failDestroy: true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		plugin := &mockPlugin{failDestroy: tt.failDestroy}
		adapter := NewGRPCAdapter(plugin)
		_, err := adapter.DestroyComponentInstance(context.Background(), tt.input)
		if tt.expectError && err == nil {
			t.Errorf("%s: expected error", tt.name)
		}
		if !tt.expectError && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}
	}
}

func TestGRPCAdapter_HandleMessage(t *testing.T) {
	tests := []struct {
		name           string
		input          *simsdkrpc.SimMessage
		expectedOutput []*simsdkrpc.SimMessage
		expectError    bool
	}{
		{
			name: "ack message",
			input: &simsdkrpc.SimMessage{
				MessageType: "Test",
				MessageId:   "msg-1",
				ComponentId: "cmp-1",
				Payload:     []byte("hello"),
				Metadata:    map[string]string{"k": "v"},
			},
			expectedOutput: []*simsdkrpc.SimMessage{
				{
					MessageType: "ACK",
					MessageId:   "ack-msg-1",
					ComponentId: "cmp-1",
					Payload:     []byte("received"),
				},
			},
		},
		{
			name: "handle failure",
			input: &simsdkrpc.SimMessage{
				MessageType: "fail",
				MessageId:   "fail-1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		plugin := &mockPlugin{}
		adapter := NewGRPCAdapter(plugin)
		resp, err := adapter.HandleMessage(context.Background(), tt.input)
		if tt.expectError && err == nil {
			t.Errorf("%s: expected error", tt.name)
		}
		if !tt.expectError && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}
		if !tt.expectError && !reflect.DeepEqual(resp.OutboundMessages, tt.expectedOutput) {
			t.Errorf("%s: unexpected output: got %+v, want %+v", tt.name, resp.OutboundMessages, tt.expectedOutput)
		}
	}
}

func TestFromProtoHelpers(t *testing.T) {
	t.Run("fromProtoCreateComponentRequest", func(t *testing.T) {
		protoReq := &simsdkrpc.CreateComponentRequest{
			ComponentType: "switch",
			ComponentId:   "sw-01",
			Parameters: map[string]string{
				"foo": "bar",
			},
		}
		got := fromProtoCreateComponentRequest(protoReq)
		want := CreateComponentRequest{
			ComponentType: "switch",
			ComponentID:   "sw-01",
			Parameters: map[string]string{
				"foo": "bar",
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("fromProtoCreateComponentRequest: got %+v, want %+v", got, want)
		}
	})

	t.Run("fromProtoSimMessage and toProtoSimMessage", func(t *testing.T) {
		protoMsg := &simsdkrpc.SimMessage{
			MessageType: "example",
			MessageId:   "123",
			ComponentId: "abc",
			Payload:     []byte("test-payload"),
			Metadata: map[string]string{
				"x": "y",
			},
		}
		sdkMsg := fromProtoSimMessage(protoMsg)
		roundTrip := toProtoSimMessage(sdkMsg)

		if !reflect.DeepEqual(protoMsg, roundTrip) {
			t.Errorf("round-trip SimMessage mismatch:\noriginal: %+v\ngot: %+v", protoMsg, roundTrip)
		}
	})
}

func TestFromProtoFieldType(t *testing.T) {
	tests := []struct {
		name     string
		input    simsdkrpc.FieldType
		expected FieldType
	}{
		{"STRING", simsdkrpc.FieldType_STRING, FieldString},
		{"INT", simsdkrpc.FieldType_INT, FieldInt},
		{"UINT", simsdkrpc.FieldType_UINT, FieldUint},
		{"FLOAT", simsdkrpc.FieldType_FLOAT, FieldFloat},
		{"BOOL", simsdkrpc.FieldType_BOOL, FieldBool},
		{"ENUM", simsdkrpc.FieldType_ENUM, FieldEnum},
		{"TIMESTAMP", simsdkrpc.FieldType_TIMESTAMP, FieldTimestamp},
		{"REPEATED", simsdkrpc.FieldType_REPEATED, FieldRepeated},
		{"OBJECT", simsdkrpc.FieldType_OBJECT, FieldObject},
		{"UNSPECIFIED", simsdkrpc.FieldType_FIELD_TYPE_UNSPECIFIED, FieldType("FIELD_TYPE_UNSPECIFIED")},
		{"UNKNOWN", simsdkrpc.FieldType(999), FieldType("FIELD_TYPE_UNSPECIFIED")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fromProtoFieldType(tt.input)
			if got != tt.expected {
				t.Errorf("fromProtoFieldType(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFromProtoFieldSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    *simsdkrpc.FieldSpec
		expected FieldSpec
	}{
		{
			name: "basic scalar",
			input: &simsdkrpc.FieldSpec{
				Name:        "speed",
				Type:        simsdkrpc.FieldType_FLOAT,
				Required:    true,
				Description: "Train speed",
			},
			expected: FieldSpec{
				Name:        "speed",
				Type:        FieldFloat,
				Required:    true,
				Description: "Train speed",
			},
		},
		{
			name: "object with subtype",
			input: &simsdkrpc.FieldSpec{
				Name:        "location",
				Type:        simsdkrpc.FieldType_OBJECT,
				Subtype:     simsdkrpc.FieldType_STRING,
				Description: "Named location",
			},
			expected: FieldSpec{
				Name:        "location",
				Type:        FieldObject,
				Description: "Named location",
				Subtype:     ptr(FieldString),
			},
		},
		{
			name: "enum field",
			input: &simsdkrpc.FieldSpec{
				Name:       "mode",
				Type:       simsdkrpc.FieldType_ENUM,
				EnumValues: []string{"AUTO", "MANUAL"},
			},
			expected: FieldSpec{
				Name:       "mode",
				Type:       FieldEnum,
				EnumValues: []string{"AUTO", "MANUAL"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fromProtoFieldSpec(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("fromProtoFieldSpec() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestFromProtoFieldSpecs(t *testing.T) {
	tests := []struct {
		name     string
		input    []*simsdkrpc.FieldSpec
		expected []FieldSpec
	}{
		{
			name: "multiple fields",
			input: []*simsdkrpc.FieldSpec{
				{Name: "x", Type: simsdkrpc.FieldType_INT},
				{Name: "y", Type: simsdkrpc.FieldType_STRING},
			},
			expected: []FieldSpec{
				{Name: "x", Type: FieldInt},
				{Name: "y", Type: FieldString},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fromProtoFieldSpecs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("fromProtoFieldSpecs() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
