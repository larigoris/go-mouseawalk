package main

import "github.com/go-vgo/robotgo"

func main() {
	width, height := robotgo.GetScreenSize()
	mouse := &MouseController{}

	engine := &Engine{
		Position: mouse.GetPosition(),

		Velocity: Vector{
			X: 1,
			Y: 1,
		},

		Width:  width,
		Height: height,
	}

	tracker := &MouseTracker{}
	tracker.Start()

	controller := &Controller{
		Engine:          engine,
		MouseController: mouse,
		MouseTracker:    tracker,
	}

	controller.Run()
}
