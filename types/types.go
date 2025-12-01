package types

type BitVal = bool

type Named interface {
	Name() string
}

type Pin interface {
	Named

	Value() BitVal
	Ref() Component
}

type WritePin interface {
	Pin

	SetRaw(BitVal)
}

type OutPin interface {
	Pin

	Connect(WritePin) // TODO: Detect and fail if already connected
	Consumers() []WritePin
	SetRaw(v BitVal)
	IsSet() bool
}

type Component interface {
	Named

	Inputs() []WritePin
	Outputs() []OutPin
	Update(Ctx)
}

type Ctx interface {
	SetOutput(pin OutPin, val BitVal)
}

type Runtime interface {
	Done() bool
	Next() Component
}

type outPin struct {
	namedPin
	set     bool
	targets []WritePin
}

func (out *outPin) Connect(wp WritePin) {
	for idx := range out.targets {
		if out.targets[idx] == wp {
			return
		}
	}
	out.targets = append(out.targets, wp)
}

func (out *outPin) Consumers() []WritePin {
	return out.targets
}

func (out *outPin) IsSet() bool {
	return out.set
}
func (out *outPin) SetRaw(v BitVal) {
	out.set = true
	out.v = v
}
