// Package simsdk provides core interfaces and types for building simulation plugins
package simsdk

import (
	"io"
	"log"

	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
)

// Plugin is the main interface all simulation plugins must implement
type Plugin interface {
	GetManifest() Manifest
}

// Manifest describes what this plugin provides
type Manifest struct {
	Name                 string                `json:"name"`
	Version              string                `json:"version"`
	MessageTypes         []MessageType         `json:"messageTypes"`
	ControlFunctionTypes []ControlFunctionType `json:"controlFunctionTypes"`
	ComponentTypes       []ComponentType       `json:"componentTypes"`
	TransportTypes       []TransportType       `json:"transportTypes"`
}

var registeredManifests []Manifest

// RegisterManifest is called by each plugin to register itself.
func RegisterManifest(m Manifest) {
	registeredManifests = append(registeredManifests, m)
}

// GetAllRegisteredManifests returns all plugin manifests registered so far.
func GetAllRegisteredManifests() []Manifest {
	return registeredManifests
}

type SimMessage struct {
	MessageType string            `json:"messageType"`
	MessageID   string            `json:"messageId"`
	ComponentID string            `json:"componentId"`
	Payload     []byte            `json:"payload"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type StreamSender interface {
	Send(msg *SimMessage) error
}

type StreamHandler interface {
	OnSimMessage(msg *SimMessage) ([]*SimMessage, error)
	OnInit(init *simsdkrpc.PluginInit) error
	OnShutdown(reason string)
}

type PluginWithHandlers interface {
	Plugin
	CreateComponentInstance(req CreateComponentRequest) error
	DestroyComponentInstance(componentID string) error
	HandleMessage(msg SimMessage) ([]SimMessage, error)
}
type CreateComponentRequest struct {
	ComponentType string            `json:"componentType"`        // Corresponds to ComponentType.ID from manifest
	ComponentID   string            `json:"componentId"`          // e.g., "locomotive-001"
	Parameters    map[string]string `json:"parameters,omitempty"` // Optional plugin-specific init parameters
}

// RegisterRequest is sent by a plugin to register itself with the simulator core.
type RegisterRequest struct {
	Plugin string `json:"plugin"` // Logical plugin name, e.g., "amqp-sender"
	Type   string `json:"type"`   // Optional: transport, system, etc.
	IP     string `json:"ip"`     // Hostname or IP of the plugin's gRPC server
	Port   int    `json:"port"`   // Port of the plugin's gRPC server
}

type RegisteredPlugins map[string]RegisterRequest

func (m Manifest) ToProto() *simsdkrpc.Manifest {
	return ToProtoManifest(m)
}

func ServeStream(handler StreamHandler, stream simsdkrpc.PluginService_MessageStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("🔚 Stream closed by client")
			return nil
		}
		if err != nil {
			log.Printf("❌ Error receiving from stream: %v\n", err)
			return err
		}

		switch msg := in.Content.(type) {
		case *simsdkrpc.PluginMessageEnvelope_Init:
			if err := handler.OnInit(msg.Init); err != nil {
				log.Printf("⚠️ OnInit failed: %v\n", err)
				return err
			}

		case *simsdkrpc.PluginMessageEnvelope_SimMessage:
			sdkMsg := FromProtoSimMessage(msg.SimMessage)
			responses, err := handler.OnSimMessage(sdkMsg)
			if err != nil {
				log.Printf("❌ OnSimMessage failed: %v\n", err)
				_ = stream.Send(&simsdkrpc.PluginMessageEnvelope{
					Content: &simsdkrpc.PluginMessageEnvelope_Nak{
						Nak: &simsdkrpc.PluginNak{
							MessageId:    msg.SimMessage.MessageId,
							ErrorMessage: err.Error(),
						},
					},
				})
				continue
			}

			for _, resp := range responses {
				if err := stream.Send(&simsdkrpc.PluginMessageEnvelope{
					Content: &simsdkrpc.PluginMessageEnvelope_SimMessage{
						SimMessage: ToProtoSimMessage(resp),
					},
				}); err != nil {
					log.Printf("❌ Failed to send SimMessage: %v\n", err)
					return err
				}
			}

			_ = stream.Send(&simsdkrpc.PluginMessageEnvelope{
				Content: &simsdkrpc.PluginMessageEnvelope_Ack{
					Ack: &simsdkrpc.PluginAck{
						MessageId: msg.SimMessage.MessageId,
					},
				},
			})

		case *simsdkrpc.PluginMessageEnvelope_Shutdown:
			handler.OnShutdown(msg.Shutdown.Reason)
			return nil

		default:
			log.Printf("⚠️ Unknown message type in PluginMessageEnvelope: %T\n", msg)
		}
	}
}
