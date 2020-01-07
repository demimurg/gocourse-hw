package midware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type handlerF func(consumer, method string) error

func (m *middleware) interceptor(ctx context.Context, method string) error {
	var consumer string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md.Get("consumer")) > 0 {
		consumer = md.Get("consumer")[0]
	}

	for _, f := range []handlerF{
		m.acl.check, m.log.share,
	} {
		if err := f(consumer, method); err != nil {
			return err
		}
	}

	return nil
}

func (m *middleware) UnaryRPC(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if e := m.interceptor(ctx, info.FullMethod); e != nil {
		return nil, e
	}

	return handler(ctx, req)
}

func (m *middleware) Stream(
	srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	if e := m.interceptor(ss.Context(), info.FullMethod); e != nil {
		return e
	}

	return handler(srv, ss)
}
