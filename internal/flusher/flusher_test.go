package flusher

import (
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

		flushErr = fmt.Errorf("flush error")
	)
	defer ctrl.Finish()

	Describe("Flush", func() {
		sequence := []model.Method{{UserId: 1}}

		DescribeTable("not flushed",
			func(chunkSize int, toFlush []model.Method, expected []model.Method) {
				result := New(chunkSize, rep).Flush(toFlush)
				Expect(result).To(Equal(expected))
			},
			Entry("chunk 0", 0, sequence, sequence),
			Entry("chunk -1", -1, sequence, sequence),
			Entry("nothing to flush", 1, nil, []model.Method{}),
		)

		DescribeTable("repository add equal",
			func(toFlush []model.Method, expected []model.Method, err error) {
				rep.EXPECT().Add(toFlush).Return(err)
				result := New(len(toFlush), rep).Flush(toFlush)
				Expect(result).To(Equal(expected))
			},
			Entry("add error", sequence, sequence, flushErr),
			Entry("add success", sequence, []model.Method{}, nil),
		)

		It("partial flush", func() {
			rep.EXPECT().Add([]model.Method{{UserId: 1}}).Return(nil)
			rep.EXPECT().Add([]model.Method{{UserId: 2}}).Return(flushErr)

			result := New(1, rep).Flush([]model.Method{{UserId: 1}, {UserId: 2}})

			Expect(result).To(Equal([]model.Method{{UserId: 2}}))
		})
	})
})
