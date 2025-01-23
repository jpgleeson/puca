package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {

	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1280, 928, "puca")
	rl.SetTargetFPS(60)
	rl.Viewport(0, 0, 1280, 928)

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

	var mesh rl.Mesh
	material := rl.LoadMaterialDefault()

	shaderLoc := rl.GetShaderLocation(material.Shader, "viewPos")

	material.Maps.Color = rl.NewColor(180, 180, 180, 255) // Neutral gray
	material.Maps.Color = rl.NewColor(255, 255, 2, 25)    // White specular
	material.Maps.Color = rl.NewColor(128, 128, 255, 255) // Normal map
	material.Shader.Locs = &shaderLoc

	mesh, modelLoaded = loadModel(modelPath)

	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {

		if rl.IsFileDropped() {
			fileName := rl.LoadDroppedFiles()[0]
			rl.UnloadMesh(&mesh)
			mesh, modelLoaded = loadModel(fileName)
			if modelLoaded {
				modelPath = fileName
			}
			rl.UnloadDroppedFiles()
		}

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

		if rl.IsKeyDown(rl.KeyW) {
			camera.Target.Y += 0.1
		}
		if rl.IsKeyDown(rl.KeyS) {
			camera.Target.Y -= 0.1
		}

		camera.Position.X = distance * float32(math.Sin(theta)*math.Cos(phi))
		camera.Position.Y = distance * float32(math.Sin(phi))
		camera.Position.Z = distance * float32(math.Cos(theta)*math.Cos(phi))

		rl.BeginDrawing()
		rl.ClearBackground(rl.LightGray)
		rl.BeginMode3D(camera)
		rl.DrawGrid(40, 10.0)
		if modelLoaded {
			rl.DrawMesh(mesh, material, rl.MatrixIdentity())
		}
		rl.EndMode3D()
		rl.DrawText(modelPath, 10, 10, 12, rl.Black)
		rl.EndDrawing()
	}
}

func loadModel(modelPath string) (rl.Mesh, bool) {
	var modelTriangles []modelFace
	var err error
	modelLoaded := true

	splitStr := strings.Split(modelPath, ".")
	fileFormat := splitStr[len(splitStr)-1]

	switch fileFormat {
	case "stl":
		isAscii, err := isASCIISTL(modelPath)
		if err != nil {
			modelLoaded = false
			modelPath = "Error checking if stl is ascii or binary."
		}
		if isAscii {
			modelTriangles, err = LoadASCIISTL(modelPath)
			if err != nil {
				modelLoaded = false
				modelPath = "Error loading model."
			}
		} else {
			modelTriangles, err = LoadBinarySTL(modelPath)
			if err != nil {
				modelLoaded = false
				modelPath = "Error loading model."
			}
		}
	case "obj":
		modelTriangles, err = loadObjText(modelPath)
		if err != nil {
			modelLoaded = false
			modelPath = "Error loading model."
		}
	}

	var meshVertices []float32
	var meshNormals []float32
	var modelPoints []rl.Vector3
	for _, face := range modelTriangles {
		meshVertices = append(meshVertices, face.Point1.X)
		meshVertices = append(meshVertices, face.Point1.Y)
		meshVertices = append(meshVertices, face.Point1.Z)
		meshVertices = append(meshVertices, face.Point2.X)
		meshVertices = append(meshVertices, face.Point2.Y)
		meshVertices = append(meshVertices, face.Point2.Z)
		meshVertices = append(meshVertices, face.Point3.X)
		meshVertices = append(meshVertices, face.Point3.Y)
		meshVertices = append(meshVertices, face.Point3.Z)

		meshNormals = append(meshNormals, 100*-face.Normal.X)
		meshNormals = append(meshNormals, 100*face.Normal.Y)
		meshNormals = append(meshNormals, face.Normal.Z)

		modelPoints = append(modelPoints, face.Point2)
		modelPoints = append(modelPoints, face.Point2)
		modelPoints = append(modelPoints, face.Point3)
	}
	boundingBox := CalculateBoundingBox(modelPoints)
	xSpan := math.Abs(float64(boundingBox.Min.X)) + math.Abs(float64(boundingBox.Max.X))
	ySpan := math.Abs(float64(boundingBox.Min.Y)) + math.Abs(float64(boundingBox.Max.Y))
	zSpan := math.Abs(float64(boundingBox.Min.Z)) + math.Abs(float64(boundingBox.Max.Z))

	maxSpan := math.Max(xSpan, ySpan)
	maxSpan = math.Max(maxSpan, zSpan)

	if maxSpan > 200 {
		// clamping to 100 units in size
		factor := math.Floor(maxSpan / 200)
		if factor != 0 {
			ScaleModel(modelTriangles, float32(1/factor))
		}
	}

	mesh := rl.Mesh{
		TriangleCount: int32(len(modelTriangles)),
		VertexCount:   int32(len(modelPoints)),
	}

	mesh.Vertices = unsafe.SliceData(meshVertices)
	mesh.Normals = unsafe.SliceData(meshNormals)

	rl.UploadMesh(&mesh, false)

	return mesh, modelLoaded
}

func ScaleModel(faces []modelFace, scale float32) []modelFace {
	for index, face := range faces {
		face.Point1 = ScaleVector3(face.Point1, scale)
		face.Point2 = ScaleVector3(face.Point2, scale)
		face.Point3 = ScaleVector3(face.Point3, scale)
		faces[index] = face
	}
	return faces
}

func isASCIISTL(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(strings.ToLower(line), "solid"), nil
}

type BoundingBox struct {
	Min, Max rl.Vector3
}

func CalculateBoundingBox(objects []rl.Vector3) BoundingBox {
	if len(objects) == 0 {
		return BoundingBox{}
	}

	min := rl.Vector3{X: objects[0].X, Y: objects[0].Y, Z: objects[0].Z}
	max := rl.Vector3{X: objects[0].X, Y: objects[0].Y, Z: objects[0].Z}

	// gross
	for _, obj := range objects {
		min.X = float32(math.Min(float64(min.X), float64(obj.X)))
		min.Y = float32(math.Min(float64(min.Y), float64(obj.Y)))
		min.Z = float32(math.Min(float64(min.Z), float64(obj.Z)))

		max.X = float32(math.Max(float64(max.X), float64(obj.X)))
		max.Y = float32(math.Max(float64(max.Y), float64(obj.Y)))
		max.Z = float32(math.Max(float64(max.Z), float64(obj.Z)))
	}

	return BoundingBox{Min: min, Max: max}
}

type modelFace struct {
	Point1 rl.Vector3
	Point2 rl.Vector3
	Point3 rl.Vector3
	Normal rl.Vector3
}

func ScaleVector3(vector rl.Vector3, scale float32) rl.Vector3 {
	vector.X = vector.X * scale
	vector.Y = vector.Y * scale
	vector.Z = vector.Z * scale
	return vector
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
			fmt.Printf("Unknown line type %s\n", line)
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
	var vertex, normal int

	vertex, err := strconv.Atoi(components[0])
	if err != nil {
		fmt.Println("Error converting vertex to int")
	}
	if len(components) > 1 {
		normal, err = strconv.Atoi(components[1])
		if err != nil {
			fmt.Println("Error converting normal to int")
		}
	}

	return vertex, normal, err
}

func getFaceNormal(normal1 rl.Vector3, normal2 rl.Vector3, normal3 rl.Vector3) rl.Vector3 {
	xNormal := (normal1.X + normal2.X + normal3.X) / 3
	yNormal := (normal1.Y + normal2.Y + normal3.Y) / 3
	zNormal := (normal1.Z + normal2.Z + normal3.Z) / 3
	return rl.NewVector3(xNormal, yNormal, zNormal)
}
