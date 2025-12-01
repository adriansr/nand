package main

import (
	"fmt"

	"nand/loader"
	"nand/types"
)

func main() {
	load, err := loader.Load([]byte(loader.Sample))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded:\n%+v\n", load)
}

func main2() {
	nand := types.NewNand("nand", 2)
	test(nand.Inputs(),
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

	test(
		[]types.WritePin{nl.Inputs()[0], nr.Inputs()[1]},
		nl.Outputs(),
		[][]byte{
			{0, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
			{0, 1, 1},
			{1, 0, 0},
			{1, 1, 0},
		})
}

func join[T any](a []T, b []T) []T {
	v := make([]T, 0, len(a)+len(b))
	v = append(v, a...)
	v = append(v, b...)
	return v
}

func test[T comparable](inp []types.WritePin, outp []types.OutPin, tt [][]T) {
	for tcIdx, tc := range tt {
		var ctx types.ChangeContext
		if len(tc) != len(inp)+len(outp) {
			panic("inputs must have the same length")
		}
		for idx := range len(inp) {
			ctx.SetInput(inp[idx], toVal(tc[idx]))
		}

		for !ctx.Done() {
			ctx.Next().Update(&ctx)
		}

		if !eq(outp[0].Value(), tc[len(tc)-1]) {
			panic(fmt.Sprintf("outputs must have expected value at [%d]: %v", tcIdx, tc))
		}
		fmt.Printf(":green: [%d]: %v\n", tcIdx, tc)
	}
}

func toVal[T comparable](t T) types.BitVal {
	var z T
	return t != z
}

func eq[S, T comparable](a S, b T) bool {
	return toVal(a) == toVal(b)
}
