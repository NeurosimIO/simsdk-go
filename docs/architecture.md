# NeuroSim Plugin Architecture

This document describes how simulation plugins interact with the NeuroSim simulation engine using the `simsdk`.

---

## 🧠 Design Goals

- Support for dynamic simulation extension through decoupled plugin binaries
- Message-type registration and schema propagation
- Component lifecycle control via RPC
- Clear contract between simulator and plugin

---

## 🔌 Plugin Roles

Each plugin is responsible for declaring its capabilities via a `Manifest`, which includes:

- MessageTypes: data payload types the plugin can emit or consume
- ComponentTypes: entities that may be instantiated by the simulation engine
- ControlFunctions: actions callable by the simulation engine or UI
- TransportTypes: communication mechanisms supported by this plugin

Plugins may also implement runtime behavior by handling messages and managing component state.

---

## 📡 gRPC Communication

Plugins run as independent binaries and expose a gRPC service implementing the `PluginService` interface defined in `plugin.proto`.

```
+-------------------+         gRPC        +---------------------+
|  Simulation Core  | <----------------> |   Plugin Binary     |
+-------------------+                    +---------------------+
```

The core engine will:

1. Connect to the plugin via gRPC
2. Call `GetManifest()` to retrieve metadata
3. Use `CreateComponentInstance()` as needed
4. Deliver simulation messages via `HandleMessage()`
5. Optionally call `DestroyComponentInstance()`

---

## 🧩 SDK Interface

Plugins implement one of the following Go interfaces:

### Basic Plugin

```go
type Plugin interface {
    Manifest() Manifest
}
```

### Advanced Plugin (with handlers)

```go
type PluginWithHandlers interface {
    Plugin
    CreateComponentInstance(req CreateComponentRequest) error
    DestroyComponentInstance(componentID string) error
    HandleMessage(msg SimMessage) ([]SimMessage, error)
}
```

The `Manifest` is a structured declaration used by the simulator to understand what the plugin offers.

---

## 🔄 Message Lifecycle

- Messages are defined using `MessageType` and `FieldSpec`
- They are sent as serialized bytes with metadata
- The plugin is responsible for encoding/decoding and routing internally

---

## 🗂️ File Structure

| File | Purpose |
|------|---------|
| `plugin.go` | Defines the core plugin interface |
| `converter.go` | Handles proto ↔ Go type mapping |
| `proto/plugin.proto` | gRPC interface definition |
| `rpc/simsdkrpc/` | Generated gRPC code |
| `README.md` | SDK overview |
| `architecture.md` | This document |

---

## 📦 Packaging Plugins

Plugins are compiled as standard Go binaries and must be available to the simulation engine via known paths. They may be discovered dynamically via configuration or pre-registration.

---

## 🔐 Security & Isolation

Each plugin runs in its own process, providing strong isolation. This enables:

- Crashing plugins to be restarted independently
- Parallel execution and sandboxing
- Separation of trusted and untrusted logic

---

## 🧪 Example Plugin

See [`sample-plugin`](https://github.com/neurosimio/sample-plugin) for a working implementation that defines message types, component handlers, and control functions.
