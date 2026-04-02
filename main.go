package main

import (
	"fmt"
	"os"

	"nand/loader"
	"nand/tester"
	"nand/types"
)

// declare a types.Component wrapper that I can create giving it the same inputs and outputs as the wrapped component
type ComponentWrapper struct {
	name    string
	inputs  []*types.Pin
	outputs []*types.Pin
}

func NewComponentWrapper(name string, inputs []*types.Pin, outputs []*types.Pin) *ComponentWrapper {
	return &ComponentWrapper{
		name:    name,
		inputs:  inputs,
		outputs: outputs,
	}
}

func (c *ComponentWrapper) Name() string {
	return c.name
}

func (c *ComponentWrapper) Inputs() []*types.Pin {
	return c.inputs
}

func (c *ComponentWrapper) Outputs() []*types.Pin {
	return c.outputs
}

func main() {
	nand := types.NewNand("nand", 2)
	tester.Test(nand, [][]byte{
		{0, 0, 1},
		{0, 1, 1},
		{1, 0, 1},
		{1, 1, 0},
	})
	nl := types.NewNand("left", 2)
	nr := types.NewNand("right", 2)
	nl.Outputs()[0].Connect(nr.Inputs()[0])
	nr.Outputs()[0].Connect(nl.Inputs()[1])

	ff := NewComponentWrapper("flip-flop", []*types.Pin{nl.Inputs()[0], nr.Inputs()[1]}, nl.Outputs())
	tester.Test(
		ff,
		[][]byte{
			{0, 1, 1},
			{1, 0, 0},
			{1, 0, 0},
			{1, 1, 0},
			{0, 1, 1},
			{1, 1, 1},
		})
	for _, path := range os.Args[1:] {
		fmt.Println("Loading project from:", path)
		proj, err := loader.LoadFile(path)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Loaded project: %v\n", proj)
	}
}
