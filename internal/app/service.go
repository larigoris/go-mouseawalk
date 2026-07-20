package app

import "time"

type Service struct {
	Engine          *Engine
	MouseController *MouseController
	MouseTracker    *MouseTracker
	StopChan        chan struct{}
}

func (c *Service) Run() {
	if c.StopChan == nil {
		c.StopChan = make(chan struct{})
	}

	const sleepInterval = 8 * time.Millisecond
	const epsilon = 8.0

	for {
		select {
		case <-c.StopChan:
			return
		default:
			currentPos := c.MouseController.GetPosition()
			if !closeEnough(currentPos, c.Engine.Position, epsilon) {
				return
			}

			c.Engine.Update()
			pos := c.Engine.Position
			c.MouseController.Move(&pos)

			time.Sleep(sleepInterval)
		}
	}
}

func closeEnough(a, b Vector, epsilon float64) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx <= epsilon && dy <= epsilon
}
