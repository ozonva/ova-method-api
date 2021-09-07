package flusher

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"ova-method-api/internal/model"
	"ova-method-api/internal/repo/mock"
)

func TestFlusher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flusher suites")
}

var _ = Describe("Flusher", func() {
	var (
		ctrl = gomock.NewController(GinkgoT())
		rep  = mock.NewMockMethodRepo(ctrl)

		flushErr   = fmt.Errorf("flush error")
		defaultCtx = context.Background()
	)
	defer ctrl.Finish()

	Describe("Flush", func() {
		sequence := []model.Method{{UserId: 1}}

		DescribeTable("not flushed",
			func(chunkSize int, toFlush []model.Method, expected []model.Method) {
				result := New(chunkSize, rep).Flush(defaultCtx, toFlush)
				Expect(result).To(Equal(expected))
			},
			Entry("chunk 0", 0, sequence, sequence),
			Entry("chunk -1", -1, sequence, sequence),
			Entry("nothing to flush", 1, nil, nil),
		)

		DescribeTable("repository add equal",
			func(toFlush []model.Method, expected []model.Method, err error) {
				rep.EXPECT().Add(defaultCtx, toFlush).Return(nil, err)
				result := New(len(toFlush), rep).Flush(defaultCtx, toFlush)
				Expect(result).To(Equal(expected))
			},
			Entry("add error", sequence, sequence, flushErr),
			Entry("add success", sequence, nil, nil),
		)

		It("partial flush", func() {
			rep.EXPECT().Add(defaultCtx, []model.Method{{UserId: 1}}).Return(nil, nil)
			rep.EXPECT().Add(defaultCtx, []model.Method{{UserId: 2}}).Return(nil, flushErr)

			result := New(1, rep).Flush(defaultCtx, []model.Method{{UserId: 1}, {UserId: 2}})

			Expect(result).To(Equal([]model.Method{{UserId: 2}}))
		})
	})
})
