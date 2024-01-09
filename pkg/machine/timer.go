package machine

type TickProvider interface {
	GetCurrentTick() int
}

type Clock struct {
	currentTick int
}

func (c *Clock) GetCurrentTick() int {
	return c.currentTick
}
