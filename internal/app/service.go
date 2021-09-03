package app

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"ova-method-api/internal"
	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
	igrpc "ova-method-api/pkg/ova-method-api"
)

var (
	RequiredIdValidationErr = fmt.Errorf("id is required field")
	EmptyValueValidationErr = fmt.Errorf("value cannot be empty")

	internalGrpcErr = status.Errorf(codes.Internal, "failed to process request")
)

const (
	chunkSizeToSave = 2
)

type СonfigurableOvaMethodApi interface {
	igrpc.OvaMethodApiServer

	SetChunkSize(chunkSize int)
}

type OvaMethodApi struct {
	rep       repo.MethodRepo
	chunkSize int

	igrpc.UnimplementedOvaMethodApiServer
}

func NewOvaMethodApi(rep repo.MethodRepo) СonfigurableOvaMethodApi {
	return &OvaMethodApi{rep: rep, chunkSize: chunkSizeToSave}
}

func (api *OvaMethodApi) SetChunkSize(chunkSize int) {
	api.chunkSize = chunkSize
}

func (api *OvaMethodApi) Create(ctx context.Context, req *igrpc.CreateRequest) (*emptypb.Empty, error) {
	if err := api.validateCreateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	err := api.rep.Add([]model.Method{api.makeMethodModelFromReq(req)})
	if err != nil {
		log.Error().
			Uint64("user_id", req.UserId).
			Str("value", req.Value).
			Err(err).
			Msg("failed create method")

		return nil, internalGrpcErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) validateCreateRequest(req *igrpc.CreateRequest) error {
	if len(req.Value) == 0 {
		return EmptyValueValidationErr
	}
	if req.UserId == 0 {
		return fmt.Errorf("user id is required field")
	}
	return nil
}

func (api *OvaMethodApi) makeMethodModelFromReq(req *igrpc.CreateRequest) model.Method {
	return model.Method{
		UserId: req.UserId,
		Value:  req.Value,
	}
}

func (api *OvaMethodApi) MultiCreate(ctx context.Context, req *igrpc.MultiCreateRequest) (*emptypb.Empty, error) {
	if err := api.validateMultiCreateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	models := make([]model.Method, 0, len(req.Methods))
	for _, createReq := range req.Methods {
		models = append(models, api.makeMethodModelFromReq(createReq))
	}

	chunkedMethods, err := internal.ListOfMethodToChunkSlice(models, api.chunkSize)
	if err != nil {
		log.Error().
			Int("methods len", len(models)).
			Int("chunk size", api.chunkSize).
			Err(err).
			Msg("failed split to chunk")

		return nil, internalGrpcErr
	}

	err = api.rep.Transaction(func(rep repo.MethodRepo) error {
		for _, methods := range chunkedMethods {
			if err = api.rep.Add(methods); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("failed multi add")
		return nil, internalGrpcErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) validateMultiCreateRequest(req *igrpc.MultiCreateRequest) error {
	for index, createReq := range req.Methods {
		if err := api.validateCreateRequest(createReq); err != nil {
			return errors.Wrapf(err, "method[%d] error", index)
		}
	}
	return nil
}

func (api *OvaMethodApi) Update(ctx context.Context, req *igrpc.UpdateRequest) (*emptypb.Empty, error) {
	if err := api.validateUpdateRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	if err := api.rep.Update(req.Id, req.Value); err != nil {
		log.Error().
			Uint64("id", req.Id).
			Str("value", req.Value).
			Err(err).
			Msg("failed update method")

		return nil, internalGrpcErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) validateUpdateRequest(req *igrpc.UpdateRequest) error {
	if req.Id == 0 {
		return RequiredIdValidationErr
	}
	if len(req.Value) == 0 {
		return EmptyValueValidationErr
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

		return nil, internalGrpcErr
	}

	return &emptypb.Empty{}, nil
}

func (api *OvaMethodApi) validateRemoveRequest(req *igrpc.RemoveRequest) error {
	if req.Id == 0 {
		return RequiredIdValidationErr
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

		return nil, internalGrpcErr
	}

	return &igrpc.DescribeResponse{Info: method.String()}, nil
}

func (api *OvaMethodApi) validateDescribeRequest(req *igrpc.DescribeRequest) error {
	if req.Id == 0 {
		return RequiredIdValidationErr
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

		return nil, internalGrpcErr
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
