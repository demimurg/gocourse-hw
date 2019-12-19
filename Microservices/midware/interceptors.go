package midware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

func (m *middleware) UnaryRPC(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		consumer = getConsumer(ctx)
		method   = info.FullMethod
	)

	if e := m.interceptor(consumer, method); e != nil {
		return nil, e
	}

	return handler(ctx, req)
}

func (m *middleware) Stream(
	srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	var (
		consumer = getConsumer(ss.Context())
		method   = info.FullMethod
	)

	if e := m.interceptor(consumer, method); e != nil {
		return e
	}

	return handler(srv, ss)
}

func (m *middleware) interceptor(consumer, method string) error {
	for _, f := range []step{
		m.checkACL, m.addLog,
	} {
		if err := f(consumer, method); err != nil {
			return err
		}
	}

	return nil
}

// it would be better to rename
type step func(consumer, method string) error

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

func (m *middleware) addLog(consumer, method string) error {

	m.logChan <- &proto.Event{
		Consumer: consumer,
		Method:   method,
		Host:     "127.0.0.1:",
	}

	return nil
}

func getConsumer(ctx context.Context) string {
	var consumer string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md.Get("consumer")) > 0 {
		consumer = md.Get("consumer")[0]
	}

	return consumer
}
