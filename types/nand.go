package types

import (
	"fmt"
)

type Nand struct {
	named

	inputs []*Pin
	output *Pin
}

func NewNand(name string, numInputs int) *Nand {
	n := &Nand{
		named: newNamed(name),
	}
	n.inputs = make([]*Pin, numInputs)
	for i := range n.inputs {
		n.inputs[i] = NewPin(fmt.Sprintf("in%d", i), n)
	}
	n.output = NewPin("out", n)
	return n
}

func (n *Nand) Inputs() []*Pin {
	return n.inputs
}

func (n *Nand) Outputs() []*Pin {
	return []*Pin{
		n.output,
	}
}

func (n *Nand) Update(ctx *ChangeContext) {
	result := false
	for _, v := range n.inputs {
		if !v.Value() {
			result = true
			break
		}
	}
	ctx.SetOutput(n.output, result)
}
