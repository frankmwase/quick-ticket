package main

import (
	"math"
	"strings"
)

type Vec3 struct {
	X, Y, Z float64
}

type Edge struct {
	A, B int
}

//rectangular shape
var ticketVertices = []Vec3{
	{-1, -1.5, -0.1}, {1, -1.5, -0.1}, {1, 1.5, -0.1}, {-1, 1.5, -0.1},
	{-1, -1.5, 0.1}, {1, -1.5, 0.1}, {1, 1.5, 0.1}, {-1, 1.5, 0.1},
}

var ticketEdges = []Edge{
	{0, 1}, {1, 2}, {2, 3}, {3, 0},
	{4, 5}, {5, 6}, {6, 7}, {7, 4},
	{0, 4}, {1, 5}, {2, 6}, {3, 7},
}

func rotate(v Vec3, angleX, angleY, angleZ float64) Vec3 {
	// Rotate X
	y := v.Y*math.Cos(angleX) - v.Z*math.Sin(angleX)
	z := v.Y*math.Sin(angleX) + v.Z*math.Cos(angleX)
	v.Y = y
	v.Z = z
	// Rotate Y
	x := v.X*math.Cos(angleY) + v.Z*math.Sin(angleY)
	z = -v.X*math.Sin(angleY) + v.Z*math.Cos(angleY)
	v.X = x
	v.Z = z
	// Rotate Z
	x = v.X*math.Cos(angleZ) - v.Y*math.Sin(angleZ)
	y = v.X*math.Sin(angleZ) + v.Y*math.Cos(angleZ)
	v.X = x
	v.Y = y
	return v
}

func drawLine(screen [][]rune, x0, y0, x1, y1 int, ch rune) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if y0 >= 0 && y0 < len(screen) && x0 >= 0 && x0 < len(screen[0]) {
			screen[y0][x0] = ch
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func RenderWireframe(width, height int, angleX, angleY, angleZ float64) string {
	screen := make([][]rune, height)
	for i := range screen {
		screen[i] = make([]rune, width)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}

	projected := make([]Vec3, len(ticketVertices))
	for i, v := range ticketVertices {
		rotated := rotate(v, angleX, angleY, angleZ)
		// projection
		z := rotated.Z + 4.0
		f := 20.0 / z
		px := rotated.X * f
		py := rotated.Y * f

		// map to screen
		projected[i] = Vec3{
			X: float64(width)/2.0 + px*2.0, // x scaling for terminal font aspect ratio
			Y: float64(height)/2.0 - py,
		}
	}

	for _, edge := range ticketEdges {
		p1 := projected[edge.A]
		p2 := projected[edge.B]
		drawLine(screen, int(p1.X), int(p1.Y), int(p2.X), int(p2.Y), '#')
	}

	var sb strings.Builder
	for _, row := range screen {
		sb.WriteString(string(row) + "\n")
	}
	return sb.String()
}
