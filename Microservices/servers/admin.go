package servers

import (
	"github.com/madmaxeatfax/homeworks/Microservices/midware"
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type admin struct {
	log *midware.Logger
}

func NewAdmin(logger *midware.Logger) proto.AdminServer {
	return &admin{logger}
}

func (a *admin) Logging(req *proto.Nothing, srv proto.Admin_LoggingServer) error {
	ch := make(chan *proto.Event)

	a.log.RLock()
	a.log.Tunnels[ch] = true
	a.log.RUnlock()

	for event := range ch {
		if err := srv.Send(event); err != nil {
			delete(a.log.Tunnels, ch)
			return err
		}
	}

	return nil
}

func (*admin) Statistics(req *proto.StatInterval, srv proto.Admin_StatisticsServer) error {
	srv.Send(&proto.Stat{})
	return nil
}
