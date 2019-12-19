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

	mware, err := midware.New(ACLdata)
	if err != nil {
		lis.Close()
		return err
	}

	grpcS := grpc.NewServer(
		grpc.UnaryInterceptor(mware.UnaryRPC),
		grpc.StreamInterceptor(mware.Stream),
	)

	proto.RegisterBizServer(grpcS, servers.NewBiz())
	proto.RegisterAdminServer(
		grpcS, servers.NewAdmin(mware.LogWaitersList()),
	)

	go mware.StartLogger()

	go func() {
		if err := grpcS.Serve(lis); err != nil {
			log.Fatalln("failed to serve: ", err)
		}
	}()

	go func() {
		<-ctx.Done()
		mware.StopLogger()
		grpcS.GracefulStop()
	}()

	return nil
}
