package main

import (
	"math"
	"sync/atomic"
	"time"

	mol "github.com/LiterMC/molecular"
	// "github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/helper"
)

var (
	sunRad    = 6.9634e8
	earthRad  = 6.371e6
	moonRad   = 1.7374e6
	sunMass   = 3.955e30
	earthMass = 5.972e24
	moonMass  = 7.34767309e22
)

const posScale = 1 << 23

type PlanetBlock struct {
	object  atomic.Pointer[mol.Object]
	mass    float64
	radius  float64
	outline *mol.Cube
	lastDis float64
	lastN   int

	// render
	mat     material.IMaterial
	geo     *geometry.Geometry
	Node    *graphic.Mesh
}

var _ mol.Block = (*PlanetBlock)(nil)

func NewPlanetBlock(mass float64, radius float64, mat material.IMaterial) *PlanetBlock {
	return &PlanetBlock{
		mass:    mass,
		radius:  radius,
		outline: mol.NewCubeFromCenter(mol.Vec3{radius * 2, radius * 2, radius * 2}),
		mat: mat,
	}
}

func (b *PlanetBlock) InitNode(dist float64) {
	var n int = 16
	if dist < b.radius {
		n = 512
	} else {
		const (
			camNear = 0.01
			camFOV  = 60.0
		)
		camConst := camNear / math.Tan(camFOV/2/180*math.Pi)
		n = int(math.Sqrt(2*math.Pi*b.radius)*b.radius/dist*camConst + 8.5)
		if n > 512 {
			n = 512
		}
	}

	if b.lastN == n {
		return
	}
	b.lastN = n

	if b.geo == nil {
		b.geo = new(geometry.Geometry)
	}
	*b.geo = *geometry.NewSphere(b.radius/posScale, n, n)
	if b.Node == nil {
		b.Node = graphic.NewMesh(b.geo, b.mat)
		b.Node.Add(helper.NewAxes(float32(b.radius) / posScale * 2))
	}
}

func (b *PlanetBlock) SetObject(o *mol.Object) {
	b.object.Store(o)
}

func (b *PlanetBlock) Mass() float64 {
	return b.mass
}

func (b *PlanetBlock) Material(f mol.Facing) *mol.Material {
	return nil
}

func (b *PlanetBlock) Outline() *mol.Cube {
	return b.outline
}

func (b *PlanetBlock) Tick(dt float64) {
}

func (b *PlanetBlock) renderTick(r *Runner, dt time.Duration) {
	cPos := r.cam.Position()
	camPos := ToMolVec3(&cPos)
	camPos.ScaleN(posScale)

	obj := b.object.Load()
	pos := obj.AbsPosLocked()
	diff := pos.Subbed(camPos)
	dist := diff.Len()
	if b.lastDis == 0 {
		b.InitNode(dist)
	} else if d := dist - b.lastDis; d < -50 || 100 < d {
		b.InitNode(dist)
		b.lastDis = dist
		return
	}
	pos.ScaleN(1. / posScale)
	b.Node.SetPosition(float32(pos.X), float32(pos.Y), float32(pos.Z))
}

func InitPlanet(p *mol.Object, mass float64, radius float64, mat material.IMaterial, dist float64) (b *PlanetBlock) {
	p.SetRadius(radius)
	p.FillGfields()
	b = NewPlanetBlock(mass, radius, mat)
	pos := p.AbsPos()
	pos.ScaleN(1. / posScale)
	b.InitNode(dist)
	b.Node.SetPosition(float32(pos.X), float32(pos.Y), float32(pos.Z))
	p.AddBlock(b)
	return
}

func (r *Runner) initSunMoon() {
	sun := r.intEng.NewObject(mol.NaturalObj, nil, mol.Vec3{0, 0, 0}, func(sun *mol.Object) {
		println("sun:", sun.String())
		sun.SetVelocity(mol.Vec3{0, 1, 0})
		sunColor := &math32.Color{1.0, 0.5, 0.2}
		sunMat := material.NewStandard(sunColor)
		// sunMat := material.NewPhysical().
		// 	SetBaseColorFactor(&math32.Color4{sunColor.R, sunColor.G, sunColor.B, 1}).
		// 	SetMetallicFactor(0).
		// 	SetRoughnessFactor(1).
		// 	SetEmissiveFactor(sunColor)
		sunNode := InitPlanet(sun, sunMass, sunRad, sunMat,
			r.playerObj.AbsPos().Subbed(sun.AbsPos()).Len()).Node
		r.mainScene.Add(sunNode)
	})

	earth := r.intEng.NewObject(mol.NaturalObj, sun, mol.Vec3{-1.496e11, 0, 0}, func(earth *mol.Object) {
		println("earth:", earth.String())
		earth.SetVelocity(mol.Vec3{0, 0, 2.97222e4})
		earthColor := &math32.Color{0.0, 0.0, 1.0}
		earthMat := material.NewStandard(earthColor)
		r.earth = InitPlanet(earth, earthMass, earthRad, earthMat,
			r.playerObj.AbsPos().Subbed(earth.AbsPos()).Len()).Node
		r.mainScene.Add(r.earth)
	})
	r.playerObj.AttachTo(earth)
	r.playerObj.SetPos(mol.Vec3{-earthRad, earthRad + 1e7, 0})
	r.playerObj.SetVelocity(mol.Vec3{0, 0, 0})

	r.intEng.NewObject(mol.NaturalObj, earth, mol.Vec3{-3e8, 1e7, 0}, func(moon *mol.Object) {
		println("moon:", moon.String())
		moon.SetVelocity(mol.Vec3{0, 0, -1.022e3})
		r.mainScene.Add(InitPlanet(moon, moonMass, moonRad,
			material.NewStandard(&math32.Color{0.6, 0.6, 0.6}),
			r.playerObj.AbsPos().Subbed(moon.AbsPos()).Len()).Node)
	})
}
