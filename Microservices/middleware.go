package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewMiddleware(ACLdata string) (*middleware, error) {
	m := middleware{}

	m.acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(ACLdata), &m.acl)
	if err != nil {
		return nil, err
	}

	m.sessions = make([]*Event, 0)

	return &m, nil
}

type middleware struct {
	*sync.Mutex
	acl      map[string][]string // map{consumer: [method1, method2...]}
	sessions []*Event
}

func (m *middleware) checkACL(consumer, method string) error {
	var (
		found   bool
		methods []string
	)
	for con, meths := range m.acl {
		if con == consumer {
			found = true
			methods = meths
			break
		}
	}
	if !found {
		return status.Errorf(
			codes.Unauthenticated, "consumer doesn't exist",
		)
	}

	var granted bool
	for _, m := range methods {
		if strings.HasSuffix(m, "*") || m == method {
			granted = true
			break
		}
	}
	if !granted {
		return status.Errorf(
			codes.Unauthenticated, "disallowed method",
		)
	}

	return nil
}

func (m *middleware) authLog(ctx context.Context, method string) error {
	var consumer string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md["consumer"]) > 0 {
		consumer = md.Get("consumer")[0]
	}

	err := m.checkACL(consumer, method)
	if err != nil {
		return err
	}

	// m.Lock()
	m.sessions = append(m.sessions, &Event{
		Consumer: consumer,
		Method:   method,
	})
	// m.Unlock()

	return nil
}

func (m *middleware) UnaryRPC(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if e := m.authLog(ctx, info.FullMethod); e != nil {
		return nil, e
	}

	return handler(ctx, req)
}

func (m *middleware) Stream(
	srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	if e := m.authLog(ss.Context(), info.FullMethod); e != nil {
		return e
	}

	return handler(srv, ss)
}
