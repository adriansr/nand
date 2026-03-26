package tester

import (
	"fmt"

	"nand/types"
)

func Test[T comparable](c types.BaseComponent, tt [][]T) {
	inp, outp := c.Inputs(), c.Outputs()
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
			panic(fmt.Sprintf("🔴 outputs must have expected value at %s[%d]: %v", c.Name(), tcIdx, tc))
		}
		fmt.Printf("🟢 %s[%d]: %v\n", c.Name(), tcIdx, tc)
	}
}

func toVal[T comparable](t T) types.BitVal {
	var z T
	return t != z
}

func eq[S, T comparable](a S, b T) bool {
	return toVal(a) == toVal(b)
}
