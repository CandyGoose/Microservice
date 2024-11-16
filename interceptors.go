package main

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func AccessStreamInterceptor(logSubs *EventSubs, stats *StatTracker) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		event := createEvent(ss.Context(), info.FullMethod)
		logSubs.Publish(event)
		stats.Track(event.Method, event.Consumer)

		return handler(srv, ss)
	}
}

func AccessUnaryInterceptor(logSubs *EventSubs, stats *StatTracker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		event := createEvent(ctx, info.FullMethod)
		logSubs.Publish(event)
		stats.Track(event.Method, event.Consumer)

		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(acl map[string][]string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		consumer := extractConsumer(ss.Context())
		if !hasAccess(info.FullMethod, acl[consumer]) {
			return status.Errorf(codes.Unauthenticated, "consumer '%s' does not have access to %s", consumer, info.FullMethod)
		}

		return handler(srv, ss)
	}
}

func AuthUnaryInterceptor(acl map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		consumer := extractConsumer(ctx)
		if !hasAccess(info.FullMethod, acl[consumer]) {
			return nil, status.Errorf(codes.Unauthenticated, "consumer '%s' does not have access to %s", consumer, info.FullMethod)
		}

		return handler(ctx, req)
	}
}

func hasAccess(method string, acl []string) bool {
	for _, m := range acl {
		if method == m || (strings.HasSuffix(m, "*") && strings.HasPrefix(method, m[:len(m)-1])) {
			return true
		}
	}
	return false
}

func createEvent(ctx context.Context, method string) *Event {
	host := ""
	if p, ok := peer.FromContext(ctx); ok {
		host = p.Addr.String()
	}

	return &Event{
		Timestamp: time.Now().Unix(),
		Consumer:  extractConsumer(ctx),
		Method:    method,
		Host:      host,
	}
}

func extractConsumer(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("consumer"); len(v) > 0 {
			return v[0]
		}
	}
	return ""
}
