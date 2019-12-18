package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/madmaxeatfax/homeworks/Microservices/midware"
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
	"github.com/madmaxeatfax/homeworks/Microservices/servers"
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

	m, err := midware.New(ACLdata)
	if err != nil {
		lis.Close()
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(m.UnaryRPC),
		grpc.StreamInterceptor(m.Stream),
	)

	proto.RegisterBizServer(grpcServer, servers.NewBiz())
	proto.RegisterAdminServer(
		grpcServer,
		servers.NewAdmin(m.GetSess()),
	)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalln("failed to serve: ", err)
		}
	}()

	return nil
}
