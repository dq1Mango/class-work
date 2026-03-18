package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/AlexEidt/Vidio"
	// "slices"
)

type SiteState int

const (
	Oil SiteState = iota
	Water
	// Visited
	// Immune
)

const ATTEMPTS = 1

// possible states of sites
// (possibly change this to 1 and -1) for easier enthalpy calculations
var StateColor = map[SiteState]color.NRGBA{
	Oil: {
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	},
	Water: {
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	},
}

// representing a cartesian point on the grid
type Point struct {
	x int
	y int
}

var CARDINALS = []Point{
	{x: 0, y: -1},
	{x: 0, y: 1},
	{x: 1, y: 0},
	{x: -1, y: 0},
}

func mid_point() Point {
	// mid := (size - 1) / 2

	return Point{x: 0, y: 0}
}

func real_mid_point(size int) Point {

	mid := (size - 1) / 2

	return Point{x: mid, y: mid}
}

func add_points(p1, p2 Point) Point {
	return Point{x: p1.x + p2.x, y: p1.y + p2.y}
}

// func (p *Point) add(p2 Point) {
// 	p.x += p2.x
// 	p.y += p2.y
// }

// func (p *Point) flip() {
// 	p.x *= -1
// 	p.y *= -1
// }

func abs(x int) int {
	return max(x, x*-1)
}

func (p *Point) radius() int {
	return max(abs(p.x), abs(p.y))
}

// abstraction of a 2d grid such that sites can be acsessd with cartesian coordinates
type Grid [][]SiteState

func (g Grid) raw_index(point Point) *SiteState {
	return &g[point.y][point.x]
}

func (g Grid) index(point Point) *SiteState {
	real_point := add_points(point, real_mid_point(len(g)))

	return &g[real_point.y][real_point.x]
}

func (g Grid) is_valid_point(point Point) bool {

	radius := len(g) / 2

	if point.x >= -radius && point.x <= radius && point.y >= -radius && point.y <= radius {
		return true
	} else {
		return false
	}
}

func (g Grid) clone() Grid {

	cloned := make(Grid, 0, len(g))

	for _, row := range g {
		new_row := make([]SiteState, 0, len(g))
		for _, value := range row {

			new_row = append(new_row, SiteState(value))
		}
		cloned = append(cloned, new_row)
	}

	return cloned
}

// model holds the different parametes that define the simulation
// as well as the grids themselves
type Model struct {
	grid  Grid
	grids []Grid
	table Grid
	useT  bool
	size  int
	time  int
	ticks int
	fails int
	inc   int
}

func (m *Model) index(point Point) *SiteState {
	return m.grid.index(point)
}

// generates an empty grid of dimensions *size*
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

func round(x float64) int {
	return int(math.Round(x))
}

func clear_screen() {
	print("\u001b[2J")
}

func clear_line() {
	print("\u001b[2K")
	print("\r")
}

func init_model(size int, ticks int, useTable bool, frames int, r *rand.Rand) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	var grid Grid
	grid = gen_grid(size)

	cutoff := 0.5

	// sets each site to a random state
	for i := range grid {
		for j := range grid {
			if r.Float64() < cutoff {
				grid[i][j] = Oil
			} else {
				grid[i][j] = Water
			}
		}
	}

	// make a table if we need one
	var table Grid
	if useTable {
		panic("cannot use table in this minimal version")
		// table = gen_yinyang_grid(size)
	}

	// pretty_picture(table, "yinyang", true)
	// panic("done")

	model := Model{
		grid:  grid,
		table: table,
		grids: []Grid{grid.clone()},
		size:  size,
		time:  0,
		ticks: ticks,
		inc:   ticks / frames,
	}

	return model

}

// calculates the enthalpy a siteState would have at a certain point
func (m *Model) calcTheoreticalEnthalpy(point Point, state SiteState) int {
	enthalpy := 0

	for _, step := range CARDINALS {
		new_point := add_points(point, step)
		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) != state {
			enthalpy++
		}
	}

	// if we have a table factor it in to the enthalpy calculation aswell
	if m.useT {
		if *m.table.index(point) != state {
			enthalpy++
		}
	}

	return enthalpy
}

// calculates the enthalpy of a point
func (m *Model) calcEnthalpy(point Point) int {
	return m.calcTheoreticalEnthalpy(point, *m.index(point))
}

func (m *Model) randomPoint(r *rand.Rand) Point {
	radius := m.size / 2

	return Point{x: r.Intn(m.size) - radius, y: r.Intn(m.size) - radius}
}

// this is what actually changes the state
func (m *Model) tick(r *rand.Rand) {

	// picks a random point
	selected := m.randomPoint(r)
	state := *m.grid.index(selected)
	// start := time.Now()

	// calculates the enthalpy of the random point
	enthalpy := m.calcEnthalpy(selected)

	if enthalpy > 0 {

		var newPoint Point

		// success := false
		// try 'ATTEMPTS' times to pick a new point which would decresase or not change the current enthalpy
		for range ATTEMPTS { //
			newPoint = m.randomPoint(r)
			// 														//			this <= is CRITICAL!!!
			if *m.index(newPoint) != state &&
				m.calcTheoreticalEnthalpy(newPoint, state) < enthalpy {
				// if *m.index(newPoint) != state {
				*m.index(selected) = *m.grid.index(newPoint)
				*m.grid.index(newPoint) = state

				// success = true
				break
			}
		}

		// if we fail to pick such a point we try again with a new inital point
		// if !success {
		// 	m.fails++
		// 	return
		// 	// fmt.Println("we failed")
		// }

		if m.time%m.inc == 0 {
			clone := m.grid.clone()
			m.grids = append(m.grids, clone)
		}

		m.time++
	}

}

// runs one *model* to completion
func (m *Model) run_trial(r *rand.Rand) {

	precision := m.ticks / 100
	progress := -1
	// run the model for the specified number of ticks (m.ticks)
	for m.time < m.ticks {

		currentProgress := m.time / precision
		if currentProgress > progress {
			progress = currentProgress
			clear_line()
			fmt.Printf("This much done: %d", progress)
		}

		// model.tick(r)
		// m.logicalTick(r)

		// tick the model
		m.tick(r)
	}

	// fmt.Println("we failed:", m.fails, "times")

}

type Arguments struct {
	output *string
}

func parse_args() Arguments {

	args := Arguments{
		output: flag.String("out", "", "prefix of output files"),
	}

	flag.Parse()

	return args
}

func main() {
	args := parse_args()

	if *args.output == "" {
		panic("no output file name specified")
	}

	filename := *args.output
	// set all the parameters:
	size := 101
	ticks := int(2e6)
	useTable := false

	// calculate some values that make a nice video
	fps := 15
	vid_time := 10 // in seconds
	frames := fps * vid_time

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// make the model
	model := init_model(size, ticks, useTable, frames, r)

	// run a trial
	model.run_trial(r)

	model.makeVid(fps, filename)

}

func pretty_picture(grid Grid, name string, save bool) *image.NRGBA {

	size := len(grid)
	screen_size := 960

	scale := screen_size / size

	// cropped := model.size - model.distance*2

	img := image.NewNRGBA(image.Rect(0, 0, size*scale, size*scale))

	// for y, row := range grid[model.distance : model.size-model.distance] {
	// 	for x, value := range row[model.distance : model.size-model.distance] {
	for y, row := range grid {
		for x, value := range row {

			var color color.Color
			if value == Oil {
				color = StateColor[Oil]
			} else {
				color = StateColor[Water]
				// color = calc_color(float64(grid[x][y]) / float64(model.infected))
			}

			for i := range scale {
				for j := range scale {
					img.Set(x*scale+j, y*scale+i, color)
				}
			}
		}
	}

	if save {
		f, err := os.Create(name + ".png")
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

	return img

}

func (m *Model) makeVid(fps int, name string) {

	// start := time.Now()

	first := pretty_picture(m.grids[0], "", false)

	bounds := first.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// options := vidio.Options{FPS: float64(fps), Loop: 0, Delay: 1000}
	options := vidio.Options{FPS: float64(fps)}
	gif, _ := vidio.NewVideoWriter(name+".mp4", w, h, &options)
	defer gif.Close()

	// prettyTime := 0.0

	for i := range m.grids {

		// prettyStart := time.Now()
		img := pretty_picture(m.grids[i], name+"-"+fmt.Sprintf("%d", i), false).Pix
		// prettyTime += time.Since(prettyStart).Seconds()
		gif.Write(img)
	}

	// elapsedTime := time.Since(start)
	// fmt.Printf("elapsed time: %.2f\n", elapsedTime.Seconds())
	// fmt.Println(prettyTime)
}
