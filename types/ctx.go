package types

type ChangeContext struct {
	changed map[Component]struct{}
}

func (c *ChangeContext) SetInput(pin WritePin, val BitVal) {
	pin.SetRaw(val)
	if c.changed == nil {
		c.changed = make(map[Component]struct{})
	}
	c.changed[pin.Ref()] = struct{}{}
}

func (c *ChangeContext) SetOutput(pin OutPin, val BitVal) {
	if pin.IsSet() && pin.Value() == val {
		return
	}
	pin.SetRaw(val)
	if c.changed == nil {
		c.changed = make(map[Component]struct{})
	}

	for _, dest := range pin.Consumers() {
		if dest.Value() == val {
			continue
		}
		dest.SetRaw(val)
		c.changed[dest.Ref()] = struct{}{}
	}
}

func (c *ChangeContext) Done() bool {
	return len(c.changed) == 0
}

func (c *ChangeContext) Next() Component {
	var k Component
	for k = range c.changed {
		break
	}
	delete(c.changed, k)
	return k
}
