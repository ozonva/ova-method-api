package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
	igrpc "ova-method-api/pkg/ova-method-api"
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

	err := api.rep.Add([]model.Method{
		{
			UserId: req.UserId,
			Value:  req.Value,
		},
	})
	if err != nil {
		// TODO log
		return nil, status.Errorf(codes.Internal, "failed create method")
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) Remove(ctx context.Context, req *igrpc.MethodIdRequest) (*emptypb.Empty, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required field")
	}

	if err := api.rep.Remove(req.Id); err != nil {
		// TODO log
		return nil, status.Errorf(codes.Internal, "failed remove method")
	}

	return nil, nil
}

func (api *OvaMethodApi) Describe(ctx context.Context, req *igrpc.MethodIdRequest) (*igrpc.MethodInfo, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required field")
	}

	method, err := api.rep.Describe(req.Id)
	if err == repo.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "method not found")
	}
	if err != nil {
		// TODO log
		return nil, status.Errorf(codes.Internal, "failed describe method")
	}

	return &igrpc.MethodInfo{Info: method.String()}, nil
}

func (api *OvaMethodApi) List(ctx context.Context, req *igrpc.MethodListRequest) (*igrpc.MethodList, error) {
	if req.Limit == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "incorrect limit value")
	}

	methods, err := api.rep.List(req.Limit, req.Offset)
	if err == repo.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "method not found")
	}
	if err != nil {
		// TODO log
		return nil, status.Errorf(codes.Internal, "failed list method")
	}

	methodList := &igrpc.MethodList{
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
