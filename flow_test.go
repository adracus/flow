package flow_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/adracus/flow"
	"github.com/adracus/flow/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFlow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flow Suite")
}

func mkError(id int) error {
	return fmt.Errorf("error %d", id)
}

func waitForContextToErrorAndReturnError(ctx context.Context) error {
	Eventually(ctx.Err).Should(HaveOccurred())
	return ctx.Err()
}

func waitForContextToErrorAndReturnStringError(ctx context.Context) (string, error) {
	Eventually(ctx.Err).Should(HaveOccurred())
	return "", ctx.Err()
}

func waitForContextToErrorAndReturnIntError(ctx context.Context) (int, error) {
	Eventually(ctx.Err).Should(HaveOccurred())
	return 0, ctx.Err()
}

func waitForContextToErrorAndReturnBoolError(ctx context.Context) (bool, error) {
	Eventually(ctx.Err).Should(HaveOccurred())
	return false, ctx.Err()
}

var _ = Describe("Flow", func() {
	var ctrl *gomock.Controller
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Parallel", func() {
		It("should execute all functions and collect their errors", func() {
			var (
				err2 = mkError(2)
				err3 = mkError(3)

				f1 = mock.NewMockFunc(ctrl)
				f2 = mock.NewMockFunc(ctrl)
				f3 = mock.NewMockFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(ctx)
			f2.EXPECT().Call(ctx).Return(err2)
			f3.EXPECT().Call(ctx).Return(err3)

			err := Parallel(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err2, err3))
		})
	})

	Describe("ParallelCancelOnError", func() {
		It("should run the functions and cancel all if one of them errors", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockFunc(ctrl)
				f2   = mock.NewMockFunc(ctrl)
				f3   = mock.NewMockFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(err1)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnError)

			err := ParallelCancelOnError(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1, context.Canceled, context.Canceled))
		})

		It("should run the functions and return no error if all succeed", func() {
			var (
				f1 = mock.NewMockFunc(ctrl)
				f2 = mock.NewMockFunc(ctrl)
				f3 = mock.NewMockFunc(ctrl)

				ctx = context.TODO()
			)
			f1.EXPECT().Call(gomock.Any())
			f2.EXPECT().Call(gomock.Any())
			f3.EXPECT().Call(gomock.Any())

			Expect(ParallelCancelOnError(ctx, f1.Call, f2.Call, f3.Call)).NotTo(HaveOccurred())
		})
	})

	Describe("Sequence", func() {
		It("should run the functions one after another", func() {
			var (
				f1  = mock.NewMockFunc(ctrl)
				f2  = mock.NewMockFunc(ctrl)
				f3  = mock.NewMockFunc(ctrl)
				ctx = context.TODO()
			)

			gomock.InOrder(
				f1.EXPECT().Call(ctx),
				f2.EXPECT().Call(ctx),
				f3.EXPECT().Call(ctx),
			)

			Expect(Sequence(ctx, f1.Call, f2.Call, f3.Call)).To(Succeed())
		})

		It("should exit with the first error encountered", func() {
			var (
				err2 = mkError(2)
				f1   = mock.NewMockFunc(ctrl)
				f2   = mock.NewMockFunc(ctrl)
				f3   = mock.NewMockFunc(ctrl)
				ctx  = context.TODO()
			)

			gomock.InOrder(
				f1.EXPECT().Call(ctx),
				f2.EXPECT().Call(ctx).Return(err2),
			)

			err := Sequence(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeIdenticalTo(err2))
		})
	})

	Describe("Race", func() {
		It("should return the result of the first function and cancel the others", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockFunc(ctrl)
				f2   = mock.NewMockFunc(ctrl)
				f3   = mock.NewMockFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(err1)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnError)

			err := Race(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeIdenticalTo(err1))
		})
	})

	Describe("ParallelString", func() {
		It("should run all computations, returning all errors and results", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockStringFunc(ctrl)
				f2   = mock.NewMockStringFunc(ctrl)
				f3   = mock.NewMockStringFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return("", err1)
			f2.EXPECT().Call(gomock.Any()).Return("foo", nil)
			f3.EXPECT().Call(gomock.Any()).Return("bar", nil)

			res, err := ParallelString(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1))
			Expect(res).To(ConsistOf("foo", "bar"))
		})
	})

	Describe("ParallelStringCancelOnError", func() {
		It("should run all computations, cancelling them when an error occurs", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockStringFunc(ctrl)
				f2   = mock.NewMockStringFunc(ctrl)
				f3   = mock.NewMockStringFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return("", err1)
			f2.EXPECT().Call(gomock.Any()).Return("foo", nil)
			f3.EXPECT().Call(gomock.Any()).Return("bar", nil)

			res, err := ParallelStringCancelOnError(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1))
			Expect(res).To(ConsistOf("foo", "bar"))
		})
	})

	Describe("RaceString", func() {
		It("should run all computations, returning as soon as one of them finishes", func() {
			var (
				f1 = mock.NewMockStringFunc(ctrl)
				f2 = mock.NewMockStringFunc(ctrl)
				f3 = mock.NewMockStringFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return("foo", nil)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnStringError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnStringError)

			res, err := RaceString(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("foo"))
		})
	})

	Describe("ParallelInt", func() {
		It("should run all computations, returning all errors and results", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockIntFunc(ctrl)
				f2   = mock.NewMockIntFunc(ctrl)
				f3   = mock.NewMockIntFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(0, err1)
			f2.EXPECT().Call(gomock.Any()).Return(1, nil)
			f3.EXPECT().Call(gomock.Any()).Return(2, nil)

			res, err := ParallelInt(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1))
			Expect(res).To(ConsistOf(1, 2))
		})
	})

	Describe("ParallelIntCancelOnError", func() {
		It("should run all computations, cancelling them when an error occurs", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockIntFunc(ctrl)
				f2   = mock.NewMockIntFunc(ctrl)
				f3   = mock.NewMockIntFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(0, err1)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnIntError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnIntError)

			res, err := ParallelIntCancelOnError(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1, context.Canceled, context.Canceled))
			Expect(res).To(BeEmpty())
		})
	})

	Describe("RaceInt", func() {
		It("should run all computations, returning as soon as one of them finishes", func() {
			var (
				f1 = mock.NewMockIntFunc(ctrl)
				f2 = mock.NewMockIntFunc(ctrl)
				f3 = mock.NewMockIntFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(1, nil)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnIntError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnIntError)

			res, err := RaceInt(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(1))
		})
	})

	Describe("ParallelBool", func() {
		It("should run all computations, returning all errors and results", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockBoolFunc(ctrl)
				f2   = mock.NewMockBoolFunc(ctrl)
				f3   = mock.NewMockBoolFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(false, err1)
			f2.EXPECT().Call(gomock.Any()).Return(false, nil)
			f3.EXPECT().Call(gomock.Any()).Return(true, nil)

			res, err := ParallelBool(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1))
			Expect(res).To(ConsistOf(false, true))
		})
	})

	Describe("ParallelBoolCancelOnError", func() {
		It("should run all computations, cancelling them when an error occurs", func() {
			var (
				err1 = mkError(1)
				f1   = mock.NewMockBoolFunc(ctrl)
				f2   = mock.NewMockBoolFunc(ctrl)
				f3   = mock.NewMockBoolFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(false, err1)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnBoolError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnBoolError)

			res, err := ParallelBoolCancelOnError(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err1, context.Canceled, context.Canceled))
			Expect(res).To(BeEmpty())
		})
	})

	Describe("RaceBool", func() {
		It("should run all computations, returning as soon as one of them finishes", func() {
			var (
				f1 = mock.NewMockBoolFunc(ctrl)
				f2 = mock.NewMockBoolFunc(ctrl)
				f3 = mock.NewMockBoolFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(true, nil)
			f2.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnBoolError)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnBoolError)

			res, err := RaceBool(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("RaceCond", func() {
		It("should run all computations, returning as soon as one of them returns no error and true", func() {
			var (
				f1 = mock.NewMockBoolFunc(ctrl)
				f2 = mock.NewMockBoolFunc(ctrl)
				f3 = mock.NewMockBoolFunc(ctrl)

				ctx = context.TODO()
			)

			f1.EXPECT().Call(gomock.Any()).Return(true, nil)
			f2.EXPECT().Call(gomock.Any()).Return(false, nil)
			f3.EXPECT().Call(gomock.Any()).DoAndReturn(waitForContextToErrorAndReturnBoolError)

			res, err := RaceCond(ctx, f1.Call, f2.Call, f3.Call)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})
})
