![CI](https://github.com/neurosimio/simsdk/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/neurosimio/simsdk)](https://goreportcard.com/report/github.com/neurosimio/simsdk)
[![Go Reference](https://pkg.go.dev/badge/github.com/neurosimio/simsdk.svg)](https://pkg.go.dev/github.com/neurosimio/simsdk)
# simsdk

The `simsdk` provides the interfaces, types, and registration utilities required to build plugins for the NeuroSim Simulation Engine.

---

## ✨ Overview

This SDK allows third-party developers to create simulation plugins that:

- Define custom message types and behaviors
- Handle their own encoding/decoding and validation
- Register new components (devices, systems) and transports
- Participate in a unified simulation timeline managed by the core engine

---

## 📦 Features

- Plugin lifecycle hooks (`Init`, `Shutdown`)
- Message registration and schema support
- Component definition for simulation actors
- Streaming (SSE) and event injection support
- Transport abstraction for real vs simulated connectivity

---

## 🏗️ Plugin Architecture

Each plugin is compiled as an isolated Go binary and communicates with the simulation core via gRPC and a well-defined SDK interface.

```
+----------------+          +------------------+
| sim-core       |  <---->  | your-plugin.bin  |
|                |   gRPC   |                  |
+----------------+          +------------------+
```

---

## 📦 Installation

```bash
go get github.com/neurosimio/simsdk@latest
```

Build proto:

```bash
protoc -I proto proto/plugin.proto   --go_out=rpc/simsdkrpc   --go-grpc_out=rpc/simsdkrpc   --go_opt=paths=source_relative   --go-grpc_opt=paths=source_relative
```

---

## 🔗 See Also

- [Architecture Guide](docs/architecture.md)
- [Sample Plugin](https://github.com/neurosimio/sample-plugin)
