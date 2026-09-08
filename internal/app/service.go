package app

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// Причины остановки службы.
var (
	// ErrUserActivity — курсор оказался не там, куда мы его вели: пользователь
	// взял мышь, ОС зажала курсор на краю, либо движение вообще не применяется.
	ErrUserActivity = errors.New("курсор сместился не по нашей команде")
	// ErrStopped — остановка по сигналу через StopChan.
	ErrStopped = errors.New("получен сигнал остановки")
)

const (
	sleepInterval = 8 * time.Millisecond
	epsilon       = 20.0

	// summaryEvery — как часто писать в журнал сводку о движении.
	summaryEvery = 5 * time.Second
	// tickHistoryLen — сколько последних тактов держать для посмертного разбора.
	tickHistoryLen = 32
)

type Service struct {
	Engine          *Engine
	MouseController *MouseController
	MouseTracker    *MouseTracker
	StopChan        chan struct{}
	Log             *slog.Logger
}

// tickRecord — снимок одного такта для разбора после остановки.
type tickRecord struct {
	n      int
	at     time.Duration
	target Vector
	actual Vector
	drift  float64
}

// stats — накопленная статистика движения за запуск.
type stats struct {
	ticks    int
	bounces  int
	distance float64
	maxDrift float64

	// warnedStuck — предупреждение о вырожденном движении уже записано.
	warnedStuck bool
}

// Run ведёт курсор до тех пор, пока не заметит чужое вмешательство или
// сигнал остановки. Возвращает причину остановки — она же попадает в журнал.
func (c *Service) Run() error {
	if c.StopChan == nil {
		c.StopChan = make(chan struct{})
	}

	log := c.Log
	if log == nil {
		log = slog.Default()
	}

	start := time.Now()
	log.Info("цикл движения запущен",
		"позиция", vec(c.Engine.Position),
		"скорость", vec(c.Engine.Velocity),
		"экран", fmt.Sprintf("%dx%d", c.Engine.Width, c.Engine.Height),
		"такт", sleepInterval.String(),
		"порог_расхождения", epsilon,
	)

	var (
		st          stats
		history     = make([]tickRecord, 0, tickHistoryLen)
		lastSummary = start
		prevPos     = c.Engine.Position
	)

	for {
		select {
		case <-c.StopChan:
			c.summary(log, "остановка по сигналу", &st, time.Since(start))
			return ErrStopped
		default:
		}

		st.ticks++
		currentPos := c.MouseController.GetPosition()

		drift := math.Max(
			math.Abs(currentPos.X-c.Engine.Position.X),
			math.Abs(currentPos.Y-c.Engine.Position.Y),
		)
		if drift > st.maxDrift {
			st.maxDrift = drift
		}

		if len(history) == tickHistoryLen {
			history = history[1:]
		}
		history = append(history, tickRecord{
			n:      st.ticks,
			at:     time.Since(start),
			target: c.Engine.Position,
			actual: currentPos,
			drift:  drift,
		})

		if !closeEnough(currentPos, c.Engine.Position, epsilon) {
			log.Warn("расхождение превысило порог — останавливаюсь",
				"такт", st.ticks,
				"прошло", time.Since(start).Round(time.Millisecond).String(),
				"ожидалось", vec(c.Engine.Position),
				"фактически", vec(currentPos),
				"расхождение", round1(drift),
				"порог", epsilon,
			)
			dumpHistory(log, history)
			c.summary(log, "итог", &st, time.Since(start))
			return ErrUserActivity
		}

		velBefore := c.Engine.Velocity
		c.Engine.Update()
		if c.Engine.Velocity != velBefore {
			st.bounces++
			log.Debug("отражение от края",
				"такт", st.ticks,
				"позиция", vec(c.Engine.Position),
				"скорость", vec(c.Engine.Velocity),
			)
		}

		pos := c.Engine.Position
		st.distance += math.Hypot(pos.X-prevPos.X, pos.Y-prevPos.Y)
		prevPos = pos

		c.MouseController.Move(&pos)

		if time.Since(lastSummary) >= summaryEvery {
			c.summary(log, "движение", &st, time.Since(start))
			c.warnIfStuck(log, &st)
			lastSummary = time.Now()
		}

		time.Sleep(sleepInterval)
	}
}

// summary пишет в журнал сводку о проделанном движении.
func (c *Service) summary(log *slog.Logger, msg string, st *stats, elapsed time.Duration) {
	rate := 0.0
	if elapsed > 0 {
		rate = float64(st.ticks) / elapsed.Seconds()
	}

	log.Info(msg,
		"такты", st.ticks,
		"тактов_в_сек", round1(rate),
		"отражений", st.bounces,
		"пройдено_px", math.Round(st.distance),
		"позиция", vec(c.Engine.Position),
		"макс_расхождение", round1(st.maxDrift),
		"прошло", elapsed.Round(time.Millisecond).String(),
	)
}

// warnIfStuck предупреждает о вырожденном движении.
//
// Область движения берётся из размеров ГЛАВНОГО экрана, а курсор может
// находиться на другом мониторе — в том числе с отрицательными координатами.
// Тогда условие выхода за границу срабатывает на каждом такте, движок
// колеблется в полосе шириной в один шаг, и программа работает вхолостую.
func (c *Service) warnIfStuck(log *slog.Logger, st *stats) {
	if st.warnedStuck || st.ticks < 100 {
		return
	}
	if float64(st.bounces) < 0.5*float64(st.ticks) {
		return
	}

	st.warnedStuck = true
	log.Warn("движение выродилось: отражение почти на каждом такте",
		"такты", st.ticks,
		"отражений", st.bounces,
		"позиция", vec(c.Engine.Position),
		"область_движения", fmt.Sprintf("X:[0..%d] Y:[0..%d]", c.Engine.Width-1, c.Engine.Height-1),
		"вероятная_причина", "курсор вне главного экрана; границы считаются только по нему",
	)
}

// dumpHistory выкладывает последние такты перед остановкой — по ним видно,
// был ли это резкий скачок (пользователь, зажим ОС) или постепенный дрейф
// (команды на перемещение не применяются).
func dumpHistory(log *slog.Logger, history []tickRecord) {
	log.Warn("последние такты перед остановкой", "количество", len(history))
	for _, r := range history {
		log.Warn("такт",
			"n", r.n,
			"t", r.at.Round(time.Millisecond).String(),
			"цель", vec(r.target),
			"факт", vec(r.actual),
			"расхождение", round1(r.drift),
		)
	}
}

func vec(v Vector) string {
	return fmt.Sprintf("(%.1f, %.1f)", v.X, v.Y)
}

func round1(f float64) float64 {
	return math.Round(f*10) / 10
}

func closeEnough(a, b Vector, epsilon float64) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx <= epsilon && dy <= epsilon
}
