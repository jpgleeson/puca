package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
		cleanedLine := strings.TrimLeftFunc(line, unicode.IsSpace)
		lineComponents := strings.Split(cleanedLine, " ")
		lineType := lineComponents[0]
		switch lineType {
		case "facet":
			normal = rl.NewVector3(stringToFloat32(lineComponents[2]), stringToFloat32(lineComponents[3]), stringToFloat32(lineComponents[4]))
		case "vertex":
			vertex := rl.NewVector3(stringToFloat32(lineComponents[1]), stringToFloat32(lineComponents[3]), stringToFloat32(lineComponents[2]))
			currentVertices = append(currentVertices, vertex)
		case "endfacet":
			face := modelFace{
				Point1: currentVertices[0],
				Point2: currentVertices[2],
				Point3: currentVertices[1],
				Normal: normal,
			}
			faces = append(faces, face)
			currentVertices = make([]rl.Vector3, 0, 3)
		default:
		}
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	return faces, nil
}

func LoadBinarySTL(modelPath string) ([]modelFace, error) {

	faces := make([]modelFace, 0)

	file, err := os.Open(modelPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	// Header - 80 bytes. Normally empty
	buffer := make([]byte, 80)

	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Fatal(err)
		return nil, err
	}

	buffer = make([]byte, 4)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Fatal(err)
		return nil, err
	}
	numberOfTriangles := binary.LittleEndian.Uint32(buffer)

	for i := 0; i < int(numberOfTriangles); i++ {
		buffer = make([]byte, 50)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
			return nil, err
		}
		normalSlice := buffer[0:11]
		vertex1Slice := buffer[12:24]
		vertex2Slice := buffer[24:36]
		vertex3Slice := buffer[36:48]
		vertexSlices := [][]byte{
			vertex1Slice,
			vertex2Slice,
			vertex3Slice,
		}
		// I don't care about this right now
		// attributeByteCount := buffer[48:49]

		normalComponents := make([]float32, 3)
		for j := 0; j < 3; j++ {
			buf := bytes.NewReader(normalSlice[j*4 : (j+1)*4])
			err := binary.Read(buf, binary.LittleEndian, &normalComponents[j])
			if err != nil {
				log.Fatal("binary.Read failed:", err)
			}
		}
		normal := rl.NewVector3(normalComponents[0], normalComponents[1], normalComponents[2])

		vertices := make([]rl.Vector3, 3)

		for index, slice := range vertexSlices {
			vertex := make([]float32, 3)
			for j := 0; j < 3; j++ {
				buf := bytes.NewReader(slice[j*4 : (j+1)*4])
				err := binary.Read(buf, binary.LittleEndian, &vertex[j])
				if err != nil {
					log.Fatal("binary.Read failed:", err)
				}
			}
			vertices[index] = rl.NewVector3(vertex[0], vertex[2], vertex[1])
		}

		faces = append(faces, modelFace{
			Point1: vertices[0],
			Point2: vertices[2],
			Point3: vertices[1],
			Normal: normal,
		})
	}

	return faces, nil
}
