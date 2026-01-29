package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

type Grid [][]int

type Point struct {
	row int
	col int
}

func add_points(p1, p2 Point) Point {
	return Point{row: p1.row + p2.row, col: p1.col + p2.col}
}

func (p Point) add(p2 Point) {
	p.row += p2.row
	p.col += p2.col
}

type Model struct {
	grid    Grid
	walkers []Point
	p       float64
}

func gen_grid(size int) Grid {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	grid := make(Grid, size)

	for row := range grid {
		grid[row] = make([]int, size)
	}

	return grid
}

func (g Grid) index(point Point) *int {
	return &g[point.row][point.col]
}

func (g *Grid) is_valid_point(point Point) bool {
	size := len(*g)

	if point.row >= 0 && point.row < size && point.col >= 0 && point.col < size {
		return true
	} else {
		return false
	}
}

func (g *Grid) print_grid() {
	block := "██"

	for _, row := range *g {
		for _, node := range row {
			switch node {
			case 0:
				color.Set(color.FgBlack)
			case 1:
				color.Set(color.FgWhite)

			}
			fmt.Print(block)
		}
		fmt.Println()
	}
}

func clear_screen() {
	print("\u001b[2J")
}

func random_step() Point {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	value := r.Float64()

	if value < 0.25 {
		return Point{row: 1, col: 0}
	} else if value < 0.5 {
		return Point{row: -1, col: 0}
	} else if value < 0.75 {
		return Point{row: 0, col: 1}
	} else {
		return Point{row: 0, col: -1}
	}

}

func init_model(size int, p float64) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	mid := (size - 1) / 2

	middle := Point{row: mid, col: mid}

	grid := gen_grid(size)

	*grid.index(middle) = 1

	model := Model{
		grid:    grid,
		walkers: []Point{middle},
		p:       p,
	}

	return model
}

func (m *Model) add_walker(point Point) {
	fmt.Println("adding walker")
	m.walkers = append(m.walkers, point)
	fmt.Println(m.walkers)
}

func (m Model) tick() {

	for index, walker := range m.walkers {
		var new_point Point

		for {
			step := random_step()
			new_point = add_points(walker, step)
			if m.grid.is_valid_point(new_point) {
				m.walkers[index] = new_point
				break
			}
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		value := r.Float64()

		if value < m.p {
			*m.grid.index(new_point) = 1
			m.add_walker(new_point)
		}

	}

}

func main() {
	fmt.Println("is this how go works?")

	size := 11
	p := 1.0

	tps := 5

	var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	fmt.Println(delay)
	fmt.Println(1.0 / float64(tps))

	model := init_model(size, p)

	for range 100 {
		clear_screen()
		model.tick()
		model.grid.print_grid()
		time.Sleep(delay)

	}

	color.Set(color.FgWhite)
	fmt.Println(delay)
	// model.grid.print_grid()
}
