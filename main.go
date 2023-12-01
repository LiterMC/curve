// Curve, a 3D space game powered by relativity
// Copyright (C) 2023 Kevin Z <zyxkad@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
	// mol "github.com/LiterMC/molecular"
)

func main() {
	app := app.App(800, 600, "Curve")
	win := app.IWindow.(*window.GlfwWindow)
	win.SetTitle("Curve | Initing")
	app.Run(func(rend *renderer.Renderer, dt time.Duration) {
		//
	})
}
