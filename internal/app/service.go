package app

import (
	"context"

	"github.com/rs/zerolog/log"
	igrpc "ova-method-api/pkg/ova-method-api"
)

type OvaMethodApi struct {
	igrpc.UnimplementedOvaMethodApiServer
}

func NewOvaMethodApi() igrpc.OvaMethodApiServer {
	return &OvaMethodApi{}
}

func (api *OvaMethodApi) Create(ctx context.Context, req *igrpc.CreateMethodRequest) (*igrpc.MethodItem, error) {
	log.Debug().Str("value", req.Value).Msg("create new method")
	return nil, nil
}

func (api *OvaMethodApi) Remove(ctx context.Context, req *igrpc.MethodIdRequest) (*igrpc.Status, error) {
	log.Debug().Uint64("id", req.Id).Msg("remove method")
	return nil, nil
}

func (api *OvaMethodApi) Describe(ctx context.Context, req *igrpc.MethodIdRequest) (*igrpc.MethodInfo, error) {
	log.Debug().Uint64("id", req.Id).Msg("describe method")
	return nil, nil
}

func (api *OvaMethodApi) List(ctx context.Context, req *igrpc.MethodListRequest) (*igrpc.MethodList, error) {
	log.Debug().
		Uint64("page", req.Page).
		Uint64("limit", req.Limit).
		Msg("list method")

	return nil, nil
}
