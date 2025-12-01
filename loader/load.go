package loader

import (
	"errors"
	"fmt"
	"strings"

	"nand/types"

	"gopkg.in/yaml.v3"
)

var Sample string = `
name: builtin-sample
components:
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
      - name: nout
        from: nand_right.out
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
	Test        [][]uint8
}

type loadedConnection struct {
	From, To pinRef
}

type inputMapping struct {
	Name string
	To   pinRef
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
	return nil, nil
}

func (lc *loadedComponent) build(build *buildInternals) (buildFn, error) {
	build.internals = make(map[string]types.Component, len(lc.Internals))
	for intName, intType := range lc.Internals {
		if _, ok := build.internals[intName]; ok {
			return nil, fmt.Errorf("duplicate internal name: %s", intName)
		}
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
		inp, err := build.lookupInput(conn.To)
		if err != nil {
			return nil, err
		}
		out.Connect(inp)
	}
	return nil, nil
}

func (bi *buildInternals) lookupInput(addr pinRef) (types.WritePin, error) {
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

func (bi *buildInternals) lookupOutput(addr pinRef) (types.OutPin, error) {
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
