package servers

import (
	"github.com/madmaxeatfax/homeworks/Microservices/midware"
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type admin struct {
	log *midware.Logger
}

// NEW CONFIG TYPE???

func NewAdmin(logger *midware.Logger) proto.AdminServer {
	return &admin{logger}
}

func (a *admin) Logging(req *proto.Nothing, srv proto.Admin_LoggingServer) error {
	ch := make(chan *proto.Event)
	logger := a.log

	logger.RLock()
	logger.Tunnels = append(logger.Tunnels, ch)
	logger.RUnlock()

	for event := range ch {
		if err := srv.Send(event); err != nil {
			// you should remove the chan from the tunnels
			return err
		}
	}

	return nil
}

func (*admin) Statistics(req *proto.StatInterval, srv proto.Admin_StatisticsServer) error {
	srv.Send(&proto.Stat{})
	return nil
}
