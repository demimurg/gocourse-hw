package servers

import (
	"context"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type biz struct{}

func NewBiz() proto.BizServer {
	return &biz{}
}

func (*biz) Check(ctx context.Context, req *proto.Nothing) (*proto.Nothing, error) {
	return &proto.Nothing{}, nil
}
func (*biz) Add(ctx context.Context, req *proto.Nothing) (*proto.Nothing, error) {
	return &proto.Nothing{}, nil
}
func (*biz) Test(ctx context.Context, req *proto.Nothing) (*proto.Nothing, error) {
	return &proto.Nothing{}, nil
}
