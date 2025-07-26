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
	Name             string
	Version          string
	MessageTypes     []MessageType
	ControlFunctions []ControlFunctionType
	ComponentTypes   []ComponentType
	TransportTypes   []TransportType
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
	MessageType string
	MessageID   string
	ComponentID string
	Payload     []byte
	Metadata    map[string]string
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
	ComponentType string            // The declared ComponentType.ID from the manifest
	ComponentID   string            // Unique instance ID (e.g., "locomotive-001")
	Parameters    map[string]string // Optional plugin-specific initialization parameters
}

func (m Manifest) ToProto() *simsdkrpc.Manifest {
	return ToProtoManifest(m)
}

func ServeStream(stream simsdkrpc.PluginService_MessageStreamServer, handler StreamHandler) error {
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
