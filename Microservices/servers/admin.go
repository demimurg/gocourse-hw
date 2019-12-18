package servers

import (
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type admin struct {
	sessions []*proto.Event
}

// CONFIG TYPE???

func NewAdmin(events []*proto.Event) proto.AdminServer {
	return &admin{events}
}

func (a *admin) Logging(req *proto.Nothing, srv proto.Admin_LoggingServer) error {
	// i := srv.Context().Value("track-from").(int)
	// tracked := a.sessions[i:]

	// for _, ses := range tracked {
	// 	if err := srv.Send(ses); err != nil {
	// 		return err
	// 	}
	// }
	srv.Send(&proto.Event{})
	return nil
}

func (*admin) Statistics(req *proto.StatInterval, srv proto.Admin_StatisticsServer) error {
	srv.Send(&proto.Stat{})
	return nil
}
