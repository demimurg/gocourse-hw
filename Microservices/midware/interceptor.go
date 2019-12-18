package midware

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

func (m *middleware) interceptor(consumer, method string) error {
	for _, f := range []step{
		m.checkACL, m.addLog,
	} {
		if err := f(consumer, method); err != nil {
			return err
		}
	}

	return nil
}

// it would be better to rename
type step func(consumer, method string) error

func (m *middleware) checkACL(consumer, method string) error {
	var (
		found   bool
		methods []string
	)
	for con, meths := range m.acl {
		if con == consumer {
			found = true
			methods = meths
			break
		}
	}
	if !found {
		return status.Errorf(
			codes.Unauthenticated, "consumer doesn't exist",
		)
	}

	var granted bool
	for _, m := range methods {
		if strings.HasSuffix(m, "*") || m == method {
			granted = true
			break
		}
	}
	if !granted {
		return status.Errorf(
			codes.Unauthenticated, "disallowed method",
		)
	}

	return nil
}

func (m *middleware) addLog(consumer, method string) error {
	// m.Lock()
	m.sessions = append(m.sessions, &proto.Event{
		Consumer: consumer,
		Method:   method,
	})
	// m.Unlock()

	return nil
}
