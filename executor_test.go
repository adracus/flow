package flow_test

import (
	"sync"
	"testing"

	"github.com/adracus/flow"
	"github.com/adracus/flow/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flow Suite")
}

var _ = Describe("Executor", func() {
	var ctrl *gomock.Controller
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("LimitingExecutor", func() {
		It("should not submit more functions than it allows", func(done Done) {
			mockEx := mock.NewMockExecutor(ctrl)
			ex := flow.LimitExecutor(2, mockEx)
			ex.Start()

			var (
				f1 = mock.NewMockSubmitFunc(ctrl)
				f2 = mock.NewMockSubmitFunc(ctrl)
				f3 = mock.NewMockSubmitFunc(ctrl)
			)

			mockEx.EXPECT().Submit(gomock.Any()).Times(3).Do(func(f func()) { go f() })

			var wg sync.WaitGroup
			wg.Add(2)

			f1Exec := f1.EXPECT().Call().Do(func() { wg.Done(); wg.Wait() })
			f2Exec := f2.EXPECT().Call().Do(func() { wg.Done(); wg.Wait() })
			f3.EXPECT().Call().After(f1Exec).After(f2Exec).Do(func() { ex.Stop(); close(done) })

			ex.Submit(f1.Call)
			ex.Submit(f2.Call)
			ex.Submit(f3.Call)
		})
	})
})
