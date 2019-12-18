package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

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

	m, err := NewMiddleware(ACLdata)
	if err != nil {
		lis.Close()
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(m.UnaryRPC),
		grpc.StreamInterceptor(m.Stream),
	)

	RegisterBizServer(server, NewBizServer())
	RegisterAdminServer(server, NewAdminServer(m.sessions))

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalln("failed to serve: ", err)
		}
	}()

	return nil
}
