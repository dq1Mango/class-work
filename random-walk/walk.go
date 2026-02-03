package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"time"

	textColor "github.com/fatih/color"
	// "github.com/go-echarts/go-echarts/v2/charts"
	// "github.com/go-echarts/go-echarts/v2/opts"
)

// some clever stack overflow code i found
type SiteState int

const (
	Susceptible SiteState = iota
	Removed
)

var StateColor = map[SiteState]color.NRGBA{
	Susceptible: {
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	},
	Removed: {
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	},
}

type Grid [][]SiteState

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

type Walker struct {
	location Point
	ttl      int
}

type Model struct {
	grid     Grid
	walkers  []Walker
	p        float64
	tau      int
	people   int
	infected int
}

func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// // this isnt dumb at all
// type Walkers []Walker
//
// func (w Walkers) remove(i int) {
// 	w[i] = w[len(w)-1]
// 	w = w[:len(w)-1]
// }

func gen_grid(size int) Grid {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	grid := make(Grid, size)

	for row := range grid {
		grid[row] = make([]SiteState, size)
	}

	return grid
}

func (g Grid) index(point Point) *SiteState {
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
				textColor.Set(textColor.FgBlack)
			case 1:
				textColor.Set(textColor.FgWhite)

			}
			fmt.Print(block)
		}
		fmt.Println()
	}
	textColor.Set(textColor.FgWhite)
}

func clear_screen() {
	print("\u001b[2J")
}

func random_step(r *rand.Rand) Point {

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

func init_model(size int, p float64, tau int) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	mid := (size - 1) / 2

	middle := Point{row: mid, col: mid}

	grid := gen_grid(size)

	*grid.index(middle) = 1

	walkers := make([]Walker, 1)
	walkers[0] = Walker{
		location: middle,
		ttl:      tau,
	}

	model := Model{
		grid:     grid,
		walkers:  walkers,
		p:        p,
		people:   size * size,
		infected: 1,
	}

	return model
}

func (m *Model) add_walker(point Point, ttl int) {
	m.walkers = append(m.walkers, Walker{
		location: point,
		ttl:      ttl,
	})
}

func (m *Model) tick(r *rand.Rand) {

	chopping_block := make([]int, 0)

	for index := range m.walkers {

		walker := &m.walkers[index]
		var new_point Point

		// start := time.Now()

		for {
			step := random_step(r)
			new_point = add_points(walker.location, step)
			if m.grid.is_valid_point(new_point) {
				walker.location = new_point
				// m.walkers[index].location = new_point
				break
			}
		}

		// stop := time.Now()
		// fmt.Println("im pretty stupid: ", model.walkers)
		// fmt.Println("getting the new point took this long: ", stop.Sub(start))

		// start = time.Now()
		// r := rand.New(rand.NewSource(time.Now().UnixNano()))
		value := r.Float64()

		if *m.grid.index(new_point) == 0 {
			if value < m.p {
				*m.grid.index(new_point) = 1
				m.add_walker(new_point, m.tau)
				m.infected++
			}
		}

		walker.ttl -= 1
		if walker.ttl == 0 {
			chopping_block = append(chopping_block, index)
		}

		// stop = time.Now()

		// fmt.Println("maybe spreading took this long: ", stop.Sub(start))

	}

	// this isnt great but maybe its fine
	for _, block := range chopping_block {
		m.walkers = remove(m.walkers, block)
	}

}

func testing() {
	data := []int{0, 1, 2, 3}

	// for _, idk := range data {
	// 	idk = 67
	// }

	data = remove(data, 1)

	fmt.Println(data)
}

func main() {
	testing()
	return

	fmt.Println("is this how go works?")

	size := 1001
	p := 0.2
	tau := 10

	tps := 5

	var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	fmt.Println(delay)
	fmt.Println(1.0 / float64(tps))

	model := init_model(size, p, tau)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	i := 0
	for model.infected < model.people {
		// fmt.Println("iter: ", i, ", infected: ", float64(model.infected)/float64(model.people))

		// start := time.Now()
		model.tick(r)
		// stop := time.Now()
		// fmt.Println("im pretty stupid: ", model.walkers)
		// fmt.Println("took this long: ", stop.Sub(start))

		// model.grid.print_grid()
		// time.Sleep(delay)
		// clear_screen()
		i++

		if i > 1000 {
			break
		}

	}

	make_graph(model)

	// textColor.Set(textColor.FgWhite)
	// fmt.Println(delay)
	// model.grid.print_grid()
}

func make_graph(model Model) {

	grid := model.grid

	size := len(grid)

	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	for x := range grid {
		for y := range grid[0] {
			img.Set(x, y, StateColor[grid[x][y]])
		}
	}

	for _, walker := range model.walkers {
		img.Set(walker.location.col, walker.location.row, color.NRGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 255,
		})
	}

	f, err := os.Create("image.png")
	if err != nil {
		// log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		// log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		// log.Fatal(err)
	}
}

// func make_graph(grid Grid) {
// 	chart := charts.NewHeatMap()
// 	// set some global options like Title/Legend/ToolTip or anything else
// 	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
// 		Title:    "My first bar chart generated by go-echarts",
// 		Subtitle: "It's extremely easy to use, right?",
// 	}))
//
// 	// Put data into instance
// 	chart.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"})
// 	// Where the magic happens
// 	f, _ := os.Create("chart.html")
// 	chart.Render(f)
// }
