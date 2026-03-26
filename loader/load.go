package loader

import (
	"errors"
	"fmt"
	"nand/tester"
	"nand/types"
	"strings"

	"gopkg.in/yaml.v3"
)

var Sample string = `
name: builtin-sample
components:
  - name: inv
    internals:
      inverter: nand
    inputs:
     - name: in
       to:
         - inverter.in0
         - inverter.in1
    outputs:
    - name: out
      from: inverter.out
    test:
    - [0, 1]
    - [1, 0]
  - name: xor
    internals:
      nand_1: nand
      nand_2: nand
      nand_3: nand
      nand_4: nand
    connections:
      - from: nand_2.out
        to: nand_1.in1
      - from: nand_2.out
        to: nand_3.in0
      - from: nand_1.out
        to: nand_4.in0
      - from: nand_3.out
        to: nand_4.in1
    inputs:
      - name: a
        to:
        - nand_1.in0
        - nand_2.in0
      - name: b
        to:
        - nand_2.in1
        - nand_3.in1
    outputs:
      - name: out
        from: nand_4.out
    test:
      - [0, 0, 0]
      - [0, 1, 1]
      - [1, 0, 1]
      - [1, 1, 0]
  - name: sr_latch
    internals:
      nand_left: nand
      nand_right: nand
    connections:
      - from: nand_left.out
        to: nand_right.in0
      - from: nand_right.out
        to: nand_left.in1
    inputs:
      - name: s
        to: nand_left.in0
      - name: c
        to: nand_right.in1
    outputs:
      - name: out
        from: nand_left.out
    test:
      - [0, 1, 1]
      - [1, 1, 1]
      - [1, 1, 1]
      - [0, 1, 1]
      - [1, 0, 0]
      - [1, 1, 0]
`

type Project struct {
	Name       string
	Components []types.Component
}

type loadFile struct {
	Name       string
	Components []loadedComponent
}

func Load(contents []byte) (*Project, error) {
	var loaded loadFile
	if err := yaml.Unmarshal(contents, &loaded); err != nil {
		return nil, err
	}
	proj, err := loaded.build()
	if err != nil {
		return nil, err
	}
	return proj, nil
}

type loadedComponent struct {
	Name        string
	Internals   map[string]string
	Connections []loadedConnection
	Inputs      []inputMapping
	Outputs     []outputMapping
	Test        [][]uint8
}

type loadedConnection struct {
	From pinRef
	To   MaybeList[pinRef]
}

type inputMapping struct {
	Name string
	To   MaybeList[pinRef]
}

type outputMapping struct {
	Name string
	From pinRef
}

type pinRef [2]string

var errBadFormat = errors.New("bad pinref format. Expected name.pin")

func (p *pinRef) UnmarshalYAML(node *yaml.Node) error {
	var fullRef string
	if err := node.Decode(&fullRef); err != nil {
		return fmt.Errorf("decoding pinRef at %d: %w", node.Line, err)
	}
	parts := strings.SplitN(fullRef, ".", 3)
	if len(parts) != 2 {
		return fmt.Errorf("decoding pinRef at %d: %w", node.Line, errBadFormat)
	}
	p[0] = parts[0]
	p[1] = parts[1]
	return nil
}

func (p *pinRef) String() string {
	return fmt.Sprintf("%s.%s", p[0], p[1])
}

type MaybeList[T any] []T

func (list *MaybeList[T]) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		var t T
		if err := node.Decode(&t); err != nil {
			return fmt.Errorf("decoding MaybeList as single item: %w", err)
		}
		*list = []T{t}
		return nil
	}
	var tt []T
	if err := node.Decode(&tt); err != nil {
		return fmt.Errorf("decoding MaybeList as sequence: %w", err)
	}
	*list = tt
	return nil
}

type buildFn func(name string) types.Component

type buildInternals struct {
	known     map[string]buildFn
	internals map[string]types.Component
}

func (lf *loadFile) build() (*Project, error) {
	var (
		proj  Project
		build buildInternals
	)
	proj.Name = lf.Name
	build.known = make(map[string]buildFn, len(lf.Components)+1)
	build.known["nand"] = func(name string) types.Component { return types.NewNand(name, 2) }

	for _, component := range lf.Components {
		if _, ok := build.known[component.Name]; ok {
			return nil, fmt.Errorf("duplicate component name: %s", component.Name)
		}
		bf, err := component.build(&build)
		if err != nil {
			return nil, fmt.Errorf("loading component name: %s: %w", component.Name, err)
		}
		build.known[component.Name] = bf
	}
	proj.Components = make([]types.Component, 0, len(lf.Components))
	for _, component := range lf.Components {
		proj.Components = append(proj.Components, build.known[component.Name](component.Name))
	}
	return &proj, nil
}

type testRunner struct {
	name    string
	inputs  []*types.Pin
	outputs []*types.Pin
	// We keep references to the internal components so they stay alive or just for structure.
	internals map[string]types.Component
}

func (tr *testRunner) Name() string {
	return tr.name
}

func (tr *testRunner) Inputs() []*types.Pin {
	return tr.inputs
}

func (tr *testRunner) Outputs() []*types.Pin {
	return tr.outputs
}

func (tr *testRunner) Update(ctx *types.ChangeContext) {
	for _, in := range tr.inputs {
		if in.IsSet() {
			val := in.Value()
			for _, dest := range in.Consumers() {
				ctx.SetInput(dest, val)
			}
		}
	}
}

func (lc *loadedComponent) build(build *buildInternals) (buildFn, error) {
	// 1. Run the test with a throwaway instance
	tr, err := lc.instantiate(lc.Name, build.known)
	if err != nil {
		return nil, err
	}
	if len(lc.Test) > 0 {
		tester.Test(tr, lc.Test)
	}

	// 2. Return the build function that produces new instances
	return func(name string) types.Component {
		comp, _ := lc.instantiate(name, build.known)
		return comp
	}, nil
}

func (lc *loadedComponent) instantiate(name string, known map[string]buildFn) (*testRunner, error) {
	build := &buildInternals{
		known:     known,
		internals: make(map[string]types.Component, len(lc.Internals)),
	}
	for intName, intType := range lc.Internals {
		fn, ok := build.known[intType]
		if !ok {
			return nil, fmt.Errorf("unknown component type: %s in %s internals for %s", intType, intName, lc.Name)
		}
		build.internals[intName] = fn(intType)
	}
	for _, conn := range lc.Connections {
		out, err := build.lookupOutput(conn.From)
		if err != nil {
			return nil, err
		}
		for _, to := range conn.To {
			inp, err := build.lookupInput(to)
			if err != nil {
				return nil, err
			}
			if !out.Connect(inp) {
				return nil, fmt.Errorf("already connected: %s", to)
			}
		}
	}
	tr := &testRunner{name: name, internals: build.internals}
	var inputs []*types.Pin = make([]*types.Pin, len(lc.Inputs))
	for idx, inp := range lc.Inputs {
		inputs[idx] = types.NewPin(inp.Name, tr)
		for _, to := range inp.To {
			target, err := build.lookupInput(to)
			if err != nil {
				return nil, err
			}
			if !inputs[idx].Connect(target) {
				return nil, fmt.Errorf("already connected: %s", to)
			}
		}
	}
	tr.inputs = inputs
	var outputs []*types.Pin = make([]*types.Pin, len(lc.Outputs))
	for idx, outp := range lc.Outputs {
		outputs[idx] = types.NewPin(outp.Name, tr)
		source, err := build.lookupOutput(outp.From)
		if err != nil {
			return nil, err
		}
		if !source.Connect(outputs[idx]) {
			return nil, fmt.Errorf("already connected: %s", outp.From)
		}
	}
	tr.outputs = outputs
	return tr, nil
}

func (bi *buildInternals) lookupInput(addr pinRef) (*types.Pin, error) {
	cc, found := bi.internals[addr[0]]
	if !found {
		return nil, fmt.Errorf("pinref %s: internal %s not found", addr, addr[0])
	}
	for _, inp := range cc.Inputs() {
		if inp.Name() == addr[1] {
			return inp, nil
		}
	}
	return nil, fmt.Errorf("pinref %s: input %s not found", addr, addr[1])
}

func (bi *buildInternals) lookupOutput(addr pinRef) (*types.Pin, error) {
	cc, found := bi.internals[addr[0]]
	if !found {
		return nil, fmt.Errorf("pinref %s: internal %s not found", addr.String(), addr[0])
	}
	for _, inp := range cc.Outputs() {
		if inp.Name() == addr[1] {
			return inp, nil
		}
	}
	return nil, fmt.Errorf("pinref %s: output %s not found", addr.String(), addr[1])
}
