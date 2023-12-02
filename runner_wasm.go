//go:build wasm

package main

import (
	"syscall/js"

	"github.com/g3n/engine/window"
)

const _ = panic("Wasm have problem right now")

var document = js.Global().Get("document")

func (r *Runner) SetTitle(title string) {
	document.Set("title", title)
}
