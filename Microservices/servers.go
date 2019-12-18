package main

import (
	"context"
)

type Biz struct{}

func NewBizServer() *Biz {
	return &Biz{}
}

func (*Biz) Check(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
func (*Biz) Add(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
func (*Biz) Test(ctx context.Context, req *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

type Admin struct {
	sessions []*Event
}

func NewAdminServer(events []*Event) *Admin {
	return &Admin{events}
}

func (a *Admin) Logging(req *Nothing, srv Admin_LoggingServer) error {
	i := srv.Context().Value("track-from").(int)
	tracked := a.sessions[i:]

	for _, ses := range tracked {
		if err := srv.Send(ses); err != nil {
			return err
		}
	}

	return nil
}

func (*Admin) Statistics(req *StatInterval, srv Admin_StatisticsServer) error {
	srv.Send(&Stat{})
	return nil
}
