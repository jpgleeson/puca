package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func LoadASCIISTL(modelPath string) ([]modelFace, error) {

	faces := make([]modelFace, 0)

	file, err := os.Open(modelPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	currentVertices := make([]rl.Vector3, 0, 3)
	var normal rl.Vector3

	for fileScanner.Scan() {
		line := fileScanner.Text()
		fmt.Println(line)
		cleanedLine := strings.TrimLeftFunc(fileScanner.Text(), unicode.IsSpace)
		lineComponents := strings.Split(cleanedLine, " ")
		lineType := lineComponents[0]
		switch lineType {
		case "solid":
			// Starting an object. Should be the first line. Second component would be the name
		case "facet":
			normal = rl.NewVector3(stringToFloat32(lineComponents[2]), stringToFloat32(lineComponents[3]), stringToFloat32(lineComponents[4]))
		case "outer":
			// start of a triangle
		case "vertex":
			// This is a vertex.)
			vertex := rl.NewVector3(stringToFloat32(lineComponents[1]), stringToFloat32(lineComponents[2]), stringToFloat32(lineComponents[3]))
			currentVertices = append(currentVertices, vertex)
		case "endloop":
			// close a loop
		case "endfacet":
			face := modelFace{
				Point1: currentVertices[0],
				Point2: currentVertices[1],
				Point3: currentVertices[2],
				Normal: normal,
			}
			faces = append(faces, face)
			currentVertices = make([]rl.Vector3, 0, 3)
		case "endsolid":
			// end of an object.
		}
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	return faces, nil
}
