package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func StartMyMicroservice(ctx context.Context, listenAddr string, aclData string) error {
	acl, err := parseacl(aclData)
	if err != nil {
		return fmt.Errorf("unmarshal aclData failed: %w", err)
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listening on %s failed: %w", listenAddr, err)
	}

	subs := NewEventSubs()
	stats := NewStatTracker()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			AccessUnaryInterceptor(subs, stats),
			AuthUnaryInterceptor(acl),
		),
		grpc.ChainStreamInterceptor(
			AccessStreamInterceptor(subs, stats),
			AuthStreamInterceptor(acl),
		),
	)

	RegisterAdminServer(server, NewAdminModule(subs, stats))
	RegisterBizServer(server, NewBizModule())

	errCh := make(chan error)

	go startGRPCServer(server, lis, errCh)
	go handleGracefulShutdown(ctx, server, errCh)

	return nil
}

func parseacl(aclData string) (map[string][]string, error) {
	acl := make(map[string][]string)
	if err := json.Unmarshal([]byte(aclData), &acl); err != nil {
		return nil, err
	}
	return acl, nil
}

func startGRPCServer(server *grpc.Server, lis net.Listener, errCh chan error) {
	if err := server.Serve(lis); err != nil {
		errCh <- fmt.Errorf("grpc server could not serve: %w", err)
	}
}

func handleGracefulShutdown(ctx context.Context, server *grpc.Server, errCh chan error) {
	select {
	case <-ctx.Done():
		server.GracefulStop()
	case err := <-errCh:
		fmt.Println(err)
	}
}
