package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type middleware struct {
	acl map[string][]string // map{consumer: [method1, method2...]}
}

func (m *middleware) Auth(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		consumer string
		methods  []string
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("consumer")) == 0 {
		return nil, status.Errorf(
			codes.Unauthenticated, "there is no consumer field in ctx",
		)
	}
	consumer = md.Get("consumer")[0]

	var found bool
	for con, meths := range m.acl {
		if con == consumer {
			found = true
			methods = meths
			break
		}
	}
	if !found {
		return nil, status.Errorf(
			codes.Unauthenticated, "consumer doesn't exist",
		)
	}

	var granted bool
	for _, meth := range methods {
		if strings.HasSuffix(meth, "*") || meth == info.FullMethod {
			granted = true
			break
		}
	}
	if !granted {
		return nil, status.Errorf(
			codes.Unauthenticated, "disallowed method",
		)
	}

	return handler(ctx, req)
}

func (m *middleware) StreamAuth(
	srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	return handler(srv, ss)
}

func (m *middleware) ParseACL(data string) error {
	m.acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(data), &m.acl)
	if err != nil {
		return err
	}

	return nil
}

// StartMyMicroservice ...
func StartMyMicroservice(
	ctx context.Context,
	socket, ACLdata string,
) error {
	lis, err := net.Listen("tcp", socket)
	if err != nil {
		return fmt.Errorf(
			"Problem with listener:\nsocket: %s, error: %s",
			socket, err,
		)
	}

	m := middleware{}
	err = m.ParseACL(ACLdata)
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(m.Auth),
		grpc.StreamInterceptor(m.StreamAuth),
		// grpc.StreamInterceptor()
	)

	RegisterBizServer(server, &Biz{})
	RegisterAdminServer(server, &Admin{})

	go func() {
		<-ctx.Done()
		server.GracefulStop()
		// err := lis.Close()
		// if err != nil {
		// 	log.Fatalln("port can't be closed: ", err)
		// }
	}()

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalln("failed to serve: ", err)
		}
	}()

	return nil
}
