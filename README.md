# sim-sdk

The `sim-sdk` provides the interfaces, types, and registration utilities required to build plugins for the NeuroSim Simulation Engine.

## ✨ Overview

This SDK allows third-party developers to create simulation plugins that:

- Define custom message types and behaviors
- Handle their own encoding/decoding and validation
- Register new components (devices, systems) and transports
- Participate in a unified simulation timeline managed by the core engine

## 📦 Features

- Plugin lifecycle hooks (`Init`, `Shutdown`)
- Message registration and schema support
- Component definition for simulation actors
- Streaming (SSE) and event injection support
- Transport abstraction for real vs simulated connectivity

## 🏗️ Plugin Architecture

Each plugin is compiled as an isolated Go binary and communicates with the simulation core via gRPC and a well-defined SDK interface.

```text
+----------------+          +------------------+
| sim-core       |  <---->  | your-plugin.bin  |
|                |   gRPC   |                  |
+----------------+          +------------------+

```bash
go get github.com/neurosimio/simsdk@latest
```