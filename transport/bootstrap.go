package transport

// bootstrap.go
//
// ServePluginWithRegistration starts a gRPC server for a simsdk plugin after performing
// dynamic port allocation and core registration. It also handles SIGINT/SIGTERM for
// graceful shutdown. This is transport-agnostic boilerplate every plugin should reuse.

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/neurosimio/simsdk-go"
	"github.com/neurosimio/simsdk-go/rpc/simsdkrpc"
	"github.com/neurosimio/simsdk-go/transport/registration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func ServePluginWithRegistration(
	ctx context.Context,
	plugin simsdk.PluginWithHandlers,
	cfg registration.RegistrationConfig,
	httpClient registration.HTTPClient,
	logger *log.Logger,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigc; logger.Printf("🛑 shutdown: %s", cfg.PluginName); cancel() }()

	init := &registration.TransportInitializer{
		Config:     cfg,
		HTTPClient: httpClient,
		Logger:     logger,
	}
	port, err := init.Init(ctx)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	grpcServer := grpc.NewServer()
	simsdkrpc.RegisterPluginServiceServer(grpcServer, simsdk.NewGRPCAdapter(plugin))
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("🚀 %s listening on %s", cfg.PluginName, addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Printf("❌ gRPC serve error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	logger.Printf("🛑 stopping gRPC server: %s", cfg.PluginName)
	grpcServer.GracefulStop()
	return nil
}
