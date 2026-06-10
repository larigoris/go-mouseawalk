package main

import hook "github.com/robotn/gohook"

type MouseTracker struct {
	Position Vector
}

func (t *MouseTracker) Start() {
	evChan := hook.Start()
	go func() {
		for ev := range evChan {
			if ev.Kind == hook.MouseMove {
				t.Position.X = int(ev.X)
				t.Position.Y = int(ev.Y)
			}
		}
	}()
}
