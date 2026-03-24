package tester

import (
	"fmt"

	"nand/types"
)

func Test[T comparable](inp []*types.Pin, outp []*types.Pin, tt [][]T) {
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
