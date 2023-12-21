package main

import (
	"fmt"
	"log"
	"time"

	mol "github.com/LiterMC/molecular"
	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type Tickable interface {
	Tick(dt time.Duration)
}

type Runner struct {
	*app.Application
	mainScene *core.Node
	cam       *camera.Camera

	// status
	lastFpsUpdate time.Time
	frameCount int
	stats         guiStatus

	intEng    *mol.Engine // internal physics engine
	player    *Player
	playerObj *mol.Object
	earth     core.INode
}

type guiStatus struct {
	Speed    float64
	guiSpeed *gui.Label
	Pos      mol.Vec3
	guiPos   *gui.Label
	Anchor   *mol.Object
	guiAnchor *gui.Label
	guiAnchorPos *gui.Label
}

func (s *guiStatus) update() {
	s.guiSpeed.SetText(fmt.Sprintf("%.2f m/s", s.Speed))
	s.guiPos.SetText(fmt.Sprintf("%.1f, %.1f, %.1f", s.Pos.X, s.Pos.Y, s.Pos.Z))
	s.guiAnchor.SetText(s.Anchor.Id().String())
	s.guiAnchorPos.SetText((fmt.Sprintf("%.1f, %.1f, %.1f", s.Anchor.Pos().X, s.Anchor.Pos().Y, s.Anchor.Pos().Z)))
}

func (r *Runner) initEngine(now time.Time) {
	r.intEng = mol.NewEngine(mol.Config{})

	go func(last time.Time) {
		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()
		logc := 0
		for {
			select {
			case t := <-ticker.C:
				dt := t.Sub(last)
				start := time.Now()
				r.intEng.Tick(dt)
				spt := time.Since(start)
				if logc++; logc > 100 {
					logc = 0
					log.Println("Time per tick:", spt, "; events:", r.intEng.Events())
				}
				last = t
			}
		}
	}(now)

}

func (r *Runner) Init() (err error) {
	now := time.Now()

	r.SetTitle("Curve")
	r.initEngine(now)

	log.Println("new scene")
	scene := core.NewNode()
	r.mainScene = scene

	r.cam = camera.NewPerspective(1, 0.01, 2*60*60*mol.C/posScale, 60, camera.Vertical)
	r.player = NewPlayer(r.cam)
	r.playerObj = r.intEng.NewObject(mol.LivingObj, nil, mol.Vec3{0, 0, 0}, func(player *mol.Object) {
		player.AddBlock(r.player)
		player.SetVelocity(mol.Vec3{0, 0, 0})
	})
	log.Println("generating sun moon")
	r.initSunMoon()
	scene.Add(r.cam)
	{
		onResize := func(name string, value any) {
			scaleX, scaleY := r.GetScale()
			w, h := r.GetSize()
			r.Gls().Viewport(0, 0, (int32)((float64)(w)*scaleX), (int32)((float64)(h)*scaleY))
			r.cam.SetAspect((float32)(w) / (float32)(h))
		}
		onResize("", nil)
		r.Subscribe(window.OnWindowSize, onResize)
	}

	r.lastFpsUpdate = now
	r.initGUIs()

	// Create and add a button to the scene
	// btn := gui.NewButton("Make Red")
	// btn.SetPosition(40, 40)
	// btn.SetSize(40, 40)
	// btn.Subscribe(gui.OnClick, func(name string, value any) {
	// 	mat.SetColor(math32.NewColor("DarkRed"))
	// })
	// scene.Add(btn)

	// Create and add lights to the scene
	log.Println("add light")
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))

	scene.Add(helper.NewAxes(0))

	log.Println("set clear color")
	r.Gls().ClearColor(0.05, 0.05, 0.05, 1)
	log.Println("done")

	gui.Manager().Set(r.mainScene)
	return
}

func (r *Runner) initGUIs() {
	w, h := r.GetSize()
	indicator, err := gui.NewImage("./assets/indicator.png")
	if err != nil {
		log.Panic(err)
	}
	indicator.SetContentSize(9, 9)
	indicator.SetPosition((float32)((w-9)/2), (float32)((h-9)/2))
	r.mainScene.Add(indicator)

	statBox := gui.NewPanel(400, 22 * 10)
	statBox.SetPosition(10, 10)
	statBox.SetPaddings(5, 5, 5, 5)
	statBox.SetColor4(&math32.Color4{0.7, 0.7, 0.7, 0.5})

	speedLb := gui.NewLabel("Speed:")
	statBox.Add(speedLb)
	r.stats.guiSpeed = gui.NewLabel("")
	r.stats.guiSpeed.SetPosition(speedLb.Width()+5, speedLb.Position().Y)
	statBox.Add(r.stats.guiSpeed)

	posLb := gui.NewLabel("Pos:")
	posLb.SetPositionY(22)
	statBox.Add(posLb)
	r.stats.guiPos = gui.NewLabel("")
	r.stats.guiPos.SetPosition(posLb.Width()+5, posLb.Position().Y)
	statBox.Add(r.stats.guiPos)

	anchorLb := gui.NewLabel("Anchor:")
	anchorLb.SetPositionY(44)
	statBox.Add(anchorLb)
	r.stats.guiAnchor = gui.NewLabel("")
	r.stats.guiAnchor.SetPosition(anchorLb.Width()+5, anchorLb.Position().Y)
	statBox.Add(r.stats.guiAnchor)

	anchorPosLb := gui.NewLabel("Anchor Pos:")
	anchorPosLb.SetPositionY(66)
	statBox.Add(anchorPosLb)
	r.stats.guiAnchorPos = gui.NewLabel("")
	r.stats.guiAnchorPos.SetPosition(anchorPosLb.Width()+5, anchorPosLb.Position().Y)
	statBox.Add(r.stats.guiAnchorPos)

	r.mainScene.Add(statBox)
}

func (r *Runner) Tick(rend *renderer.Renderer, dt time.Duration) {
	now := time.Now()
	if r.frameCount++; now.Sub(r.lastFpsUpdate) >= time.Second {
		r.SetTitle(fmt.Sprintf("Curve | FPS: %d", r.frameCount))
		r.lastFpsUpdate = now
		r.frameCount = 0
	}

	r.intEng.ForeachBlock(func(b mol.Block) {
		if t, ok := b.(interface {
			renderTick(r *Runner, dt time.Duration)
		}); ok {
			t.renderTick(r, dt)
		}
	})

	r.stats.update()

	r.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	rend.Render(r.mainScene, r.cam)
}
