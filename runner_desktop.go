//go:build !wasm

package main

import (
	"github.com/g3n/engine/window"
)

func (r *Runner) SetTitle(title string) {
	r.Application.IWindow.(*window.GlfwWindow).SetTitle(title)
}
