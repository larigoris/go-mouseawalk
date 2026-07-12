package main

type Engine struct { // Хранит нынешнюю и следующую позицию мыши, а также высоту и длину экрана
	Position Vector
	Velocity Vector
	Width    int
	Height   int
}

func (e *Engine) Update() {
	const substeps = 4
	for i := 0; i < substeps; i++ {
		dx := e.Velocity.X / float64(substeps)
		dy := e.Velocity.Y / float64(substeps)

		nextX := e.Position.X + dx
		nextY := e.Position.Y + dy

		if nextX >= float64(e.Width) || nextX < 0 {
			e.Velocity.X = -e.Velocity.X
			dx = e.Velocity.X / float64(substeps)
		}
		if nextY >= float64(e.Height) || nextY < 0 {
			e.Velocity.Y = -e.Velocity.Y
			dy = e.Velocity.Y / float64(substeps)
		}

		e.Position.X += dx
		e.Position.Y += dy
	}
}
