package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Grid [][]bool

func gen_grid(size int, p float64) Grid {
	grid := make([][]bool, size)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for row := range grid {
		grid[row] = make([]bool, size)

		for i := range size {
			if r.Float64() <= p {
				grid[row][i] = true
			}
		}
	}

	return grid
}

func disp_grid(grid Grid) {
	for _, row := range grid {
		fmt.Println(row)
	}
}

func main() {

	p := 0.5
	size := 5

	grid := gen_grid(size, p)

	disp_grid(grid)

	println("idk how this works")
}
