package app

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"

	igrpc "ova-method-api/pkg/ova-method-api"
)

type OvaMethodApi struct {
	igrpc.UnimplementedOvaMethodApiServer
}

func NewOvaMethodApi() igrpc.OvaMethodApiServer {
	return &OvaMethodApi{}
}

func (api *OvaMethodApi) Create(ctx context.Context, req *igrpc.CreateMethodRequest) (*emptypb.Empty, error) {
	log.Debug().
		Str("value", req.Value).
		Uint64("user_id", req.UserId).
		Msg("create new method")

	return nil, nil
}

func (api *OvaMethodApi) Remove(ctx context.Context, req *igrpc.MethodIdRequest) (*emptypb.Empty, error) {
	log.Debug().Uint64("id", req.Id).Msg("remove method")
	return nil, nil
}

func (api *OvaMethodApi) Describe(ctx context.Context, req *igrpc.MethodIdRequest) (*igrpc.MethodInfo, error) {
	log.Debug().Uint64("id", req.Id).Msg("describe method")
	return nil, nil
}

func (api *OvaMethodApi) List(ctx context.Context, req *igrpc.MethodListRequest) (*igrpc.MethodList, error) {
	log.Debug().
		Uint64("limit", req.Limit).
		Uint64("offset", req.Offset).
		Msg("list method")

	return nil, nil
}
