package main

import (
	"fmt"
	"math"
)

type Vector struct {
	x float64
	y float64
}

func (v Vector) String() string {
	return fmt.Sprintf("(%.3f, %.3f)", v.x, v.y)
}

func (v *Vector) normalize() {
	magnitude := math.Hypot(v.x, v.y)
	if magnitude == 0 {
		return
	}

	v.x /= magnitude
	v.y /= magnitude
}

func mouseWeightVector(cursorX, cursorY, width, height int) Vector {
	if width <= 0 || height <= 0 {
		return Vector{}
	}

	v := Vector{
		x: float64(width/2 - cursorX),
		y: float64(height/2 - cursorY),
	}

	v.normalize()
	return v
}

func mouseTargetVector(cursorX, cursorY, width, height int) Vector {
	if width <= 0 || height <= 0 {
		return Vector{}
	}

	return Vector{
		x: (float64(cursorX)/float64(width) - 0.5) * 960,
		y: (0.5 - float64(cursorY)/float64(height)) * 960,
	}
}

func attractionForce(from, to Vector) Vector {
	force := Vector{x: to.x - from.x, y: to.y - from.y}
	force.normalize()
	return force
}

func main() {
	width, height := 960, 960
	walker := Vector{x: -120, y: 40}

	samples := []struct {
		label string
		x     int
		y     int
	}{
		{label: "center", x: width / 2, y: height / 2},
		{label: "right", x: int(float64(width) * 0.8), y: height / 2},
		{label: "left", x: int(float64(width) * 0.2), y: height / 2},
		{label: "top", x: width / 2, y: int(float64(height) * 0.2)},
		{label: "bottom", x: width / 2, y: int(float64(height) * 0.8)},
	}

	fmt.Println("Mouse vector demo")
	fmt.Println("walker:", walker)
	fmt.Println()

	for _, sample := range samples {
		mouseTarget := mouseTargetVector(sample.x, sample.y, width, height)
		mouseWeight := mouseWeightVector(sample.x, sample.y, width, height)
		force := attractionForce(walker, mouseTarget)

		fmt.Printf("%-7s cursor=(%3d,%3d) target=%s weight=%s force=%s\n",
			sample.label, sample.x, sample.y, mouseTarget, mouseWeight, force)
	}
}
