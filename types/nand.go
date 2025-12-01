package types

import (
	"fmt"
)

type Nand struct {
	named

	inputs []WritePin
	output OutPin
}

func NewNand(name string, numInputs int) *Nand {
	n := &Nand{
		named: newNamed(name),
	}
	n.inputs = make([]WritePin, numInputs)
	for i := range n.inputs {
		n.inputs[i] = newPin(fmt.Sprintf("in%d", i), n)
	}
	n.output = &outPin{
		namedPin: *newPin("out0", n),
	}
	return n
}

func (n *Nand) Inputs() []WritePin {
	return n.inputs
}

func (n *Nand) Outputs() []OutPin {
	return []OutPin{
		n.output,
	}
}

func (n *Nand) Update(ctx Ctx) {
	for _, v := range n.inputs {
		if !v.Value() {
			ctx.SetOutput(n.output, true)
			return
		}
	}
	ctx.SetOutput(n.output, false)
}

type namedPin struct {
	named
	ref Component
	v   BitVal
}

func newPin(name string, ref Component) *namedPin {
	return &namedPin{
		named: newNamed(name),
		ref:   ref,
	}
}

func (n *namedPin) Value() BitVal {
	return n.v
}

func (n *namedPin) Ref() Component {
	return n.ref
}

func (n *namedPin) SetRaw(v BitVal) {
	n.v = v
}

type named string

func (n named) Name() string {
	return string(n)
}

func newNamed(name string) named {
	return named(name)
}
