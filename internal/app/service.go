package app

import (
	"context"
	"fmt"

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

func (api *OvaMethodApi) Create(ctx context.Context, req *igrpc.CreateRequest) (*emptypb.Empty, error) {
	if err := api.validateCreateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
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

func (api *OvaMethodApi) validateCreateRequest(req *igrpc.CreateRequest) error {
	if len(req.Value) == 0 {
		return fmt.Errorf("value cannot be empty")
	}
	if req.UserId == 0 {
		return fmt.Errorf("user id is required field")
	}
	return nil
}

func (api *OvaMethodApi) Remove(ctx context.Context, req *igrpc.RemoveRequest) (*emptypb.Empty, error) {
	if err := api.validateRemoveRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
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

func (api *OvaMethodApi) validateRemoveRequest(req *igrpc.RemoveRequest) error {
	if req.Id == 0 {
		return fmt.Errorf("id is required field")
	}
	return nil
}

func (api *OvaMethodApi) Describe(ctx context.Context, req *igrpc.DescribeRequest) (*igrpc.DescribeResponse, error) {
	if err := api.validateDescribeRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
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

	return &igrpc.DescribeResponse{Info: method.String()}, nil
}

func (api *OvaMethodApi) validateDescribeRequest(req *igrpc.DescribeRequest) error {
	if req.Id == 0 {
		return fmt.Errorf("id is required field")
	}
	return nil
}

func (api *OvaMethodApi) List(ctx context.Context, req *igrpc.ListRequest) (*igrpc.ListResponse, error) {
	if err := api.validateListRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
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

	methodList := &igrpc.ListResponse{
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

func (api *OvaMethodApi) validateListRequest(req *igrpc.ListRequest) error {
	if req.Limit == 0 {
		return fmt.Errorf("incorrect limit value")
	}
	return nil
}
