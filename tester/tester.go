package tester

import (
	"fmt"

	"nand/types"
)

type testOptions struct {
	debug bool
}

type TestOpt func(*testOptions)

func WithDebug(ops *testOptions) {
	ops.debug = true
}

func Test[T comparable](c types.BaseComponent, tt [][]T, opts ...TestOpt) {
	var options testOptions
	for _, opt := range opts {
		opt(&options)
	}

	inp, outp := c.Inputs(), c.Outputs()
	for tcIdx, tc := range tt {
		var ctx types.ChangeContext
		if len(tc) != len(inp)+len(outp) {
			panic("inputs must have the same length")
		}
		for idx := range len(inp) {
			ctx.SetInput(inp[idx], toVal(tc[idx]))
		}

		iters := 0
		if options.debug {
			fmt.Printf("\n")
		}
		for !ctx.Done() {
			if options.debug {
				iters++
				fmt.Printf("-- debug: iters=%d len=%d\r", iters, ctx.Pending())
			}
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
