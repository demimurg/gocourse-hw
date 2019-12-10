package main

import (
	"context"
)

type Biz struct{}

func (*Biz) Check(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
func (*Biz) Add(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
func (*Biz) Test(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

type Admin struct{}

func (*Admin) Logging(req *Nothing, srv Admin_LoggingServer) error {
	srv.Send(&Event{})
	return nil
}
func (*Admin) Statistics(req *StatInterval, srv Admin_StatisticsServer) error {
	srv.Send(&Stat{})
	return nil
}
