package gui

import (
	_ "embed"

	"github.com/NikosGour/BigDownloadP2P/assets"
	"github.com/NikosGour/BigDownloadP2P/build"
	log "github.com/NikosGour/logging/src"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const default_font_filename = "Lexend-Regular.ttf"

type GUI struct {
	desired_monitor int
	current_monitor int
	monitor_height  int
	monitor_width   int
	screen_height   int
	screen_width    int

	default_font rl.Font

	// // TODO : Change name of NavBar to AlgorithmHud
	// navbar    *NavBar
	// grid      *Grid
	// color_hud *ColorHud

	initilized bool
	debug_mode bool
}

func NewGUI() *GUI {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagVsyncHint | rl.FlagWindowHighdpi | rl.FlagMsaa4xHint)

	log.Debug("Monitor Count: %#v", rl.GetMonitorCount())

	monitor := 0
	monitor_width := int32(1920)  //int32(rl.GetMonitorWidth(monitor))
	monitor_height := int32(1080) //int32(rl.GetMonitorHeight(monitor))

	rl.InitWindow(
		monitor_width,
		monitor_height,
		"Path Finding Visualization by Nikos Gournakis")

	rl.SetTargetFPS(60)

	rl.SetWindowPosition(int(rl.GetMonitorPosition(monitor).X), int(rl.GetMonitorPosition(monitor).Y))
	rl.SetWindowMonitor(monitor)

	g := &GUI{desired_monitor: monitor, initilized: false, debug_mode: build.DEBUG_MODE}
	g.configureMonitorScreenSizes()

	default_font, err := assets.Fonts.ReadFile(default_font_filename)
	if err != nil {
		log.Fatal("Couldn't load font: %s, error: %s", default_font_filename, err)
	}
	g.default_font = rl.LoadFontFromMemory(".ttf", default_font, 512, nil)
	rl.SetTextureFilter(g.default_font.Texture, rl.FilterPoint)
	log.Debug("%#v", g)
	return g
}
func (g *GUI) init() {
	if !rl.IsWindowFullscreen() {
		rl.SetWindowSize(g.monitor_width, g.monitor_height)
		rl.ToggleFullscreen()
		g.configureMonitorScreenSizes()
		// rl.ClearWindowState(rl.FlagWindowResizable)
	}

	g.initilized = true
}
func (g *GUI) runMainLoop() {
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		// Wait for window initilization
		// log.Debug("monitor: %v, monitor_height: %v, monitor_width: %v", this.current_monitor, this.monitor_height, this.monitor_width)
		// log.Debug("screen_height: %v, screen_width: %v", this.screen_height, this.screen_width)
		g.configureMonitorScreenSizes()
		if !g.initilized && g.current_monitor == g.desired_monitor {
			g.init()
		}
		// end of initilization

		rl.BeginDrawing()
		// ---------------- DRAWING ----------------------------
		rl.ClearBackground(rl.NewColor(0x18, 0x18, 0x18, 0xFF))

		if g.initilized {
			rl.DrawTextEx(g.default_font, "Hello World", rl.Vector2{X: 0, Y: 0}, 32, 0, rl.Blue)
		}
		// ---------------- END DRAWING ------------------------
		rl.EndDrawing()
	}
}
func (g *GUI) configureMonitorScreenSizes() {
	g.current_monitor = rl.GetCurrentMonitor()
	g.screen_height = rl.GetScreenHeight()
	g.screen_width = rl.GetScreenWidth()

	g.monitor_height = rl.GetMonitorHeight(g.current_monitor)
	g.monitor_width = rl.GetMonitorWidth(g.current_monitor)
}
