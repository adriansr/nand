package types

type ChangeContext struct {
	changed map[Component]struct{}
}

func (c *ChangeContext) Pending() int {
	return len(c.changed)
}

func (c *ChangeContext) SetInput(pin *Pin, val BitVal) {
	if pin.IsSet() && pin.Value() == val {
		return
	}
	pin.Set(val)
	if c.changed == nil {
		c.changed = make(map[Component]struct{})
	}
	c.changed[pin.Ref()] = struct{}{}
}

func (c *ChangeContext) SetOutput(pin *Pin, val BitVal) {
	if pin.IsSet() && pin.Value() == val {
		return
	}
	pin.Set(val)
	if c.changed == nil {
		c.changed = make(map[Component]struct{})
	}

	for _, dest := range pin.Consumers() {
		if dest.Value() == val {
			continue
		}
		dest.Set(val)
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
