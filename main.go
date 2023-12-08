package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "puca")
	rl.SetTargetFPS(60)

	var modelPath string
	if len(os.Args) > 1 {
		modelPath = os.Args[1]
	}
	modelLoaded := true

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(100.0, 10.0, 10.0)
	camera.Target = rl.NewVector3(0.0, 10.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	distance := float32(100)
	theta := float64(0.2)
	phi := float64(0.2)

	modelOffset := rl.NewVector3(0, 0, 0)

	modelTriangles, err := loadObjText(modelPath)
	if err != nil {
		modelLoaded = false
		modelPath = "Error loading model."
	}

	faceColours := makeNormalColourCache(modelTriangles)

	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {

		if rl.IsKeyDown(rl.KeyRight) {
			theta += 0.01
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			theta -= 0.01
		}
		if rl.IsKeyDown(rl.KeyUp) {
			phi += 0.01
		}
		if rl.IsKeyDown(rl.KeyDown) {
			phi -= 0.01
		}
		if rl.IsKeyPressed(rl.KeyPageUp) {
			distance -= 10
		}
		if rl.IsKeyPressed(rl.KeyPageDown) {
			distance += 10
		}

		if rl.IsKeyPressed(rl.KeyW) {
			modelOffset.X += 10
		}
		if rl.IsKeyPressed(rl.KeyS) {
			modelOffset.X -= 10
		}
		if rl.IsKeyPressed(rl.KeyA) {
			modelOffset.Z += 10
		}
		if rl.IsKeyPressed(rl.KeyD) {
			modelOffset.Z -= 10
		}

		camera.Position.X = distance * float32(math.Sin(theta)*math.Cos(phi))
		camera.Position.Y = distance * float32(math.Sin(phi))
		camera.Position.Z = distance * float32(math.Cos(theta)*math.Cos(phi))

		rl.BeginDrawing()
		rl.ClearBackground(rl.LightGray)
		rl.BeginMode3D(camera)
		rl.DrawGrid(20, 10.0)
		if modelLoaded {
			for _, face := range modelTriangles {
				rl.DrawTriangle3D(OffsetVector3(face.Point1, modelOffset), OffsetVector3(face.Point2, modelOffset), OffsetVector3(face.Point3, modelOffset), faceColours[face])
			}
		}
		rl.EndMode3D()
		rl.DrawText(modelPath, 10, 10, 12, rl.Black)
		rl.EndDrawing()
	}
}

type modelFace struct {
	Point1 rl.Vector3
	Point2 rl.Vector3
	Point3 rl.Vector3
	Normal rl.Vector3
}

func OffsetVector3(vector rl.Vector3, offset rl.Vector3) rl.Vector3 {
	vector.X = vector.X + offset.X
	vector.Y = vector.Y + offset.Y
	vector.Z = vector.Z + offset.Z
	return vector
}

func loadObjText(modelPath string) ([]modelFace, error) {
	vertices := make(map[int]rl.Vector3)
	verticeIndex := 1

	normals := make(map[int]rl.Vector3)
	normalIndex := 1

	faces := make([]modelFace, 0)

	file, err := os.Open(modelPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineComponents := strings.Split(fileScanner.Text(), " ")
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
			normalX := stringToFloat32(lineComponents[1])
			normalY := stringToFloat32(lineComponents[2])
			normalZ := stringToFloat32(lineComponents[3])
			normals[normalIndex] = rl.NewVector3(normalX, normalY, normalZ)
			normalIndex = normalIndex + 1
		case "f":
			point1Components := strings.Split(lineComponents[1], "//")
			point2Components := strings.Split(lineComponents[2], "//")
			point3Components := strings.Split(lineComponents[3], "//")

			point1Vertex, point1Normal, err := getVertexAndNormalIndices(point1Components)
			if err != nil {
				fmt.Println("Error getting indices")
				return nil, err
			}
			point2Vertex, point2Normal, err := getVertexAndNormalIndices(point2Components)
			if err != nil {
				fmt.Println("Error getting indices")
				return nil, err
			}
			point3Vertex, point3Normal, err := getVertexAndNormalIndices(point3Components)
			if err != nil {
				fmt.Println("Error getting indices")
				return nil, err
			}

			faceNormal := getFaceNormal(normals[point1Normal], normals[point2Normal], normals[point3Normal])

			faces = append(faces, modelFace{
				Point1: vertices[point1Vertex],
				Point2: vertices[point2Vertex],
				Point3: vertices[point3Vertex],
				Normal: faceNormal,
			})
		default:
			fmt.Sprintln("Unknown line type %s", line)
		}
	}

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	return faces, nil
}

func stringToFloat32(input string) float32 {
	parsedValue, err := strconv.ParseFloat(input, 32)
	if err != nil {
		fmt.Printf("Error parsing X")
	}
	floatValue := float32(parsedValue)
	return floatValue
}

func getVertexAndNormalIndices(components []string) (int, int, error) {
	vertex, err := strconv.Atoi(components[0])
	if err != nil {
		fmt.Println("Error converting vertex to int")
	}
	normal, err := strconv.Atoi(components[1])
	if err != nil {
		fmt.Println("Error converting normal to int")
	}

	return vertex, normal, err
}

func getFaceNormal(normal1 rl.Vector3, normal2 rl.Vector3, normal3 rl.Vector3) rl.Vector3 {
	xNormal := (normal1.X + normal2.X + normal3.X) / 3
	yNormal := (normal1.Y + normal2.Y + normal3.Y) / 3
	zNormal := (normal1.Z + normal2.Z + normal3.Z) / 3

	fmt.Sprintf("x normal %f, ynormal %f, znormal %f", xNormal, yNormal, zNormal)

	return rl.NewVector3(xNormal, yNormal, zNormal)
}

func makeNormalColourCache(faces []modelFace) map[modelFace]rl.Color {
	faceColours := make(map[modelFace]rl.Color)
	for _, face := range faces {
		faceColour := rl.NewColor(122, 122, 122, 255)

		faceColour.R = faceColour.R + uint8(100*face.Normal.Y)
		faceColour.G = faceColour.G + uint8(100*face.Normal.Y)
		faceColour.B = faceColour.B + uint8(100*face.Normal.Y)

		faceColours[face] = faceColour
	}
	return faceColours
}
