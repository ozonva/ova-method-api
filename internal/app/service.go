package app

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	internalErr = status.Errorf(codes.Internal, "failed to process request")
)

type OvaMethodApi struct {
	rep repo.MethodRepo

	igrpc.UnimplementedOvaMethodApiServer
}

func NewOvaMethodApi(rep repo.MethodRepo) igrpc.OvaMethodApiServer {
	return &OvaMethodApi{rep: rep}
}

func (api *OvaMethodApi) Create(ctx context.Context, req *igrpc.CreateMethodRequest) (*emptypb.Empty, error) {
	if len(req.Value) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "value cannot be empty")
	}
	if req.UserId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user id is required field")
	}

	err := api.rep.Add([]model.Method{{UserId: req.UserId, Value: req.Value}})
	if err != nil {
		log.Error().
			Uint64("user_id", req.UserId).
			Str("value", req.Value).
			Err(err).
			Msg("failed create method")

		return nil, internalErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) Remove(ctx context.Context, req *igrpc.MethodIdRequest) (*emptypb.Empty, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required field")
	}

	if err := api.rep.Remove(req.Id); err != nil {
		log.Error().
			Uint64("id", req.Id).
			Err(err).
			Msg("failed remove method")

		return nil, internalErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) Describe(ctx context.Context, req *igrpc.MethodIdRequest) (*igrpc.MethodInfoResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required field")
	}

	method, err := api.rep.Describe(req.Id)
	if err == repo.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "method not found")
	}
	if err != nil {
		log.Error().
			Uint64("id", req.Id).
			Err(err).
			Msg("failed describe method")

		return nil, internalErr
	}

	return &igrpc.MethodInfoResponse{Info: method.String()}, nil
}

func (api *OvaMethodApi) List(ctx context.Context, req *igrpc.MethodListRequest) (*igrpc.MethodListResponse, error) {
	if req.Limit == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect limit value")
	}

	methods, err := api.rep.List(req.Limit, req.Offset)
	if err != nil && err != repo.ErrNoRows {
		log.Error().
			Uint64("limit", req.Limit).
			Uint64("offset", req.Offset).
			Err(err).
			Msg("failed list method")

		return nil, internalErr
	}

	methodList := &igrpc.MethodListResponse{
		Methods: make([]*igrpc.MethodItem, 0, len(methods)),
	}

	for _, method := range methods {
		methodList.Methods = append(methodList.Methods, &igrpc.MethodItem{
			Id:        method.Id,
			UserId:    method.UserId,
			Value:     method.Value,
			CreatedAt: method.CreatedAt.Unix(),
		})
	}

	return methodList, nil
}
