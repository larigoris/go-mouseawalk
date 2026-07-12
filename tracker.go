package main

import (
	"sync"

	hook "github.com/robotn/gohook"
)

type MouseTracker struct {
	mu       sync.Mutex
	Position Vector
}

func (t *MouseTracker) Start() {
	evChan := hook.Start()
	go func() {
		for ev := range evChan {
			if ev.Kind == hook.MouseMove {
				t.mu.Lock()
				t.Position.X = float64(ev.X)
				t.Position.Y = float64(ev.Y)
				t.mu.Unlock()
			}
		}
	}()
}

func (t *MouseTracker) GetPosition() Vector {
	t.mu.Lock()
	pos := t.Position
	t.mu.Unlock()
	return pos
}
