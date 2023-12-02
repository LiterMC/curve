package main

import (
	"fmt"
	"time"

	mol "github.com/LiterMC/molecular"
	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type Runner struct {
	*app.Application
	scene   *core.Node
	cam     *camera.Camera
	camctrl *camera.OrbitControl

	// status
	lastFpsUpdate time.Time

	internalSvr *mol.Engine
}

func (r *Runner) Init() (err error) {
	now := time.Now()
	r.SetTitle("Curve")

	scene := core.NewNode()
	r.scene = scene
	gui.Manager().Set(scene)

	r.cam = camera.New(1)
	r.cam.SetPosition(0, 0, 3)
	scene.Add(r.cam)
	r.camctrl = camera.NewOrbitControl(r.cam)
	{
		onResize := func(name string, value any) {
			scaleX, scaleY := r.GetScale()
			w, h := r.GetSize()
			r.Gls().Viewport(0, 0, (int32)((float64)(w)*scaleX), (int32)((float64)(h)*scaleY))
			r.cam.SetAspect((float32)(w) / (float32)(h))
		}
		r.Subscribe(window.OnWindowSize, onResize)
		onResize("", nil)
	}

	r.lastFpsUpdate = now

	// Create a blue torus and add it to the scene
	geom := geometry.NewTorus(1, .4, 12, 32, math32.Pi*2)
	mat := material.NewStandard(math32.NewColor("DarkBlue"))
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)

	// Create and add a button to the scene
	btn := gui.NewButton("Make Red")
	btn.SetPosition(40, 40)
	btn.SetSize(40, 40)
	btn.Subscribe(gui.OnClick, func(name string, value any) {
		mat.SetColor(math32.NewColor("DarkRed"))
	})
	scene.Add(btn)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	scene.Add(helper.NewAxes(1))

	r.Gls().ClearColor(0.05, 0.05, 0.05, 1.0)

	return
}

func (r *Runner) Tick(rend *renderer.Renderer, dt time.Duration) {
	now := time.Now()
	if now.Sub(r.lastFpsUpdate) >= time.Second {
		r.lastFpsUpdate = now
		r.SetTitle(fmt.Sprintf("Curve | FPS: %.1f", (float32)(time.Second)/(float32)(dt)))
	}

	r.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	rend.Render(r.scene, r.cam)
}
