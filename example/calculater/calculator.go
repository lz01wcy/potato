package main

import (
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/murang/potato/example/nicepb/nice"
)

type CalculatorImpl struct{}

func (c CalculatorImpl) Init(ctx cluster.GrainContext) {

}

func (c CalculatorImpl) Terminate(ctx cluster.GrainContext) {
}

func (c CalculatorImpl) ReceiveDefault(ctx cluster.GrainContext) {
}

func (c CalculatorImpl) Sum(req *nice.Input, ctx cluster.GrainContext) (*nice.Output, error) {
	return &nice.Output{Result: req.A + req.B}, nil
}
