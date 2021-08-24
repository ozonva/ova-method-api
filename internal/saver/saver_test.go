package saver

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"ova-method-api/internal/flusher"
	"ova-method-api/internal/model"
	"ova-method-api/internal/repo/mock"
)

func TestSaver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Saver suites")
}

var _ = Describe("Saver", func() {
	var (
		ctrl = gomock.NewController(GinkgoT())
		rep  = mock.NewMockMethodRepo(ctrl)

		method     = model.Method{UserId: 1}
		defaultCtx = context.Background()
		flushErr   = fmt.Errorf("flush error")
	)
	defer ctrl.Finish()

	cancelableCtx, cancel := context.WithCancel(context.Background())

	DescribeTable("Save success",
		func(ctx context.Context, delay uint, fn func(ctx context.Context, s Saver)) {
			var success bool
			var wg sync.WaitGroup

			wg.Add(1)
			rep.EXPECT().Add([]model.Method{method}).DoAndReturn(func(items []model.Method) error {
				defer wg.Done()
				success = true
				return nil
			})

			flushService := flusher.New(1, rep)
			saverService := New(ctx, 1, delay, flushService)

			fn(ctx, saverService)

			wg.Wait()
			Expect(success).To(Equal(true))
		},
		Entry("flush after buffer full", defaultCtx, uint(10), func(ctx context.Context, s Saver) {
			s.Save(method)
			s.Save(method)
		}),
		Entry("flush after delay", defaultCtx, uint(1), func(ctx context.Context, s Saver) {
			s.Save(method)
			time.Sleep(1100 * time.Millisecond)
		}),
		Entry("flush after close", defaultCtx, uint(10), func(ctx context.Context, s Saver) {
			s.Save(method)
			s.Close()
		}),
		Entry("flush after context done", cancelableCtx, uint(10), func(ctx context.Context, s Saver) {
			s.Save(method)
			cancel()
			time.Sleep(100 * time.Millisecond)
		}),
	)

	It("nothing to save", func() {
		flushService := flusher.New(1, rep)
		saverService := New(defaultCtx, 1, 1, flushService)
		saverService.Close()
	})

	It("panic after retry", func() {
		rep.EXPECT().Add([]model.Method{method}).Return(flushErr)
		rep.EXPECT().Add([]model.Method{method}).Return(flushErr)

		flushService := flusher.New(1, rep)
		saverService := New(defaultCtx, 1, 1, flushService)

		saverService.Save(method)
		Expect(saverService.Close).Should(Panic())
	})
})
