package main

import (
	"sync/atomic"
	"time"

	mol "github.com/LiterMC/molecular"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/math32"
)

var playerStandCube = mol.NewCube(mol.Vec3{-0.25, -1, -0.1}, mol.Vec3{0.5, 1.8, 0.2})
var playerSneakCube = mol.NewCube(mol.Vec3{-0.25, -0.4, -0.2}, mol.Vec3{0.5, 0.9, 0.4})

type Player struct {
	ctrl    *FollowControl
	outline *mol.Cube

	object *mol.Object
	queued atomic.Bool

	// status
	enabled      FollowEnabled
	status       followStatus
	lostFocus    bool
	lastX, lastY float32
}

func NewPlayer(cam *camera.Camera) (p *Player) {
	p = new(Player)
	p.ctrl = NewFollowControl(cam)
	p.ctrl.OnMove = func(dist float32, direction *math32.Vector3) bool {
		var quat math32.Quaternion
		p.ctrl.Camera().WorldQuaternion(&quat)
		dir := *direction
		dir.ApplyQuaternion(&quat)
		dir.Normalize()
		dir.MultiplyScalar(dist)
		// Get world position
		vel := p.object.Velocity()
		vel.Add(ToMolVec3(&dir))
		p.object.SetVelocity(vel)
		return true
	}
	p.ctrl.SetEnabled(FollowRot | FollowZoom | FollowMove | FollowKeys)
	p.outline = playerStandCube

	p.enabled = FollowAll
	return
}

func (p *Player) SetObject(o *mol.Object) {
	p.object = o
}

func (p *Player) Mass() float64 {
	return 1e3
}

func (p *Player) Material(f mol.Facing) *mol.Material {
	return nil
}

func (p *Player) Outline() *mol.Cube {
	return p.outline
}

func (p *Player) Tick(dt float64) {
}

func (p *Player) renderTick(r *Runner, dt time.Duration) {
	obj := p.object
	pos := obj.AbsPos()

	r.stats.Speed = obj.Velocity().Len()
	r.stats.Pos = pos

	cam := p.ctrl.Camera()
	p.ctrl.Tick(dt)
	pos.ScaleN(1. / posScale)
	cam.SetPosition(float32(pos.X), float32(pos.Y), float32(pos.Z))
}
