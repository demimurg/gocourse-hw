package servers

import (
	"time"

	"github.com/madmaxeatfax/homeworks/Microservices/midware"
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

func NewAdmin(log midware.Logger) proto.AdminServer {
	return &admin{log}
}

type admin struct {
	log midware.Logger
}

func (a *admin) Logging(req *proto.Nothing, srv proto.Admin_LoggingServer) error {
	var (
		logCh = a.log.NewTunnel()
		event = proto.Event{Host: "127.0.0.1:"}
	)

	for recieved := range logCh {
		event.Consumer = recieved.Consumer
		event.Method = recieved.Method

		if err := srv.Send(&event); err != nil {
			a.log.DeleteTunnel(logCh)
			return err
		}
	}

	return nil
}

func (a *admin) Statistics(req *proto.StatInterval, srv proto.Admin_StatisticsServer) error {
	var (
		logCh  = a.log.NewTunnel()
		ticker = time.NewTicker(
			time.Duration(req.IntervalSeconds) * time.Second,
		)

		statistics = proto.Stat{
			ByConsumer: make(map[string]uint64, 0),
			ByMethod:   make(map[string]uint64, 0),
		}

		increment = func(stat map[string]uint64, key string) {
			if _, ok := stat[key]; !ok {
				stat[key] = 0
			}
			stat[key]++
		}
		clear = func(stat *map[string]uint64) {
			*stat = make(map[string]uint64, 0)
		}
	)

	defer ticker.Stop()

	for {
		select {
		case recieved, ok := <-logCh:
			if !ok {
				return nil
			}
			increment(statistics.ByConsumer, recieved.Consumer)
			increment(statistics.ByMethod, recieved.Method)
		case <-ticker.C:
			if err := srv.Send(&statistics); err != nil {
				a.log.DeleteTunnel(logCh)
				return err
			}
			clear(&statistics.ByConsumer)
			clear(&statistics.ByMethod)
		}
	}
}
