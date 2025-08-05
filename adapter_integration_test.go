package simsdk

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// dummyPlugin is a test plugin for integration testing.
type dummyPlugin struct{}

func (d *dummyPlugin) GetManifest() Manifest {
	return Manifest{
		Name:    "TestManifest",
		Version: "1.0",
	}
}

func (d *dummyPlugin) CreateComponentInstance(req CreateComponentRequest) error {
	return nil
}

func (d *dummyPlugin) DestroyComponentInstance(componentID string) error {
	return nil
}

func (d *dummyPlugin) HandleMessage(msg SimMessage) ([]SimMessage, error) {
	return []SimMessage{
		{
			MessageType: "ReplyType",
			MessageID:   "Reply1",
			ComponentID: "CompReply",
		},
	}, nil
}

func (d *dummyPlugin) GetStreamHandler() StreamHandler {
	return &dummyStreamHandler{}
}

// dummyStreamHandler just records that it was called.
type dummyStreamHandler struct{}

func (h *dummyStreamHandler) OnSimMessage(msg *SimMessage) ([]*SimMessage, error) {
	return []*SimMessage{
		{
			MessageType: "StreamReply",
			MessageID:   "Stream1",
			ComponentID: "CompStream",
		},
	}, nil
}

func (h *dummyStreamHandler) OnInit(init *simsdkrpc.PluginInit) error {
	return nil
}

func (h *dummyStreamHandler) OnShutdown(reason string) {
}

// --- Test ---

func TestGrpcAdapterIntegration(t *testing.T) {
	// Create a listener on a random port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	simsdkrpc.RegisterPluginServiceServer(grpcServer, NewGRPCAdapter(&dummyPlugin{}))

	// Start serving in background
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	// Connect client
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	require.NoError(t, err)
	defer conn.Close()

	client := simsdkrpc.NewPluginServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// --- GetManifest ---
	manifestResp, err := client.GetManifest(ctx, &simsdkrpc.ManifestRequest{})
	require.NoError(t, err)
	require.Equal(t, "TestManifest", manifestResp.Manifest.Name)

	// --- CreateComponentInstance ---
	_, err = client.CreateComponentInstance(ctx, &simsdkrpc.CreateComponentRequest{
		ComponentType: "Type1",
		ComponentId:   "Comp1",
	})
	require.NoError(t, err)

	// --- DestroyComponentInstance ---
	_, err = client.DestroyComponentInstance(ctx, wrapperspb.String("Comp1"))
	require.NoError(t, err)

	// --- HandleMessage ---
	handleResp, err := client.HandleMessage(ctx, &simsdkrpc.SimMessage{
		MessageType: "MsgType",
		MessageId:   "123",
		ComponentId: "CompX",
	})
	require.NoError(t, err)
	require.Len(t, handleResp.OutboundMessages, 1)
	require.Equal(t, "ReplyType", handleResp.OutboundMessages[0].MessageType)

	// --- MessageStream ---
	stream, err := client.MessageStream(ctx)
	require.NoError(t, err)

	// Send Init
	err = stream.Send(&simsdkrpc.PluginMessageEnvelope{
		Content: &simsdkrpc.PluginMessageEnvelope_Init{
			Init: &simsdkrpc.PluginInit{ComponentId: "StreamComp"},
		},
	})
	require.NoError(t, err)

	// Send SimMessage
	err = stream.Send(&simsdkrpc.PluginMessageEnvelope{
		Content: &simsdkrpc.PluginMessageEnvelope_SimMessage{
			SimMessage: &simsdkrpc.SimMessage{
				MessageType: "StreamType",
				MessageId:   "StreamMsg1",
				ComponentId: "StreamComp",
			},
		},
	})
	require.NoError(t, err)

	// Read response
	resp, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, "StreamReply", resp.GetSimMessage().GetMessageType())

	// Send Shutdown
	err = stream.Send(&simsdkrpc.PluginMessageEnvelope{
		Content: &simsdkrpc.PluginMessageEnvelope_Shutdown{
			Shutdown: &simsdkrpc.PluginShutdown{Reason: "Done"},
		},
	})
	require.NoError(t, err)

	// Close stream
	err = stream.CloseSend()
	require.NoError(t, err)
}
