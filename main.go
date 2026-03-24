package main

import (
	"fmt"

	"nand/loader"
	"nand/tester"
	"nand/types"
)

func main() {
	nand := types.NewNand("nand", 2)
	tester.Test(nand.Inputs(),
		nand.Outputs(),
		[][]byte{
			{0, 0, 1},
			{0, 1, 1},
			{1, 0, 1},
			{1, 1, 0},
		})
	nl := types.NewNand("left", 2)
	nr := types.NewNand("right", 2)
	nl.Outputs()[0].Connect(nr.Inputs()[0])
	nr.Outputs()[0].Connect(nl.Inputs()[1])

	tester.Test(
		[]*types.Pin{nl.Inputs()[0], nr.Inputs()[1]},
		nl.Outputs(),
		[][]byte{
			{0, 1, 1},
			{1, 0, 0},
			{1, 0, 0},
			{1, 1, 0},
			{0, 1, 1},
			{1, 1, 1},
		})
	proj, err := loader.Load([]byte(loader.Sample))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Loaded project: %v\n", proj)
}
