package main

import (
	"math"
	"time"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type FollowEnabled int

const (
	FollowNone FollowEnabled = 0x00
	FollowRot  FollowEnabled = 0x01
	FollowZoom FollowEnabled = 0x02
	FollowMove FollowEnabled = 0x04
	FollowKeys FollowEnabled = 0x8000
	FollowAll  FollowEnabled = 0xffff
)

type followStatus uint32

const (
	followStatNone followStatus = 0
	followFocusing followStatus = 1 << iota
	followMoveUp
	followMoveDown
	followMoveForward
	followMoveBackward
	followMoveLeft
	followMoveRight
	followRotUp
	followRotDown
	followRotLeft
	followRotRight
	followRollLeft
	followRollRight
	followZoomIn
	followZoomOut
	followSprint
)

type FollowControl struct {
	core.Dispatcher
	cam *camera.Camera

	// status
	enabled      FollowEnabled
	status       followStatus
	lostFocus    bool
	lastX, lastY float32

	// configs
	MinFOV, MaxFOV float32
	MoveSpeed      float32
	MouseRotSpeed  float32
	KeyRotSpeed    float32
	KeyZoomSpeed   float32

	// hooks
	OnMove   func(dist float32, direction *math32.Vector3) bool
	OnRotate func(pitch, yaw, roll float32) bool
}

func NewFollowControl(cam *camera.Camera) (fc *FollowControl) {
	fc = new(FollowControl)
	fc.Dispatcher.Initialize()
	fc.cam = cam
	fc.enabled = FollowAll

	fc.MinFOV = 10.0
	fc.MaxFOV = 100.0
	fc.MoveSpeed = 1e6
	fc.MouseRotSpeed = 0.1
	fc.KeyRotSpeed = 30 * math32.Pi / 180
	fc.KeyZoomSpeed = 5.0

	mnr := gui.Manager()
	mnr.SubscribeID(window.OnMouseDown, &fc, fc.onMouse)
	mnr.SubscribeID(window.OnScroll, &fc, fc.onScroll)
	mnr.SubscribeID(window.OnKeyUp, &fc, fc.onKey)
	mnr.SubscribeID(window.OnKeyDown, &fc, fc.onKey)
	fc.SubscribeID(window.OnCursor, &fc, fc.onCursor)
	window.Get().SubscribeID(window.OnWindowFocus, &fc, fc.onWindowFocus)

	return
}

func (fc *FollowControl) Dispose() {
	mnr := gui.Manager()
	mnr.UnsubscribeID(window.OnMouseDown, &fc)
	mnr.UnsubscribeID(window.OnScroll, &fc)
	mnr.UnsubscribeID(window.OnKeyUp, &fc)
	mnr.UnsubscribeID(window.OnKeyDown, &fc)
	fc.UnsubscribeID(window.OnCursor, &fc)
	window.Get().UnsubscribeID(window.OnWindowFocus, &fc)
}

func (fc *FollowControl) Tick(dt time.Duration) {
	if fc.lostFocus {
		fc.Pause()
		fc.lostFocus = false
	}
	dts := float32(dt.Seconds())
	if fc.enabled&FollowRot != 0 {
		rotSpeed := fc.KeyRotSpeed * dts
		if fc.status&followRotUp != 0 {
			fc.Rotate(rotSpeed, 0, 0)
		}
		if fc.status&followRotDown != 0 {
			fc.Rotate(-rotSpeed, 0, 0)
		}
		if fc.status&followRotLeft != 0 {
			fc.Rotate(0, rotSpeed, 0)
		}
		if fc.status&followRotRight != 0 {
			fc.Rotate(0, -rotSpeed, 0)
		}
		if fc.status&followRollLeft != 0 {
			fc.Rotate(0, 0, rotSpeed*2)
		}
		if fc.status&followRollRight != 0 {
			fc.Rotate(0, 0, -rotSpeed*2)
		}
	}
	if fc.enabled&FollowMove != 0 {
		moveSpeed := fc.MoveSpeed * dts
		if fc.status&followSprint != 0 {
			moveSpeed *= 1000
		}
		var dir math32.Vector3
		if fc.status&followMoveForward != 0 {
			dir.Sub(unitZ)
		}
		if fc.status&followMoveBackward != 0 {
			dir.Add(unitZ)
		}
		if fc.status&followMoveLeft != 0 {
			dir.Sub(unitX)
		}
		if fc.status&followMoveRight != 0 {
			dir.Add(unitX)
		}
		if fc.status&followMoveUp != 0 {
			dir.Add(unitY)
		}
		if fc.status&followMoveDown != 0 {
			dir.Sub(unitY)
		}
		if dir.LengthSq() > 0 {
			dir.Normalize()
			fc.Move(moveSpeed, &dir)
		}
	}
	if fc.enabled&FollowZoom != 0 {
		zoomSpeed := fc.KeyZoomSpeed * dts
		if fc.status&followZoomIn != 0 {
			fc.Zoom(zoomSpeed)
		}
		if fc.status&followZoomOut != 0 {
			fc.Zoom(-zoomSpeed)
		}
	}
}

func (fc *FollowControl) Camera() *camera.Camera {
	return fc.cam
}

func (fc *FollowControl) Enabled() FollowEnabled {
	return fc.enabled
}

func (fc *FollowControl) SetEnabled(bitmask FollowEnabled) {
	fc.enabled = bitmask
}

// Focus will start focusing the camera and disable the cursor
func (fc *FollowControl) Focus() {
	if fc.status&followFocusing == 0 {
		fc.status = followFocusing
		gui.Manager().SetCursorFocus(fc)
		win := window.Get().(*window.GlfwWindow)
		win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		win.SetCursorPos(0, 0)
		fc.lastX, fc.lastY = 0, 0
	}
}

// Pause will stop focusing the camera and release the cursor
func (fc *FollowControl) Pause() {
	if fc.status&followFocusing != 0 {
		fc.status = 0
		gui.Manager().SetCursorFocus(nil)
		win := window.Get().(*window.GlfwWindow)
		win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		if !fc.lostFocus {
			w, h := win.GetSize()
			win.SetCursorPos((float64)(w/2), (float64)(h/2))
		}
	}
}

var (
	unitX = &math32.Vector3{1, 0, 0}
	unitY = &math32.Vector3{0, 1, 0}
	unitZ = &math32.Vector3{0, 0, 1}
)

func (fc *FollowControl) Rotate(pitch, yaw, roll float32) {
	if fc.OnRotate != nil && fc.OnRotate(pitch, yaw, roll) {
		return
	}
	var q math32.Quaternion
	if pitch != 0 {
		q.SetFromAxisAngle(unitX, pitch)
		fc.cam.QuaternionMult(&q)
	}
	if yaw != 0 {
		q.SetFromAxisAngle(unitY, yaw)
		fc.cam.QuaternionMult(&q)
	}
	if roll != 0 {
		q.SetFromAxisAngle(unitZ, roll)
		fc.cam.QuaternionMult(&q)
	}
}

func (fc *FollowControl) Zoom(delta float32) {
	fov := fc.cam.Fov() - delta
	if fov < fc.MinFOV {
		fov = fc.MinFOV
	} else if fov > fc.MaxFOV {
		fov = fc.MaxFOV
	}
	fc.cam.SetFov(fov)
}

func (fc *FollowControl) Move(dist float32, direction *math32.Vector3) {
	if fc.OnMove != nil && fc.OnMove(dist, direction) {
		return
	}
	// Get world direction
	var quat math32.Quaternion
	fc.cam.WorldQuaternion(&quat)
	dir := *direction
	dir.ApplyQuaternion(&quat)
	dir.Normalize()
	dir.MultiplyScalar(dist)
	// Get world position
	var position math32.Vector3
	fc.cam.WorldPosition(&position)
	position.Add(&dir)
	fc.cam.SetPositionVec(&position)
}

func (fc *FollowControl) onMouse(evname string, ev any) {
	if fc.enabled == FollowNone {
		return
	}

	mev := ev.(*window.MouseEvent)
	switch mev.Button {
	case window.MouseButtonLeft:
		fc.Focus()
	}
}

func (fc *FollowControl) onWindowFocus(evname string, ev any) {
	if !ev.(*window.FocusEvent).Focused { // when lost focus
		if fc.status&followFocusing != 0 {
			fc.lostFocus = true
		}
	}
}

// onCursor is called when an OnCursor event is received.
func (fc *FollowControl) onCursor(evname string, ev any) {
	if fc.enabled&FollowRot == 0 || fc.status&followFocusing == 0 {
		return
	}
	cev := ev.(*window.CursorEvent)
	dx, dy := cev.Xpos-fc.lastX, cev.Ypos-fc.lastY
	fc.lastX, fc.lastY = cev.Xpos, cev.Ypos
	fc.Rotate(-dy*math.Pi/180*fc.MouseRotSpeed, -dx*math.Pi/180*fc.MouseRotSpeed, 0)
}

// onScroll is called when an OnScroll event is received.
func (fc *FollowControl) onScroll(evname string, ev any) {
	if fc.enabled&FollowZoom != 0 {
		sev := ev.(*window.ScrollEvent)
		fc.Zoom(sev.Yoffset)
	}
}

// onKey is called when an OnKeyUp/OnKeyDown event is received.
func (fc *FollowControl) onKey(evname string, ev any) {
	if fc.enabled&FollowKeys == 0 || fc.status&followFocusing == 0 {
		return
	}

	kev := ev.(*window.KeyEvent)
	switch evname {
	case window.OnKeyUp:
		switch kev.Key {
		case window.KeyLeftShift:
			fc.status &^= followSprint
		case window.KeyUp:
			fc.status &^= followRotUp | followZoomIn
		case window.KeyDown:
			fc.status &^= followRotDown | followZoomOut
		case window.KeyLeft:
			fc.status &^= followRotLeft
		case window.KeyRight:
			fc.status &^= followRotRight
		case window.KeyQ:
			fc.status &^= followRollLeft
		case window.KeyE:
			fc.status &^= followRollRight
		case window.KeyW:
			fc.status &^= followMoveForward
		case window.KeyS:
			fc.status &^= followMoveBackward
		case window.KeyA:
			fc.status &^= followMoveLeft
		case window.KeyD:
			fc.status &^= followMoveRight
		case window.KeySpace:
			fc.status &^= followMoveUp
		case window.KeyC:
			fc.status &^= followMoveDown
		}
	case window.OnKeyDown:
		switch kev.Key {
		case window.KeyEscape:
			fc.Pause()
		case window.KeyLeftShift:
			fc.status |= followSprint
		case window.KeyUp:
			if kev.Mods&window.ModAlt != 0 {
				fc.status |= followZoomIn
			} else {
				fc.status |= followRotUp
			}
		case window.KeyDown:
			if kev.Mods&window.ModAlt != 0 {
				fc.status |= followZoomOut
			} else {
				fc.status |= followRotDown
			}
		case window.KeyLeft:
			fc.status |= followRotLeft
		case window.KeyRight:
			fc.status |= followRotRight
		case window.KeyQ:
			fc.status |= followRollLeft
		case window.KeyE:
			fc.status |= followRollRight
		case window.KeyW:
			fc.status |= followMoveForward
		case window.KeyS:
			fc.status |= followMoveBackward
		case window.KeyA:
			fc.status |= followMoveLeft
		case window.KeyD:
			fc.status |= followMoveRight
		case window.KeySpace:
			fc.status |= followMoveUp
		case window.KeyC:
			fc.status |= followMoveDown
		}
	}
}
