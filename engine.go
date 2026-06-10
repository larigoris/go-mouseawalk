package main

type Engine struct { // Хранит нынешнюю и следующую позицию мыши, а также высоту и длину экрана
	Position Vector
	Velocity Vector
	Width    int
	Height   int
}

func (e *Engine) Update() {
	nextX := e.Position.X + e.Velocity.X
	nextY := e.Position.Y + e.Velocity.Y

	if nextX >= e.Width || nextX < 0 {
		e.Velocity.X = -e.Velocity.X
	}
	if nextY >= e.Height || nextY < 0 {
		e.Velocity.Y = -e.Velocity.Y
	}
	e.Position.X += e.Velocity.X
	e.Position.Y += e.Velocity.Y
}
