package main

import "time"

type Controller struct {
	Engine          *Engine
	MouseController *MouseController
	MouseTracker    *MouseTracker
	StopChan        chan struct{}
}

func (c *Controller) Run() {
	for {
		select {
		case <-c.StopChan:
			return
		default:
			c.Engine.Update()

			pos := c.Engine.Position
			c.MouseController.Move(&pos)

			realPos := c.MouseTracker.Position

			if pos != realPos {
				return
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}
