package etcd

import (
	"fmt"

	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"golang.org/x/net/context"
)

func checkEtcdError(err error) error {
	if err != nil {
		switch err {
		case context.Canceled:
			return ErrCanceled
		case context.DeadlineExceeded:
			return ErrDeadlineExceeded
		case rpctypes.ErrEmptyKey:
			return fmt.Errorf("client-side error:%v\n", err)
		default:
			return fmt.Errorf("bad cluster endpoints, which are not etcd servers:%v\n", err)
		}
	}

	return nil
}
