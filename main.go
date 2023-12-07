package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "puca")

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(10.0, 10.0, 10.0)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Up = rl.NewVector3(0, 1, 0)
	camera.Projection = rl.CameraPerspective

	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.LightGray)
		rl.BeginMode3D(camera)
		rl.EndMode3D()
		rl.EndDrawing()
	}
}
