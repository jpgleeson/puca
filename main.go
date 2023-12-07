package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "puca")

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(100.0, 10.0, 10.0)
	camera.Target = rl.NewVector3(0.0, 10.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	modelTriangles := loadObjText()

	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.LightGray)
		rl.BeginMode3D(camera)
		rl.DrawGrid(20, 10.0)
		for _, face := range modelTriangles {
			rl.DrawPoint3D(face.Point1, rl.Blue)
			rl.DrawPoint3D(face.Point2, rl.Blue)
			rl.DrawPoint3D(face.Point3, rl.Blue)
		}
		rl.EndMode3D()
		rl.EndDrawing()
	}
}

type modelFace struct {
	Point1 rl.Vector3
	Point2 rl.Vector3
	Point3 rl.Vector3
}

func loadObjText() []modelFace {
	vertices := make(map[int]rl.Vector3)
	verticeIndex := 1

	faces := make([]modelFace, 0)

	file, err := os.Open("Spearman.obj")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineComponents := strings.Split(fileScanner.Text(), " ")
		// fmt.Printf("line: %s\n", line)
		lineType := lineComponents[0]
		switch lineType {
		// Lines that start with v are vertex declarations
		case "v":
			vertexX := stringToFloat32(lineComponents[1])
			vertexY := stringToFloat32(lineComponents[2])
			vertexZ := stringToFloat32(lineComponents[3])
			vertices[verticeIndex] = rl.NewVector3(vertexX, vertexY, vertexZ)
			verticeIndex = verticeIndex + 1
		case "vn":
			// fmt.Println("This line is a normal")
		case "f":
			point1Components := strings.Split(lineComponents[1], "/")
			point2Components := strings.Split(lineComponents[2], "/")
			point3Components := strings.Split(lineComponents[3], "/")

			point1Vertex, err := strconv.Atoi(point1Components[0])
			if err != nil {
				fmt.Println("Error converting to int")
			}
			point2Vertex, err := strconv.Atoi(point2Components[0])
			if err != nil {
				fmt.Println("Error converting to int")
			}
			point3Vertex, err := strconv.Atoi(point3Components[0])
			if err != nil {
				fmt.Println("Error converting to int")
			}

			faces = append(faces, modelFace{
				Point1: vertices[point1Vertex],
				Point2: vertices[point2Vertex],
				Point3: vertices[point3Vertex],
			})
		default:
			fmt.Sprintln("Unknown line type %s", line)
		}
	}

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	return faces
}

func stringToFloat32(input string) float32 {
	parsedValue, err := strconv.ParseFloat(input, 32)
	if err != nil {
		fmt.Printf("Error parsing X")
	}
	floatValue := float32(parsedValue)
	return floatValue
}
