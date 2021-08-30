package app

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

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
	"ova-method-api/internal/repo"
	"ova-method-api/internal/repo/mock"
	proto "ova-method-api/pkg/ova-method-api"
)

const (
	listenAddr string = "localhost:3000"
)

var (
	server *grpc.Server
	conn   *grpc.ClientConn
	client proto.OvaMethodApiClient

	ctrl = gomock.NewController(GinkgoT())
	rep  = mock.NewMockMethodRepo(ctrl)

	method     = model.Method{UserId: 1, Value: "hello"}
	defaultCtx = context.Background()
	defaultErr = fmt.Errorf("something went wrong")
)

func TestOvaMethodApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OvaMethodApi suites")
}

var _ = BeforeSuite(func() {
	server = grpc.NewServer()
	proto.RegisterOvaMethodApiServer(server, NewOvaMethodApi(rep))

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
			func(req *proto.CreateMethodRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
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
				rep.EXPECT().Add([]model.Method{{UserId: 1, Value: "1"}}).Return(defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Add([]model.Method{{UserId: 1, Value: "1"}}).Return(nil)

			result, err := client.Create(defaultCtx, makeCreateReq(1, "1"))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("Remove", func() {
		DescribeTable("check error",
			func(req *proto.MethodIdRequest, getExpectedRes func() (*emptypb.Empty, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Remove(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("required id field", makeIdReq(0), func() (*emptypb.Empty, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep error", makeIdReq(1), func() (*emptypb.Empty, codes.Code) {
				rep.EXPECT().Remove(uint64(1)).Return(defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Remove(uint64(1)).Return(nil)

			result, err := client.Remove(defaultCtx, makeIdReq(1))
			Expect(err).To(BeNil())
			Expect(result).Should(BeAssignableToTypeOf(&emptypb.Empty{}))
		})
	})

	Describe("Describe", func() {
		DescribeTable("check error",
			func(req *proto.MethodIdRequest, getExpectedRes func() (*proto.MethodInfoResponse, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.Describe(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("required id field", makeIdReq(0), func() (*proto.MethodInfoResponse, codes.Code) {
				return nil, codes.InvalidArgument
			}),
			Entry("rep not found", makeIdReq(1), func() (*proto.MethodInfoResponse, codes.Code) {
				rep.EXPECT().Describe(uint64(1)).Return(nil, repo.ErrNoRows)
				return nil, codes.NotFound
			}),
			Entry("rep error", makeIdReq(1), func() (*proto.MethodInfoResponse, codes.Code) {
				rep.EXPECT().Describe(uint64(1)).Return(nil, defaultErr)
				return nil, codes.Internal
			}),
		)

		It("successful", func() {
			rep.EXPECT().Describe(uint64(1)).Return(&method, nil)

			result, err := client.Describe(defaultCtx, makeIdReq(1))
			Expect(err).To(BeNil())
			Expect(result.Info).To(Equal(method.String()))
		})
	})

	Describe("List", func() {
		DescribeTable("check error",
			func(req *proto.MethodListRequest, getExpectedRes func() (*proto.MethodListResponse, codes.Code)) {
				expectRes, expectCode := getExpectedRes()
				result, err := client.List(defaultCtx, req)
				st, _ := status.FromError(err)

				Expect(st.Code()).To(Equal(expectCode))
				Expect(result).To(Equal(expectRes))
			},
			Entry("incorrect limit", makeListReq(0, 0),
				func() (*proto.MethodListResponse, codes.Code) {
					return nil, codes.InvalidArgument
				}),
			Entry("rep error", makeListReq(1, 0),
				func() (*proto.MethodListResponse, codes.Code) {
					rep.EXPECT().List(uint64(1), uint64(0)).Return(nil, defaultErr)
					return nil, codes.Internal
				}),
		)

		It("rep not found", func() {
			rep.EXPECT().List(uint64(1), uint64(0)).Return(nil, repo.ErrNoRows)

			result, err := client.List(defaultCtx, makeListReq(1, 0))
			Expect(err).To(BeNil())
			Expect(len(result.Methods)).To(Equal(0))
		})

		It("successful", func() {
			rep.EXPECT().List(uint64(2), uint64(0)).Return([]model.Method{method, method}, nil)

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

func makeCreateReq(userId uint64, value string) *proto.CreateMethodRequest {
	return &proto.CreateMethodRequest{
		UserId: userId,
		Value:  value,
	}
}

func makeIdReq(id uint64) *proto.MethodIdRequest {
	return &proto.MethodIdRequest{
		Id: id,
	}
}

func makeListReq(limit, offset uint64) *proto.MethodListRequest {
	return &proto.MethodListRequest{
		Limit:  limit,
		Offset: offset,
	}
}
