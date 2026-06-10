package main

import "github.com/go-vgo/robotgo"

type MouseController struct {
	Position Vector
	Velocity Vector
}

func (m *MouseController) GetPosition() Vector {
	m.Position.X, m.Position.Y = robotgo.Location()
	return m.Position
}

func (m *MouseController) Move(pos *Vector) {
	robotgo.Move(pos.X, pos.Y)
}
