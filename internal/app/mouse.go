package app

import (
	"math"
	"sync"

	"github.com/go-vgo/robotgo"
)

type MouseController struct {
	mu       sync.Mutex
	Position Vector
	Velocity Vector
}

func (m *MouseController) GetPosition() Vector {
	x, y := robotgo.Location()
	m.mu.Lock()
	m.Position.X = float64(x)
	m.Position.Y = float64(y)
	pos := m.Position
	m.mu.Unlock()
	return pos
}

func (m *MouseController) Move(pos *Vector) {
	robotgo.Move(int(math.Round(pos.X)), int(math.Round(pos.Y)))
	m.mu.Lock()
	m.Position = *pos
	m.mu.Unlock()
}
