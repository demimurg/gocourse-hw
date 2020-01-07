package midware

import (
	"encoding/json"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// map{consumer: [method1, method2...]}
type aclmod map[string][]string

func (acl *aclmod) scan(data string) error {
	*acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(data), acl)
	if err != nil {
		return err
	}

	return nil
}

func (acl aclmod) check(consumer, method string) error {
	var (
		found   bool
		methods []string
	)
	for con, meths := range acl {
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
