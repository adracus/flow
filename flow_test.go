package flow_test

import (
	"context"
	. "flow"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestFlow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flow Suite")
}

func trackExecution(f Func) (Func, *bool) {
	executed := false
	return func(ctx context.Context) error {
        executed = true
        return f(ctx)
	}, &executed
}

func mkFunc(err error) Func {
	return func(context.Context) error {
		return err
	}
}

type seqFuncBuilder struct {
	cur int
	total int
}

func NewSeqFuncBuilder() *seqFuncBuilder {
	return &seqFuncBuilder{}
}

func (s *seqFuncBuilder) mkFunc(err error) Func {
	expected := s.total
	s.total++
    return func(context.Context) error {
		Expect(s.cur).To(Equal(expected), "function not executed in correct sequence: current %d, expected %d", s.cur, expected)
        s.cur++
		return err
	}
}

func mkError(id int) error {
	return fmt.Errorf("error %d", id)
}

var _ = Describe("Flow", func() {
	Describe("Parallel", func() {
        It("should execute all functions and collect their errors", func() {
			var (
				err2 = mkError(2)
				err3 = mkError(3)

				f1, e1 = trackExecution(mkFunc(nil))
				f2, e2 = trackExecution(mkFunc(err2))
				f3, e3 = trackExecution(mkFunc(err3))
			)

			err := Parallel(context.TODO(), f1, f2, f3)
			Expect(err).To(HaveOccurred())
			Expect(Errors(err)).To(ConsistOf(err2, err3))
			Expect(*e1).To(BeTrue())
			Expect(*e2).To(BeTrue())
			Expect(*e3).To(BeTrue())
		})
	})

	Describe("Sequence", func() {
		It("should run the functions one after another", func() {
			var (
				b = NewSeqFuncBuilder()
				f1, e1 = trackExecution(b.mkFunc(nil))
				f2, e2 = trackExecution(b.mkFunc(nil))
				f3, e3 = trackExecution(b.mkFunc(nil))
			)

            Expect(Sequence(context.TODO(), f1, f2, f3)).To(Succeed())
            Expect(*e1).To(BeTrue())
			Expect(*e2).To(BeTrue())
			Expect(*e3).To(BeTrue())
		})

		It("should exit with the first error encountered", func() {
			var (
				b = NewSeqFuncBuilder()
				err2 = mkError(2)
				f1, e1 = trackExecution(b.mkFunc(nil))
				f2, e2 = trackExecution(b.mkFunc(err2))
				f3, e3 = trackExecution(b.mkFunc(nil))
			)

			err := Sequence(context.TODO(), f1, f2, f3)
            Expect(err).To(HaveOccurred())
			Expect(err).To(BeIdenticalTo(err2))
			Expect(*e1).To(BeTrue())
			Expect(*e2).To(BeTrue())
			Expect(*e3).To(BeFalse())
		})
	})

	Describe("Race", func() {
        It("should return the result of the first function and cancel the others", func() {
        	var (
        		err1 = mkError(1)
                f1, e1 = trackExecution(mkFunc(err1))
                f2, e2 = trackExecution(func(ctx context.Context) error {
                    <-ctx.Done()
                    return nil
				})
                f3, e3 = trackExecution(func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				})
			)

            err := Race(context.TODO(), f1, f2, f3)
            Expect(err).To(HaveOccurred())
            Expect(err).To(BeIdenticalTo(err1))
			Expect(*e1).To(BeTrue())
			Expect(*e2).To(BeTrue())
			Expect(*e3).To(BeTrue())
		})
	})
})
