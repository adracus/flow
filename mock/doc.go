//go:generate mockgen -destination=funcs.go -package mock github.com/adracus/flow/mock Func,StringFunc,IntFunc,BoolFunc,SubmitFunc
//go:generate mockgen -destination=mocks.go -package mock github.com/adracus/flow Executor
package mock

import "context"

type SubmitFunc interface {
	Call()
}

type Func interface {
	Call(context.Context) error
}

type StringFunc interface {
	Call(context.Context) (string, error)
}

type IntFunc interface {
	Call(context.Context) (int, error)
}

type BoolFunc interface {
	Call(context.Context) (bool, error)
}
