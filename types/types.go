package types

type BitVal = bool

type Named interface {
	Name() string
}

/*
	type Pin interface {
		Named

		Ref() Component
		SetRaw(BitVal)
		Value() BitVal
		Connect(Pin) bool
		Consumers() []Pin
		IsSet() bool
	}
*/
type BaseComponent interface {
	Named

	Inputs() []*Pin
	Outputs() []*Pin
}

type Component interface {
	BaseComponent

	Update(ctx *ChangeContext)
}

//type Ctx interface {
//	SetOutput(pin Pin, val BitVal)
//}

type Runtime interface {
	Done() bool
	Next() Component
	Pending() int
}

type named string

func (n named) Name() string {
	return string(n)
}

func newNamed(name string) named {
	return named(name)
}

type Pin struct {
	named
	ref     Component
	source  *Pin
	targets []*Pin
	v       BitVal
	set     bool
}

func NewPin(name string, ref Component) *Pin {
	return &Pin{
		named: newNamed(name),
		ref:   ref,
	}
}

func (p *Pin) Connect(target *Pin) bool {
	for _, t := range p.targets {
		if t == target {
			return true
		}
	}

	prev := target.source
	target.source = p
	if prev != nil {
		return false
	}

	p.targets = append(p.targets, target)
	return true
}

func (p *Pin) Value() BitVal {
	return p.v
}

func (p *Pin) Ref() Component {
	return p.ref
}

func (p *Pin) Set(v BitVal) {
	p.set = true
	p.v = v
}

func (p *Pin) Consumers() []*Pin {
	return p.targets
}

func (p *Pin) IsSet() bool {
	return p.set
}

type WrappingComponent struct {
	named

	inputs  []Pin
	outputs []Pin
}

func (wc *WrappingComponent) Inputs() []Pin {
	return wc.inputs
}

func (wc *WrappingComponent) Outputs() []Pin {
	return wc.outputs
}

func (wc *WrappingComponent) Update() {
	panic("implement me")
}
