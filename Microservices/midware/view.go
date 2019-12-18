package midware

import (
	"context"
	"encoding/json"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type middleware struct {
	*sync.Mutex
	acl      map[string][]string // map{consumer: [method1, method2...]}
	sessions []*proto.Event
}

func (m *middleware) GetSess() []*proto.Event {
	return m.sessions
}

func New(ACLdata string) (*middleware, error) {
	m := middleware{}

	m.acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(ACLdata), &m.acl)
	if err != nil {
		return nil, err
	}

	m.sessions = make([]*proto.Event, 0)

	return &m, nil
}

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

func getConsumer(ctx context.Context) string {
	var consumer string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md.Get("consumer")) > 0 {
		consumer = md.Get("consumer")[0]
	}

	return consumer
}
