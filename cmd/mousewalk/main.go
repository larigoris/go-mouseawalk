package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"GOwalk/internal/app"
	"GOwalk/internal/logging"

	"github.com/go-vgo/robotgo"
)

func main() {
	// Каталог журналов лежит в ~/Library/Logs, который Finder скрывает,
	// поэтому путь к нему нужно уметь получить не заглядывая в исходники.
	showLogDir := flag.Bool("logs", false, "показать каталог журналов и выйти")
	flag.Parse()

	if *showLogDir {
		dir, err := logging.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "не удалось определить каталог журналов:", err)
			os.Exit(1)
		}
		fmt.Println(dir)
		return
	}

	// GOWALK_VERBOSE=1 включает отладочные записи (отражения от краёв и пр.).
	session, err := logging.Start(os.Getenv("GOWALK_VERBOSE") != "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "не удалось открыть журнал:", err)
		os.Exit(1)
	}
	log := session.Logger

	width, height := robotgo.GetScreenSize()
	log.Info("главный экран", "ширина", width, "высота", height)
	logDisplays(log)

	mouse := &app.MouseController{}
	startPos := mouse.GetPosition()
	log.Info("исходная позиция курсора", "позиция", fmt.Sprintf("(%.0f, %.0f)", startPos.X, startPos.Y))

	engine := &app.Engine{
		Position: startPos,
		Velocity: app.Vector{
			X: 5.5,
			Y: 5.5,
		},
		Width:  width,
		Height: height,
	}

	tracker := &app.MouseTracker{}
	tracker.Start()
	log.Info("глобальный хук ввода запущен")

	controller := &app.Service{
		Engine:          engine,
		MouseController: mouse,
		MouseTracker:    tracker,
		Log:             log,
	}

	session.Finish(controller.Run())
}

// logDisplays записывает раскладку мониторов: по ней видно, попадает ли
// область движения в реально существующее пространство курсора.
func logDisplays(log *slog.Logger) {
	n := robotgo.DisplaysNum()
	log.Info("дисплеи", "количество", n)

	for i := 0; i < n; i++ {
		r := robotgo.GetScreenRect(i)
		log.Info("дисплей",
			"индекс", i,
			"origin", fmt.Sprintf("(%d, %d)", r.X, r.Y),
			"размер", fmt.Sprintf("%dx%d", r.W, r.H),
		)
	}
}
