package simsdk

import (
	"context"

	"github.com/neurosimio/simsdk/rpc/simsdkrpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type grpcAdapter struct {
	plugin PluginWithHandlers
	simsdkrpc.UnimplementedPluginServiceServer
}

func NewGRPCAdapter(p PluginWithHandlers) simsdkrpc.PluginServiceServer {
	return &grpcAdapter{plugin: p}
}

func (g *grpcAdapter) GetManifest(ctx context.Context, _ *simsdkrpc.ManifestRequest) (*simsdkrpc.ManifestResponse, error) {
	return &simsdkrpc.ManifestResponse{
		Manifest: toProtoManifest(g.plugin.Manifest()),
	}, nil
}

func (g *grpcAdapter) CreateComponentInstance(ctx context.Context, req *simsdkrpc.CreateComponentRequest) (*simsdkrpc.CreateComponentResponse, error) {
	sdkReq := fromProtoCreateComponentRequest(req)
	err := g.plugin.CreateComponentInstance(sdkReq)
	if err != nil {
		return nil, err
	}
	return &simsdkrpc.CreateComponentResponse{}, nil
}

func (g *grpcAdapter) DestroyComponentInstance(ctx context.Context, id *wrapperspb.StringValue) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, g.plugin.DestroyComponentInstance(id.Value)
}

func (g *grpcAdapter) HandleMessage(ctx context.Context, msg *simsdkrpc.SimMessage) (*simsdkrpc.MessageResponse, error) {
	in := fromProtoSimMessage(msg)
	outbound, err := g.plugin.HandleMessage(in)
	if err != nil {
		return nil, err
	}

	response := &simsdkrpc.MessageResponse{}
	for _, out := range outbound {
		response.OutboundMessages = append(response.OutboundMessages, toProtoSimMessage(out))
	}
	return response, nil
}

// --- helper converters for adapter ---

func fromProtoCreateComponentRequest(req *simsdkrpc.CreateComponentRequest) CreateComponentRequest {
	return CreateComponentRequest{
		ComponentType: req.GetComponentType(),
		ComponentID:   req.GetComponentId(),
		Parameters:    req.GetParameters(),
	}
}

func fromProtoSimMessage(p *simsdkrpc.SimMessage) SimMessage {
	return SimMessage{
		MessageType: p.GetMessageType(),
		MessageID:   p.GetMessageId(),
		ComponentID: p.GetComponentId(),
		Payload:     p.GetPayload(),
		Metadata:    p.GetMetadata(),
	}
}

func toProtoSimMessage(m SimMessage) *simsdkrpc.SimMessage {
	return &simsdkrpc.SimMessage{
		MessageType: m.MessageType,
		MessageId:   m.MessageID,
		ComponentId: m.ComponentID,
		Payload:     m.Payload,
		Metadata:    m.Metadata,
	}
}
