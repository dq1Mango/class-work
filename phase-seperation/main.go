package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/AlexEidt/Vidio"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/montanaflynn/stats"
	// "slices"
)

type SiteState int

const (
	Oil SiteState = iota
	Water
	// Visited
	// Immune
)

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
	table Grid
	grids []Grid
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

// func heart_equation_derive(x, y float64) float64 {
// 	return 3 * math.Pow(x * x + y * y - 1, 2)
//
// }

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

// func random_step(r *rand.Rand) Point {
//
// 	value := r.Float64()
//
// 	if value < 0.25 {
// 		return Point{x: 1, y: 0}
// 	} else if value < 0.5 {
// 		return Point{x: -1, y: 0}
// 	} else if value < 0.75 {
// 		return Point{x: 0, y: 1}
// 	} else {
// 		return Point{x: 0, y: -1}
// 	}
//
// }

// func circleEquation(x float64) float64 {
// 	return math.Sqrt(1- x*x)
// }

func (g Grid) draw_circle(center Point, radius int, value SiteState) {

	circleEquation := func(x float64) float64 {
		return math.Sqrt(float64(radius*radius) - x*x)
	}

	for x := -radius; x <= radius; x++ {
		height := round(circleEquation(float64(x)))

		for y := -height; y <= height; y++ {
			point := add_points(Point{x: x, y: y}, center)
			// fmt.Println(point)
			*g.index(point) = value
		}
	}
}

// generates a grid with the yinying pattern
func gen_yinyang_grid(size int) Grid {
	yinyang := gen_grid(size)

	radius := size / 2

	for y := -radius; y <= radius; y++ {
		for x := -radius; x < 0; x++ {
			*yinyang.index(Point{x, y}) = Oil
		}
		for x := 0; x <= radius; x++ {
			*yinyang.index(Point{x, y}) = Water
		}
	}

	yinyang.draw_circle(Point{x: 0, y: -radius / 2}, radius/2, SiteState(0))
	yinyang.draw_circle(Point{x: 0, y: radius / 2}, radius/2, SiteState(1))

	yinyang.draw_circle(Point{x: 0, y: -radius / 2}, radius/6, SiteState(1))
	yinyang.draw_circle(Point{x: 0, y: radius / 2}, radius/6, SiteState(0))

	return yinyang
}

func init_model(size int, ticks int, frames int, r *rand.Rand) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	// grid_type := "normal"
	// grid_type := "yinyang"

	// wether to use the yinyang pattern or not
	// (this should probably be configurable from cmdline flags)
	yinyang := true

	var grid Grid
	if !yinyang {
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
	} else {
		grid = gen_yinyang_grid(size)
		for _, row := range grid {
			for j := range row {
				row[j] = row[j] ^ 1
			}
		}

		// heart_radius := 30.0
		// grid = gen_heart_grid(size, heart_radius)
	}

	table := gen_yinyang_grid(size)

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

	// currently this is set to use the table model,
	// but this should also be configurable
	if *m.table.index(point) != state {
		enthalpy++
	}

	return enthalpy
}

// calculates the enthalpy of a point
func (m *Model) calcEnthalpy(point Point) int {
	return m.calcTheoreticalEnthalpy(point, *m.index(point))
}

func (m *Model) getNonNeihbors(point Point) []Point {
	var nonNeighbors []Point

	politics := *m.grid.index(point)

	for _, step := range CARDINALS {
		new_point := add_points(point, step)
		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) != politics {
			nonNeighbors = append(nonNeighbors, new_point)
		}
	}

	return nonNeighbors

}

func (m *Model) randomPoint(r *rand.Rand) Point {
	radius := m.size / 2

	return Point{x: r.Intn(m.size) - radius, y: r.Intn(m.size) - radius}
}

func (m *Model) shouldSwitch(point Point, r *rand.Rand) ([]Point, bool) {
	nonNeihbors := m.getNonNeihbors(point)

	slope := 0.25

	probability := float64(len(nonNeihbors)) * slope

	return nonNeihbors, r.Float64() < probability
}

// dont worry abt this
func (m *Model) tick(r *rand.Rand) {

	selected := m.randomPoint(r)
	state := *m.grid.index(selected)
	// start := time.Now()

	if choices, should := m.shouldSwitch(selected, r); should == true {

		choice := r.Intn(len(choices))
		new_point := choices[choice]

		*m.index(selected) = *m.grid.index(new_point)
		*m.grid.index(new_point) = state

		if m.time%m.inc == 0 {
			m.grids = append(m.grids, m.grid.clone())

		}
		m.time++
	}

}

// func (m *Model) enthalpicSwitch(point Point, r *rand.Rand) ([]Point, bool) {
//
// 	neihbors := m.countNeibors(point)
//
// 	var options []Point
//
// 	politics := *m.grid.index(point)
//
// 	for _, step := range CARDINALS {
// 		new_point := add_points(point, step)
// 		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) != politics {
// 			nonNeighbors = append(nonNeighbors, new_point)
// 		}
// 	}
//
// 	return nonNeighbors
//
// 	nonNeihbors := m.getNonNeihbors(point)
//
// 	slope := 0.25
//
// 	probability := float64(len(nonNeihbors)) * slope
//
// 	return nonNeihbors, r.Float64() < probability
// }

// dont worry abt this either
func (m *Model) logicalTick(r *rand.Rand) {

	selected := m.randomPoint(r)
	state := *m.grid.index(selected)
	// start := time.Now()

	if choices, should := m.shouldSwitch(selected, r); should == true {

		choice := r.Intn(len(choices))
		new_point := choices[choice]

		*m.index(selected) = *m.grid.index(new_point)
		*m.grid.index(new_point) = state

		if m.time%m.inc == 0 {
			m.grids = append(m.grids, m.grid.clone())

		}
		m.time++
	}

}

// this is what actually changes the state
func (m *Model) balazsTick(r *rand.Rand) {

	// picks a random point
	selected := m.randomPoint(r)
	state := *m.grid.index(selected)
	// start := time.Now()

	// calculates the enthalpy of the random point
	enthalpy := m.calcEnthalpy(selected)

	if enthalpy > 0 {

		var newPoint Point

		success := false
		// try 100 times to pick a new point which would decresase or not change the current enthalpy
		for range 100 {
			newPoint = m.randomPoint(r)
			// 														//																	this <= is CRITICAL!!!
			if *m.index(newPoint) != state && m.calcTheoreticalEnthalpy(newPoint, state) <= enthalpy {
				// if *m.index(newPoint) != state {
				*m.index(selected) = *m.grid.index(newPoint)
				*m.grid.index(newPoint) = state

				success = true
				break
			}
		}

		// if we fail to pick such a point we try again with a new inital point
		if !success {
			m.fails++
			return
			// fmt.Println("we failed")
		}

		if m.time%m.inc == 0 {
			clone := m.grid.clone()
			m.grids = append(m.grids, clone)
		}

		m.time++
	}

}

// runs one *model* to completion
func (m *Model) run_trial(r *rand.Rand) Data {

	// run the model for the specified number of ticks (m.ticks)
	for m.time < m.ticks {
		// model.tick(r)
		// m.logicalTick(r)

		// tick the model
		m.balazsTick(r)
	}

	// fmt.Println("we failed:", m.fails, "times")

	return make([]DataPoint, 0)

}

// ignore all these functions that r not rly used
func run_simulation() stats.Series {
	size := 201
	num_points := 100.0

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	series := make(stats.Series, 0, int(num_points))

	for p := 0.01; p < 1.0; p += 0.01 {
		// p := p / num_points

		clear_line()
		fmt.Print("this much done: ", p*100, "%")

		model := init_model(size, 1000, 100, r)

		_ = model.run_trial(r)

	}

	// pretty_picture(model, "testing", 5)
	return series

}

func testing() {

	grid := gen_grid(3)

	*grid.index(Point{x: 0, y: 0}) = 1
	// fmt.Println(grid.index(Point{x: 0, y: 0}) == &grid[1][1])
	// data = unordered_remove(data, 1)

	clone := grid.clone()
	*grid.index(Point{x: 0, y: 0}) = 2

	fmt.Println(clone)
}

type DataPoint struct {
	radius int
	filled int
}

type Data []DataPoint

type Arguments struct {
	file *string
	// operation *string
	chart  *string
	output *string
}

func parse_args() Arguments {
	// args := os.Args[1:]
	// if len(args) >= 2 {
	// 	if args[0] == "--file" || args[0] == "-f" {
	// 		return Arguments{
	// 			file:    args[1],
	// 			is_file: true,
	// 		}
	// 	}
	// }
	//
	// return Arguments{
	// 	file:    "",
	// 	is_file: false,
	// }

	args := Arguments{
		file: flag.String("file", "", "path to data file"),
		// operation: flag.String("op", "", "operation to perform"),
		chart:  flag.String("chart", "", "type of chart to make"),
		output: flag.String("out", "", "prefix of output files"),
	}

	flag.Parse()

	return args
}

func (d *Data) toSeries() stats.Series {
	series := make([]stats.Coordinate, 0, len(*d))

	for _, point := range *d {
		series = append(series, stats.Coordinate{X: float64(point.radius), Y: float64(point.filled)})
	}

	return series
}

// START HERE (duh)
func main() {
	// testing()
	// return

	var data stats.Series

	args := parse_args()

	if *args.output == "" {
		panic("no output file name specified")
	}

	one_trial(*args.output)
	return

	// only run one_trial for testing purposes

	if file := *args.file; file != "" {
		json_data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(json_data, &data)

	} else {
		// data = run_simulation()
		// size := 201
		// p := 0.01
		// distance := 20

		// r := rand.New(rand.NewSource(time.Now().UnixNano()))
		// model := init_model(size, p, distance)

		data = run_simulation()

		// dataa := model.run_trial(r)
		// data := logLog(dataa.toSeries())

		for _, d := range data {
			fmt.Println(d.X)
		}

		for _, d := range data {
			fmt.Println(d.Y)
		}
		// fmt.Println(intercept, gradient)

		file, err := os.Create(*args.output + "data.json")

		if err != nil {
			panic(err)
		}

		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(data)

	}

	switch *args.chart {

	default:
		fmt.Println("uknown chart type: ", *args.chart)
	}

}

func Pointer[T any](t T) *T {
	return &t
}

func find_max(data []opts.Chart3DData) float32 {
	greatest := float64(0)
	for _, value := range data {
		v := value.Value[2]
		switch v := v.(type) {
		case float64:
			if v > greatest {
				greatest = v
			}
		default:
			panic("howd that get in here")
		}

	}

	return float32(greatest)
}

func make_3d_chart(data []opts.Chart3DData) {
	surface := charts.NewSurface3D()

	surface.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Random Walker Data",
		}),
		charts.WithXAxis3DOpts(opts.XAxis3D{
			Name: "Infection Chance (%)",
		}),
		charts.WithYAxis3DOpts(opts.YAxis3D{
			Name: "# of Steps",
		}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{
			Name: "Average Spread Rate",
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Max:        find_max(data),
			Min:        0,
			Dimension:  "z",
			Calculable: Pointer(true),
			// Range:      []float32{1, 100},
			InRange: &opts.VisualMapInRange{
				Color: []string{"blue", "red"},
			},
		}),
	)

	surface.AddSeries("", data)
	// SetSeriesOptions(
	// 	charts.WithShading("realistic"),
	// )

	f, _ := os.Create("surface.html")
	defer f.Close()
	surface.Render(f)
	err := exec.Command("xdg-open", "surface.html").Run()
	if err != nil {
		fmt.Println("couldnt open chart:", err)
	}
	exec.Command("hyprctl", "dispatch", "workspace", "2").Run()

}

// func cast_to_float(input []any) ([]float64, error) {
// 	output := make([]float64, len(input))
//
// 	for i, elem := range input {
// 		switch e := elem.(type) {
// 		case float64:
// 			output[i] = e
//
// 		default:
// 			return []float64{}, fmt.Errorf("not of type float64")
// 		}
// 	}
//
// 	return output, nil
// }

func one_trial(filename string) {
	// set all the parameters:
	size := 101
	ticks := int(3e4)

	// calculate some values that make a nice video
	fps := 15
	vid_time := 10 // in seconds
	frames := fps * vid_time

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// make the model
	model := init_model(size, ticks, frames, r)

	// run a trial
	_ = model.run_trial(r)

	// for _, point := range data {
	// 	fmt.Println(point.radius)
	// }
	// for _, point := range data {
	// 	fmt.Println(point.filled)
	// }

	// pretty_picture(model.grid, filename, true)
	model.makeVid(fps, filename)

}

func calc_color(percent float64) color.NRGBA {
	RStart, REnd := 255.0, 1.0

	// WOW !!! great code
	return color.NRGBA{
		R: uint8(RStart + (REnd-RStart)*percent),
		A: 255,
		// A: uint8(int(color_start.A) + round(float64(color_end.A-color_start.A)*percent)),
	}

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

	// file, err := os.Create(name + "-data.png")
	//
	// if err != nil {
	// 	panic(err)
	// }
	//
	// defer file.Close()
	//
	// encoder := json.NewEncoder(file)
	// encoder.SetIndent("", "  ")
	// // encoder.Encode(model)
	// encoder.Encode(struct {
	// 	P          float64 `json:"p"`
	// 	Size       int     `json:"size"`
	// 	Infected   int     `json:"infected"`
	// 	Population int     `json:"population"`
	// }{
	// 	model.p,
	// 	model.size,
	// 	model.infected,
	// 	model.people,
	// })

	// exec.Command("chafa", "image.png").Run()
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

// func (m *Model)

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
