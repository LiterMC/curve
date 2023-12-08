package main

import (
	"math"
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

const posScale = 1e7

type PlanetBlock struct {
	object  *mol.Object
	mass    float64
	radius  float64
	outline *mol.Cube
	lastDis float64

	// render
	mat     material.IMaterial
	geo     *geometry.Geometry
	surface *graphic.Lines
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

func (b *PlanetBlock) InitNode(dist float64, radius float64) {
	var n int = 16
	// if sqrDist < b.radius {
	// 	n = 512
	// }else{
	// 	n = int(math.Sqrt(2 * math.Pi * b.radius * b.radius / sqrDist) / math.Tan(60.0 / 180 * math.Pi / 2) * b.radius / 1e5 + 0.5)
	// }
	// println("dist:", math.Sqrt(sqrDist) - b.radius, "r:", b.radius * b.radius, "n:", n)
	// if n > 512 {
	// 	n = 512
	// }
	if b.geo == nil {
		b.geo = new(geometry.Geometry)
	}
	*b.geo = *geometry.NewSphere(radius/posScale, n, n)
	if b.surface == nil {
		b.surface = new(graphic.Lines)
	}
	b.surface.Init(b.geo, material.NewBasic())
	if b.Node == nil {
		b.Node = graphic.NewMesh(b.geo, b.mat)
		b.Node.Add(helper.NewAxes(float32(radius) / posScale * 2))
		b.Node.Add(b.surface)
	}
}

func (b *PlanetBlock) SetObject(o *mol.Object) {
	b.object = o
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

	pos := b.object.AbsPos()
	diff := pos.Subbed(camPos)
	dist := diff.Len()
	far := (float64)(r.cam.Far()) * posScale / 2
	if dist > far {
		newDist := far / 2 + math.Log2(dist + b.radius)
		newR := b.radius / dist * newDist
		b.InitNode(newDist, newR)
		diff.ScaleN(newDist / dist)
		diff.Add(camPos)
		diff.ScaleN(1. / posScale)
		if b.radius < 2e6 {
			println("radius:", b.radius, newR, newDist / dist)
			println("dist:", dist, newDist)
		}
		b.Node.SetPosition(float32(diff.X), float32(diff.Y), float32(diff.Z))
		return
	}
	if b.lastDis == 0 {
		b.InitNode(dist, b.radius)
	}else if d := dist - b.lastDis; d < -50 || 100 < d {
		b.InitNode(dist, b.radius)
		b.lastDis = dist
		return
	}
	pos.ScaleN(1. / posScale)
	b.Node.SetPosition(float32(pos.X), float32(pos.Y), float32(pos.Z))
}

func InitPlanet(p *mol.Object, mass float64, radius float64, mat material.IMaterial, dist float64) (b *PlanetBlock) {
	p.SetGField(mol.NewGravityField(mass, radius))
	b = NewPlanetBlock(mass, radius, mat)
	pos := p.AbsPos()
	pos.ScaleN(1. / posScale)
	b.InitNode(dist, radius)
	b.Node.SetPosition(float32(pos.X), float32(pos.Y), float32(pos.Z))
	p.AddBlock(b)
	return
}

func (r *Runner) initSunMoon() {
	sun := r.intEng.NewObject(mol.NaturalObj, nil, mol.Vec3{0, 0, 0}, func(sun *mol.Object) {
		sun.SetVelocity(mol.Vec3{0, 1, 0})
		r.scene.Add(InitPlanet(sun, sunMass, sunRad,
			material.NewStandard(&math32.Color{1.0, 0.5, 0.2}),
			r.playerObj.AbsPos().Subbed(sun.AbsPos()).Len()).Node)
	})

	earth := r.intEng.NewObject(mol.NaturalObj, sun, mol.Vec3{-1.496e11, 0, 0}, func(earth *mol.Object) {
		earth.SetVelocity(mol.Vec3{0, 0, 2.97222e4})
		r.earth = InitPlanet(earth, earthMass*2, earthRad,
			material.NewStandard(&math32.Color{0.0, 0.0, 1.0}),
			r.playerObj.AbsPos().Subbed(earth.AbsPos()).Len()).Node
		r.scene.Add(r.earth)
	})
	r.playerObj.AttachTo(earth)
	r.playerObj.SetPos(mol.Vec3{-earthRad, earthRad + 1e7, 0})
	r.playerObj.SetVelocity(mol.Vec3{0, 0, 0})

	r.intEng.NewObject(mol.NaturalObj, earth, mol.Vec3{3e8, 1e6, 0}, func(moon *mol.Object) {
		moon.SetVelocity(mol.Vec3{0, 0, -1.022e3})
		r.scene.Add(InitPlanet(moon, moonMass, moonRad,
			material.NewStandard(&math32.Color{0.6, 0.6, 0.6}),
			r.playerObj.AbsPos().Subbed(moon.AbsPos()).Len()).Node)
	})
}
