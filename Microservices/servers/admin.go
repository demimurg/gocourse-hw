package servers

import (
	"sync"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type admin struct {
	sync.Mutex

	logRecievers *[]chan *proto.Event
}

// NEW CONFIG TYPE???

func  NewAdmin(logs *[]chan *proto.Event) proto.AdminServer {
	return &admin{logRecievers: logs}
}

func (a *admin) Logging(req *proto.Nothing, srv proto.Admin_LoggingServer) error {
	ch := make(chan *proto.Event)

	a.Lock()
	*a.logRecievers = append(*a.logRecievers, ch)
	a.Unlock()

	for event := range ch {
		if err := srv.Send(event); err != nil {
			return err
		}
	}

	return nil
}

func (*admin) Statistics(req *proto.StatInterval, srv proto.Admin_StatisticsServer) error {
	srv.Send(&proto.Stat{})
	return nil
}
