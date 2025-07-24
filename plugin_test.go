package simsdk

import "testing"

func TestRegisterAndRetrieveManifest(t *testing.T) {
	// Clear global state before the test
	registeredManifests = nil

	m := Manifest{
		Name:    "TestPlugin",
		Version: "1.0",
	}
	RegisterManifest(m)

	all := GetAllRegisteredManifests()
	if len(all) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(all))
	}
	if all[0].Name != "TestPlugin" || all[0].Version != "1.0" {
		t.Errorf("unexpected manifest content: %+v", all[0])
	}
}

func TestMultipleManifestRegistrations(t *testing.T) {
	registeredManifests = nil // reset global state

	RegisterManifest(Manifest{Name: "PluginA", Version: "1.0"})
	RegisterManifest(Manifest{Name: "PluginB", Version: "2.0"})

	all := GetAllRegisteredManifests()
	if len(all) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(all))
	}
	if all[0].Name != "PluginA" || all[1].Name != "PluginB" {
		t.Errorf("unexpected manifest ordering or content: %+v", all)
	}
}

func TestRegisterEmptyManifest(t *testing.T) {
	registeredManifests = nil

	RegisterManifest(Manifest{})
	all := GetAllRegisteredManifests()

	if len(all) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(all))
	}
	if all[0].Name != "" || all[0].Version != "" {
		t.Errorf("expected empty Name/Version, got %+v", all[0])
	}
}

func TestSimMessageStruct(t *testing.T) {
	msg := SimMessage{
		MessageType: "msg.test",
		MessageID:   "123",
		ComponentID: "comp01",
		Payload:     []byte(`{"key":"value"}`),
		Metadata:    map[string]string{"trace": "abc123"},
	}

	if msg.MessageType != "msg.test" || msg.ComponentID != "comp01" {
		t.Errorf("unexpected message content: %+v", msg)
	}
	if len(msg.Metadata) != 1 || msg.Metadata["trace"] != "abc123" {
		t.Errorf("metadata mismatch: %+v", msg.Metadata)
	}
}
