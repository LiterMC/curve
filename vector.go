package main

import (
	mol "github.com/LiterMC/molecular"
	"github.com/g3n/engine/math32"
)

func ToMolVec3(vec3 *math32.Vector3) mol.Vec3 {
	return mol.Vec3{
		X: (float64)(vec3.X),
		Y: (float64)(vec3.Y),
		Z: (float64)(vec3.Z),
	}
}

func ToG3NVec3(vec3 *mol.Vec3) *math32.Vector3 {
	return &math32.Vector3{
		X: (float32)(vec3.X),
		Y: (float32)(vec3.Y),
		Z: (float32)(vec3.Z),
	}
}
