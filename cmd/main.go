package main

import (
	"GOwalk/internal/app"

	"github.com/go-vgo/robotgo"
)

func main() {
	width, height := robotgo.GetScreenSize()
	mouse := &app.MouseController{}

	engine := &app.Engine{
		Position: mouse.GetPosition(),
		Velocity: app.Vector{
			X: 5.5,
			Y: 5.5,
		},
		Width:  width,
		Height: height,
	}

	tracker := &app.MouseTracker{}
	tracker.Start()

	controller := &app.Controller{
		Engine:          engine,
		MouseController: mouse,
		MouseTracker:    tracker,
	}

	controller.Run()
}
