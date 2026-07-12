package main

import "time"

type Controller struct {
	Engine          *Engine
	MouseController *MouseController
	MouseTracker    *MouseTracker
	StopChan        chan struct{}
}

func (c *Controller) Run() {
	if c.StopChan == nil {
		c.StopChan = make(chan struct{})
	}

	const sleepInterval = 2 * time.Millisecond
	const epsilon = 2.0

	for {
		select {
		case <-c.StopChan:
			return
		default:
			c.Engine.Update()

			pos := c.Engine.Position
			before := c.MouseTracker.GetPosition()
			c.MouseController.Move(&pos)

			for i := 0; i < 20; i++ {
				realPos := c.MouseTracker.GetPosition()
				if closeEnough(pos, realPos, epsilon) {
					break
				}
				if !closeEnough(before, realPos, epsilon) && !closeEnough(pos, realPos, epsilon) {
					return
				}
				time.Sleep(sleepInterval)
			}

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
