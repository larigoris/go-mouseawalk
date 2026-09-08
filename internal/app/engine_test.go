package app

import "testing"

const (
	testWidth  = 1000
	testHeight = 800
)

func newEngine(x, y, vx, vy float64) *Engine {
	return &Engine{
		Position: Vector{X: x, Y: y},
		Velocity: Vector{X: vx, Y: vy},
		Width:    testWidth,
		Height:   testHeight,
	}
}

func TestUpdateMovesByVelocity(t *testing.T) {
	e := newEngine(100, 200, 5.5, -3)
	e.Update()

	if e.Position.X != 105.5 || e.Position.Y != 197 {
		t.Errorf("позиция = %v, ожидалось (105.5, 197)", e.Position)
	}
}

func TestUpdateReflectsAtEdges(t *testing.T) {
	cases := []struct {
		name           string
		x, y, vx, vy   float64
		wantVX, wantVY float64
	}{
		{"правый край", testWidth - 2, 400, 5.5, 5.5, -5.5, 5.5},
		{"левый край", 2, 400, -5.5, 5.5, 5.5, 5.5},
		{"нижний край", 500, testHeight - 2, 5.5, 5.5, 5.5, -5.5},
		{"верхний край", 500, 2, 5.5, -5.5, 5.5, 5.5},
		{"угол", testWidth - 2, testHeight - 2, 5.5, 5.5, -5.5, -5.5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newEngine(c.x, c.y, c.vx, c.vy)
			e.Update()

			if e.Velocity.X != c.wantVX || e.Velocity.Y != c.wantVY {
				t.Errorf("скорость = %v, ожидалось (%v, %v)", e.Velocity, c.wantVX, c.wantVY)
			}
		})
	}
}

// TestUpdateKeepsPositionInsideBounds — начав внутри области, движок обязан
// оставаться внутри неё сколько угодно долго.
func TestUpdateKeepsPositionInsideBounds(t *testing.T) {
	e := newEngine(500, 400, 5.5, 5.5)

	for i := 0; i < 20000; i++ {
		e.Update()

		if e.Position.X < 0 || e.Position.X >= testWidth {
			t.Fatalf("шаг %d: X=%v вышел за [0, %d)", i, e.Position.X, testWidth)
		}
		if e.Position.Y < 0 || e.Position.Y >= testHeight {
			t.Fatalf("шаг %d: Y=%v вышел за [0, %d)", i, e.Position.Y, testHeight)
		}
	}
}

// TestUpdateDoesNotRecoverFromOutOfBounds фиксирует ИЗВЕСТНЫЙ ДЕФЕКТ.
//
// Update() только меняет знак скорости, но не возвращает точку в границы.
// Если стартовая позиция лежит снаружи (курсор на другом мониторе — например
// на экране с отрицательными координатами), условие выхода за границу
// срабатывает на каждом такте, и точка вечно колеблется в полосе шириной
// в один шаг, так и не вернувшись в область.
//
// Когда отражение исправят на зеркальное с возвратом внутрь, этот тест
// упадёт — и его нужно будет заменить на проверку возврата в границы.
func TestUpdateDoesNotRecoverFromOutOfBounds(t *testing.T) {
	e := newEngine(-500, 400, 5.5, 5.5)

	seen := map[float64]bool{}
	for i := 0; i < 100; i++ {
		e.Update()
		seen[e.Position.X] = true
	}

	if e.Position.X >= 0 {
		t.Fatalf("дефект исправлен: X=%v вернулся в границы — обнови этот тест", e.Position.X)
	}
	if len(seen) > 2 {
		t.Errorf("ожидалось колебание между двумя значениями X, получено %d различных", len(seen))
	}
}

func TestCloseEnough(t *testing.T) {
	cases := []struct {
		name string
		a, b Vector
		want bool
	}{
		{"совпадают", Vector{10, 10}, Vector{10, 10}, true},
		{"ровно на пороге", Vector{10, 10}, Vector{30, 30}, true},
		{"за порогом по X", Vector{10, 10}, Vector{31, 10}, false},
		{"за порогом по Y", Vector{10, 10}, Vector{10, 31}, false},
		{"отрицательная разница", Vector{30, 30}, Vector{10, 10}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := closeEnough(c.a, c.b, epsilon); got != c.want {
				t.Errorf("closeEnough(%v, %v, %v) = %v, ожидалось %v", c.a, c.b, epsilon, got, c.want)
			}
		})
	}
}
