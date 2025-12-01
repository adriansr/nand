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
		return &loadFile{}, err
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

func (lf *loadFile) build() (*Project, error) {
	var proj Project
	proj.Name = lf.Name
	knownComponents := make(map[string]types.Component, len(lf.Components)+1)
	for _, component := range lf.Components {
		if _, ok := knownComponents[component.Name]; ok {
			return nil, fmt.Errorf("duplicate component name: %s", component.Name)
		}
	}

}
