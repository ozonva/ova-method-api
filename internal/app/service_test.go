package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"ova-method-api/internal/model"
	iqueue "ova-method-api/internal/queue"
	qmock "ova-method-api/internal/queue/mock"
	"ova-method-api/internal/repo"
	"ova-method-api/internal/repo/mock"
	proto "ova-method-api/pkg/ova-method-api"
)

const (
	listenAddr string = "localhost:3000"
)

var (
	server  *grpc.Server
	conn    *grpc.ClientConn
	client  proto.OvaMethodApiClient
	service СonfigurableOvaMethodApi

	ctrl  = gomock.NewController(GinkgoT())
	rep   = mock.NewMockMethodRepo(ctrl)
	queue = qmock.NewMockQueue(ctrl)

	method  = model.Method{UserId: 1, Value: "hello"}
	txProxy = func(ctx context.Context, fn func(rep repo.MethodRepo) error) error {
		return fn(rep)
	}

	defaultTopic = "ova-method"
	defaultCtx   = context.Background()
	defaultErr   = fmt.Errorf("something went wrong")
)

func TestOvaMethodApi(t *testing.T) {
	initLoggerStub()
	RegisterFailHandler(Fail)
	RunSpecs(t, "OvaMethodApi suites")
}

var _ = BeforeSuite(func() {
	server = grpc.NewServer()
	service = NewOvaMethodApi(rep, queue)
	proto.RegisterOvaMethodApiServer(server, service)

	go func() {
		listen, err := net.Listen("tcp", listenAddr)
		if err != nil {
			GinkgoT().Fatalf("failed create net listen: %v", err)
		}

		if err = server.Serve(listen); err != nil {
			GinkgoT().Fatalf("failed start grpc server: %v", err)
		}
	}()

	connection, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		GinkgoT().Fatalf("failed connect to grpc: %v", err)
	}

	conn = connection
	client = proto.NewOvaMethodApiClient(conn)
})

var _ = AfterSuite(func() {
	if err := conn.Close(); err != nil {
		GinkgoT().Fatalf("failed close grpc connection: %v", err)
	}
	server.Stop()
})

var _ = Describe("OvaMethodApi", func() {
	Describe("Create", func() {
		DescribeTable("check error",
			func(req *proto.CreateRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Create(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("invalid value", makeCreateReq(1, ""), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("invalid user_id", makeCreateReq(0, "1"), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep error", makeCreateReq(1, "1"), func() (*emptypb.Empty, codes.Code) {
				rep.EXPECT().Add(gomock.Any(), []model.Method{{UserId: 1, Value: "1"}}).Return(nil, defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().
				Add(gomock.Any(), []model.Method{{UserId: 1, Value: "1"}}).
				Return([]model.Method{{Id: 1}}, nil)

			queue.EXPECT().Send(defaultTopic, makeQueueMsg("created", 1)).Return(nil)

			result, err := client.Create(defaultCtx, makeCreateReq(1, "1"))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("MultiCreate", func() {
		DescribeTable("check error",
			func(req *proto.MultiCreateRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.MultiCreate(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("invalid value",
				makeMultiCreateRequest(makeCreateReq(0, "1"), makeCreateReq(1, "1")),
				func() (*emptypb.Empty, codes.Code) {
					return nil, codes.InvalidArgument
				}),
			Entry("invalid user_id",
				makeMultiCreateRequest(makeCreateReq(1, "1"), makeCreateReq(1, "")),
				func() (*emptypb.Empty, codes.Code) {
					return nil, codes.InvalidArgument
				}),
			Entry("rep error",
				makeMultiCreateRequest(makeCreateReq(1, "1")),
				func() (*emptypb.Empty, codes.Code) {
					rep.EXPECT().Transaction(gomock.Any(), gomock.Any()).Do(txProxy).Return(defaultErr)
					rep.EXPECT().Add(gomock.Any(), []model.Method{{UserId: 1, Value: "1"}}).Return(nil, defaultErr)

					return nil, codes.Internal
				}),
		)

		Context("with configure ova service", func() {
			BeforeEach(func() {
				service.SetChunkSize(0)
			})
			AfterEach(func() {
				service.SetChunkSize(2)
			})

			It("failed split to chunk", func() {
				req := makeMultiCreateRequest(makeCreateReq(1, "1"))
				result, err := client.MultiCreate(defaultCtx, req)
				st, _ := status.FromError(err)

				var expectRes *emptypb.Empty
				Expect(st.Code()).To(Equal(codes.Internal))
				Expect(result).To(Equal(expectRes))
			})
		})

		It("successful", func() {
			rep.EXPECT().Transaction(gomock.Any(), gomock.Any()).Do(txProxy).Return(nil)
			rep.EXPECT().
				Add(gomock.Any(), []model.Method{{UserId: 1, Value: "1"}}).
				Return([]model.Method{{Id: 1}}, nil)

			queue.EXPECT().Send(defaultTopic, makeQueueMsg("created", 1)).Return(nil)

			result, err := client.MultiCreate(defaultCtx, makeMultiCreateRequest(makeCreateReq(1, "1")))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("Update", func() {
		DescribeTable("check error",
			func(req *proto.UpdateRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Update(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("required id field", makeUpdateReq(0, "1"), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("invalid value", makeUpdateReq(1, ""), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep error", makeUpdateReq(1, "1"), func() (*emptypb.Empty, codes.Code) {
				rep.EXPECT().Update(gomock.Any(), uint64(1), "1").Return(defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Update(gomock.Any(), uint64(1), "1").Return(nil)
			queue.EXPECT().Send(defaultTopic, makeQueueMsg("updated", 1)).Return(nil)

			result, err := client.Update(defaultCtx, makeUpdateReq(1, "1"))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("Remove", func() {
		DescribeTable("check error",
			func(req *proto.RemoveRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Remove(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("required id field", makeRemoveReq(0), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep error", makeRemoveReq(1), func() (*emptypb.Empty, codes.Code) {
				rep.EXPECT().Remove(gomock.Any(), uint64(1)).Return(defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Remove(gomock.Any(), uint64(1)).Return(nil)
			queue.EXPECT().Send(defaultTopic, makeQueueMsg("deleted", 1)).Return(nil)

			result, err := client.Remove(defaultCtx, makeRemoveReq(1))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("Describe", func() {
		DescribeTable("check error",
			func(req *proto.DescribeRequest, getExpectedRes func() (*proto.DescribeResponse, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Describe(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("required id field", makeDescribeReq(0), func() (*proto.DescribeResponse, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep not found", makeDescribeReq(1), func() (*proto.DescribeResponse, codes.Code) {
				rep.EXPECT().Describe(gomock.Any(), uint64(1)).Return(nil, repo.ErrNoRows)
				return nil, codes.NotFound
			}),
			Entry("rep error", makeDescribeReq(1), func() (*proto.DescribeResponse, codes.Code) {
				rep.EXPECT().Describe(gomock.Any(), uint64(1)).Return(nil, defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Describe(gomock.Any(), uint64(1)).Return(&method, nil)

			result, err := client.Describe(defaultCtx, makeDescribeReq(1))
			Expect(err).To(BeNil())
			Expect(result.Info).To(Equal(method.String()))
		})
	})

	Describe("List", func() {
		DescribeTable("check error",
			func(req *proto.ListRequest, getExpectedRes func() (*proto.ListResponse, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.List(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("incorrect limit", makeListReq(0, 0),
				func() (*proto.ListResponse, codes.Code) {
					return nil, codes.InvalidArgument
				}),
			Entry("rep error", makeListReq(1, 0),
				func() (*proto.ListResponse, codes.Code) {
					rep.EXPECT().List(gomock.Any(), uint64(1), uint64(0)).Return(nil, defaultErr)
					return nil, codes.Internal
				}),
		)

		It("rep not found", func() {
			rep.EXPECT().List(gomock.Any(), uint64(1), uint64(0)).Return(nil, repo.ErrNoRows)

			result, err := client.List(defaultCtx, makeListReq(1, 0))
			Expect(err).To(BeNil())
			Expect(len(result.Methods)).To(Equal(0))
		})

		It("successful", func() {
			rep.EXPECT().List(gomock.Any(), uint64(2), uint64(0)).Return([]model.Method{method, method}, nil)

			result, err := client.List(defaultCtx, makeListReq(2, 0))

			id := func(index int, _ interface{}) string {
				return strconv.Itoa(index)
			}

			Expect(err).To(BeNil())
			Expect(result).To(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Methods": MatchAllElementsWithIndex(id, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"UserId": Equal(uint64(1)),
							"Value":  Equal("hello"),
						})),
						"1": PointTo(MatchFields(IgnoreExtras, Fields{
							"UserId": Equal(uint64(1)),
							"Value":  Equal("hello"),
						})),
					}),
				})),
			)
		})
	})
})

func initLoggerStub() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: ioutil.Discard})
}

func makeCreateReq(userId uint64, value string) *proto.CreateRequest {
	return &proto.CreateRequest{
		UserId: userId,
		Value:  value,
	}
}

func makeMultiCreateRequest(items ...*proto.CreateRequest) *proto.MultiCreateRequest {
	return &proto.MultiCreateRequest{Methods: items}
}

func makeUpdateReq(id uint64, value string) *proto.UpdateRequest {
	return &proto.UpdateRequest{
		Id:    id,
		Value: value,
	}
}

func makeRemoveReq(id uint64) *proto.RemoveRequest {
	return &proto.RemoveRequest{
		Id: id,
	}
}

func makeDescribeReq(id uint64) *proto.DescribeRequest {
	return &proto.DescribeRequest{
		Id: id,
	}
}

func makeListReq(limit, offset uint64) *proto.ListRequest {
	return &proto.ListRequest{
		Limit:  limit,
		Offset: offset,
	}
}

func makeQueueMsg(action string, id uint64) iqueue.QueueMsg {
	return iqueue.NewMessage(action, iqueue.Body{
		"id": id,
	})
}
