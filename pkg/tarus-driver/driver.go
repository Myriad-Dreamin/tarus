package tarus_driver

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"sync"
)

type Driver interface {
	CreateJudgeRequest(ctx context.Context) (*tarus.MakeJudgeRequest, error)
}

type InitContext struct {
	Arguments map[string]string
}

type Registration struct {
	Id   string
	Init func(args *InitContext) (Driver, error)
}

var register = struct {
	sync.RWMutex
	r []*Registration
}{}

func Register(fac *Registration) {
	register.Lock()
	defer register.Unlock()

	register.r = append(register.r, fac)
}

func FindById(id string) (r []*Registration, _ error) {
	for i := range register.r {
		if register.r[i].Id == id {
			r = append(r, register.r[i])
		}
	}
	return
}
