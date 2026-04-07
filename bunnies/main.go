package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	// "math/cmplx"
	"math/rand"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"time"

	"github.com/AlexEidt/Vidio"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/montanaflynn/stats"
	"github.com/scientificgo/fft"
	// "slices"
)

// constants that dictate the simulation
const (
	ATTEMPTS = 1
	SIZE     = 101
	TICKS    = 1e3

	INITIAL_BUNNY = 2
	INITIAL_FOX   = 50

	BUNNY_BABY = 0.1
	CATCH      = 0.5
	FOX_STARVE = 0.1

	RECORD = true
	TPF    = 1 // ticks per frame
	FPS    = 30
	// VID_TIME = 10 // in seconds
)

type SiteState int

// possible states of sites
const (
	Empty SiteState = iota
	Bunny
	Fox
	// Visited
	// Immune
)

// catppuccin color mapping
var StateColor = map[SiteState]color.NRGBA{
	Empty: HexColor("#a6e3a1"),
	Bunny: HexColor("#89b4fa"),
	Fox:   HexColor("#f38ba8"),
}

// var StateColor = map[SiteState]color.NRGBA{
// 	Empty: {R: 0, G: 255, B: 0, A: 255},
// 	Bunny: {
// 		R: 0,
// 		G: 0,
// 		B: 255,
// 		A: 255,
// 	},
// 	Fox: {
// 		R: 255,
// 		G: 0,
// 		B: 0,
// 		A: 255,
// 	},
// }

// hey look at this 'citation injection':

// Source - https://stackoverflow.com/a/77740085
// Posted by zoltron, modified by community. See post 'Timeline' for change history
// Retrieved 2026-03-25, License - CC BY-SA 4.0

// HexColor converts hex color to color.RGBA with "#FFFFFF" format
func HexColor(hex string) color.NRGBA {
	values, _ := strconv.ParseUint(string(hex[1:]), 16, 32)
	return color.NRGBA{
		R: uint8(values >> 16),
		G: uint8((values >> 8) & 0xFF),
		B: uint8(values & 0xFF),
		A: 255,
	}
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

// representing a cartesian point on the grid
type Point struct {
	x int
	y int
}

func (p *Point) radius() int {
	return max(abs(p.x), abs(p.y))
}

func (p *Point) scale(r int) {
	p.x *= r
	p.y *= r
}

var CARDINALS = []Point{
	{x: 0, y: -1},
	{x: 0, y: 1},
	{x: 1, y: 0},
	{x: -1, y: 0},
}

func factorial(n int) int {
	result := 1
	for i := 1; i <= n; i++ {
		result *= i
	}
	return result
}

func heaps(output *[][]int, A []int, k int) {
	A = slices.Clone(A)

	if k == 1 {
		*output = append(*output, A)
		return
	}

	heaps(output, A, k-1)

	for i := range k - 1 {
		if k%2 == 0 {
			A[i], A[k-1] = A[k-1], A[i]
		} else {
			A[0], A[k-1] = A[k-1], A[0]
		}
		heaps(output, A, k-1)
	}

}

func Permutations(n int) [][]int {
	nPickN := factorial(n)
	permutations := make([][]int, 0, nPickN)

	// for i := range permutations {
	// 	permutations[i] = make([]int, 0, n)
	// }

	initial := make([]int, n)
	for i := range n {
		initial[i] = i
	}

	heaps(&permutations, initial, n)

	// fmt.Println(permutations)

	return permutations

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

var PERMUTATIONS = Permutations(4)

// returns the CARDINALS in a random order based on
// precomputed permunations on n = 4
func randomDirections(r *rand.Rand) []Point {
	directions := make([]Point, 0, 4)

	for _, value := range PERMUTATIONS[r.Int63n(24)] {
		directions = append(directions, CARDINALS[value])
	}

	return directions
}

func abs(x int) int {
	return max(x, x*-1)
}

// abstraction of a 2d grid such that sites can be acsessd with cartesian coordinates
type Grid [][]SiteState

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

func (g Grid) raw_index(point Point) *SiteState {
	return &g[point.y][point.x]
}

func (g Grid) index(point Point) *SiteState {
	return g.raw_index(point)
	// real_point := add_points(point, real_mid_point(len(g)))
	//
	// return &g[real_point.y][real_point.x]
}

// with periodic bounds all points are valid ig
// func (g Grid) is_valid_point(point Point) bool { return true }

func (g Grid) is_valid_point(point Point) bool {

	if point.x >= 0 && point.x < len(g) && point.y >= 0 && point.y < len(g) {
		return true
	} else {
		return false
	}

	// radius := len(g) / 2
	//
	// if point.x >= -radius && point.x <= radius && point.y >= -radius && point.y <= radius {
	// 	return true
	// } else {
	// 	return false
	// }
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
	grid    Grid
	grids   []Grid
	size    int
	time    int
	ticks   int
	bunnies []Point
	foxes   []Point
	frames  int
}

// initializizes a model
func init_model(
	size int,
	ticks int,
	frames int,
	r *rand.Rand,
) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	// grid_type := "normal"
	// grid_type := "yinyang"

	// wether to use the yinyang pattern or not
	// (this should probably be configurable from cmdline flags)

	var grid Grid

	grid = gen_grid(size)

	// pretty_picture(table, "yinyang", true)
	// panic("done")

	bunnies := make([]Point, 0, INITIAL_BUNNY)
	foxes := make([]Point, 0, INITIAL_FOX)

	model := Model{
		grid:    grid,
		grids:   []Grid{grid.clone()},
		size:    size,
		time:    1,
		ticks:   ticks,
		bunnies: bunnies,
		foxes:   foxes,
		frames:  frames,
	}

	i := 0
	for i < INITIAL_BUNNY {
		point := model.randomPoint(r)
		if *model.index(point) == Empty {
			*model.index(point) = Bunny
			model.bunnies = append(model.bunnies, point)
			i++
		}
	}

	i = 0
	for i < INITIAL_FOX {
		point := model.randomPoint(r)
		if *model.index(point) == Empty {
			*model.index(point) = Fox
			model.foxes = append(model.foxes, point)
			i++
		}
	}

	return model

}

func (m *Model) index(point Point) *SiteState {
	return m.grid.index(point)
}

func (m *Model) modPoint(point Point) Point {
	return Point{x: remEuclid(point.x, m.size), y: remEuclid(point.y, m.size)}
	// size := len(m.grid)
	// midPoint := real_mid_point(size)
	//
	// realPoint := add_points(midPoint, *point)
	//
	// modded := Point{x: remEuclid(realPoint.x, size), y: remEuclid(realPoint.y, size)}
	// midPoint.scale(-1)
	//
	// *point = add_points(midPoint, modded)
}

// draws a circle ... duh
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

func (m *Model) randomPoint(r *rand.Rand) Point {
	return Point{x: r.Intn(m.size), y: r.Intn(m.size)}

	// radius := m.size / 2
	//
	// return Point{x: r.Intn(m.size) - radius, y: r.Intn(m.size) - radius}
}

// i really dont think i need this anymore
// func (m *Model) lametick(r *rand.Rand) {
//
// 	for {
// 		var selected Point
// 		var state SiteState
//
// 		for {
// 			selected = m.randomPoint(r)
// 			state = *m.index(selected)
// 			if state != Empty {
// 				break
// 			}
// 		}
//
// 		if state == Bunny {
// 			for _, direction := range randomDirections(r) {
// 				newPoint := add_points(selected, direction)
//
// 				if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Empty {
// 					*m.index(newPoint) = SiteState(Bunny)
// 					if r.Float64() < BUNNY_REPRO {
// 						m.bunnies++
// 					} else {
// 						*m.index(selected) = Empty
// 					}
// 					return
// 				}
// 			}
//
// 		} else if state == Fox {
// 			if r.Float64() < FOX_STARVE {
// 				*m.index(selected) = Empty
// 				m.foxes--
// 				return
// 			}
//
// 			for _, direction := range randomDirections(r) {
// 				newPoint := add_points(selected, direction)
//
// 				if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Bunny {
// 					if r.Float64() < CATCH {
// 						*m.index(newPoint) = Fox
// 						m.foxes++
// 						m.bunnies--
// 						return
// 					}
//
// 				}
// 			}
// 			for _, direction := range randomDirections(r) {
// 				newPoint := add_points(selected, direction)
//
// 				if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Empty {
// 					*m.index(newPoint) = Fox
// 					*m.index(selected) = Empty
// 					return
// 				}
// 			}
//
// 		} else {
// 			panic("uknown state")
// 		}
// 	}
// }

// if the order of elements of a slice does not matter,
// this is a blazingly fast alternative to append(arr[:i], arr[i+1:])
func unorderedRemove[T any](arr []T, index int) []T {
	length := len(arr)
	i := index

	if i >= 0 && i < length {
		arr[i], arr[length-1] = arr[length-1], arr[i]
		return arr[:length-1]
	} else {
		panic(fmt.Sprintf("index out of bounds on remove - index: %d, length %d", index, length))
	}
}

// where the secret sauce resides
func (m *Model) tick(r *rand.Rand) {

	// maybe should just round this ...
	deadFoxes := int(float64(len(m.foxes)) * FOX_STARVE)

	for range deadFoxes {
		index := r.Intn(len(m.foxes))
		*m.index(m.foxes[index]) = Empty
		m.foxes = unorderedRemove(m.foxes, index)
	}

	// tick all of the foxes:

	for i, fox := range m.foxes {

		// success := false
		for _, direction := range randomDirections(r) {
			newPoint := add_points(fox, direction)
			newPoint = m.modPoint(newPoint)

			if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Bunny {
				if r.Float64() < CATCH {
					*m.index(newPoint) = Fox
					m.foxes = append(m.foxes, newPoint)

					for index, bunny := range m.bunnies {
						if bunny == newPoint {
							// TODO: should make the map a hash map to advoid this linear search
							m.bunnies = unorderedRemove(m.bunnies, index)

							// 'goto considered harmful' - Dijkstra 1968
							goto foxEnd
							// break
						}
					}

					panic("could not find bunny lol")

					// success = true
					// break
				}
			}
		}
		// if success {
		// 	continue
		// }

		for _, direction := range randomDirections(r) {
			newPoint := add_points(fox, direction)
			newPoint = m.modPoint(newPoint)

			if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Empty {
				*m.index(newPoint) = Fox
				*m.index(fox) = Empty
				m.foxes[i] = newPoint
				break
			}
		}

	foxEnd:
	}

	// tick all of the bunnies

	for i, bunny := range m.bunnies {
		for _, direction := range randomDirections(r) {
			newPoint := add_points(bunny, direction)
			newPoint = m.modPoint(newPoint)

			if m.grid.is_valid_point(newPoint) && *m.index(newPoint) == Empty {

				*m.index(newPoint) = Bunny
				m.bunnies[i] = newPoint

				if bunny == newPoint {
					panic("not different")
				}

				if r.Float64() < BUNNY_BABY {
					m.bunnies = append(m.bunnies, bunny)
				} else {
					*m.index(bunny) = Empty
				}

				break
			}
		}
	}
}

// runs one *model* to completion
func (m *Model) run_trial(r *rand.Rand) Data {

	// precision := m.ticks / 100
	// progress := -1

	data := make(Data, 0, m.frames)
	interval := m.ticks / m.frames

	// run the model for the specified number of ticks (m.ticks)
	for m.time < m.ticks {

		currentProgress := float64(m.time) / float64(m.ticks)
		clear_line()
		fmt.Printf("This much done: %.1f%%", currentProgress*100)

		if m.time%interval == 0 {

			if RECORD {
				clone := m.grid.clone()
				m.grids = append(m.grids, clone)
			}

			// if realEnthalpy < 800 {
			// 	fmt.Println(realEnthalpy)
			// }

			data = append(
				data,
				DataPoint{Time: m.time, Bunnies: len(m.bunnies), Foxes: len(m.foxes)},
			)
		}
		// model.tick(r)
		// m.logicalTick(r)

		// tick the model
		// m.enthalpicTick(r)
		m.tick(r)

		if len(m.bunnies) == 0 {
			panic("bunnies perished")
		} else if len(m.foxes) == 0 {
			panic("foxes perished")
		}

		m.time++
		// tick(m, r)

	}

	fmt.Println("\nDone!")

	// fmt.Println("we failed:", m.fails, "times")

	return data

}

func run_simulation() stats.Series {
	num_points := 100

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	series := make(stats.Series, 0, int(num_points))

	// startTemp := 0.0001
	// endTemp := 0.01
	// // endTemp := 0.1
	// dataPoints := 100

	// for temp := startTemp; temp <= endTemp; temp += (endTemp - startTemp) / float64(dataPoints) {

	for checkers := 1; checkers <= 50; checkers++ {

		// for
		// p := p / num_points

		clear_line()
		fmt.Print("this much done: ", checkers, "%")

		model := init_model(SIZE, TICKS, num_points, r)

		data := model.run_trial(r)

		logged := logLog(data.toSeries())

		_, slope, _ := LinearRegression(logged)

		series = append(series, stats.Coordinate{X: float64(checkers), Y: slope})
	}

	// pretty_picture(model, "testing", 5)
	return series

}

func remEuclid(x, y int) int {
	for x < 0 {
		x += y
	}

	return x % y
}

type DataPoint struct {
	Time    int `json:"time"`
	Bunnies int `json:"Bunnies"`
	Foxes   int `json:"Foxes"`
}

type Data []DataPoint

func (d Data) WriteToCSV(filename string) {
	fmt.Println("wrtingt to this: ", filename)

	var csv string

	for _, point := range d {
		csv += fmt.Sprintf("%d, %d, %d\n", point.Time, point.Bunnies, point.Foxes)
	}

	err := os.WriteFile(filename+".csv", []byte(csv), 0644)
	if err != nil {
		panic(err)
	}
}

func (d Data) WriteToJSON(filename string) {
	file, err := os.Create(filename + ".json")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(d)
}

func (d Data) Fft() ([]complex128, []complex128) {
	bunnies := make([]complex128, len(d))
	foxes := make([]complex128, len(d))

	for i, point := range d {
		bunnies[i] = complex(float64(point.Bunnies), 0)
		foxes[i] = complex(float64(point.Foxes), 0)
	}

	return fft.Fft(bunnies, false), fft.Fft(foxes, false)
}

func WriteToCSV(series stats.Series, filename string) {
	fmt.Println("wrtingt to this: ", filename)

	var csv string

	for _, point := range series {
		csv += fmt.Sprintf("%f, %f\n", point.X, point.Y)
	}

	err := os.WriteFile(filename+".csv", []byte(csv), 0644)
	if err != nil {
		panic(err)
	}
}

func LinearRegression(s stats.Series) (float64, float64, error) {

	if len(s) == 0 {
		return 0, 0, nil
	}

	// Placeholder for the math to be done
	var sum [5]float64

	// Loop over data keeping index in place
	i := 0
	for ; i < len(s); i++ {
		sum[0] += s[i].X
		sum[1] += s[i].Y
		sum[2] += s[i].X * s[i].X
		sum[3] += s[i].X * s[i].Y
		sum[4] += s[i].Y * s[i].Y
	}

	// Find gradient and intercept
	f := float64(i)
	gradient := (f*sum[3] - sum[0]*sum[1]) / (f*sum[2] - sum[0]*sum[0])
	intercept := (sum[1] / f) - (gradient * sum[0] / f)

	return intercept, gradient, nil
}

func logLog(series stats.Series) stats.Series {

	logged := make(stats.Series, 0, len(series))

	for _, point := range series {
		logged = append(logged, stats.Coordinate{X: math.Log(point.X), Y: math.Log(point.Y)})
	}

	return logged
}

func (d *Data) toSeries() stats.Series {
	series := make([]stats.Coordinate, 0, len(*d))

	for _, point := range *d {
		series = append(
			series,
			stats.Coordinate{X: float64(point.Time), Y: float64(point.Bunnies)},
		)
	}

	return series
}

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

func testing() {

	fmt.Println(remEuclid(4, 5))

	grid := gen_grid(3)

	*grid.index(Point{x: 0, y: 0}) = 1
	// fmt.Println(grid.index(Point{x: 0, y: 0}) == &grid[1][1])

	clone := grid.clone()
	*grid.index(Point{x: 0, y: 0}) = 2

	fmt.Println(clone)
}

func main() {
	// testing()
	// return

	var data Data

	args := parse_args()

	if *args.output == "" {
		panic("no output file name specified")
	}

	if file := *args.file; file != "" {
		json_data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(json_data, &data)

	} else {

		// data = run_simulation()
		// data = one_trial()
		// only run one_trial for testing purposes
		data = one_trial(*args.output)

		data.WriteToCSV("data/" + *args.output)

		// for _, d := range data {
		// 	fmt.Println(d.X)
		// }
		//
		// for _, d := range data {
		// 	fmt.Println(d.Y)
		// }
		// fmt.Println(intercept, gradient)

		// file, err := os.Create(*args.output + "data.json")
		//
		// if err != nil {
		// 	panic(err)
		// }
		//
		// defer file.Close()
		//
		// encoder := json.NewEncoder(file)
		// encoder.SetIndent("", "  ")
		// encoder.Encode(data)

	}

	switch *args.chart {

	case "race":
		makeRaceChart("charts/"+*args.output, data)

	case "fft":
		makeFFTChart("charts/"+*args.output, data)

	default:
		fmt.Println("unknown chart type: ", *args.chart)
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

func one_trial(filename string) Data {
	// set all the parameters:

	// calculate some values that make a nice video
	frames := int(TICKS / TPF)
	// fps := frames / VID_TIME
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// make the model
	model := init_model(SIZE, TICKS, frames, r)
	// model := init_model(SIZE, TICKS, temp, useTable, 10, 100, r)

	// run a trial
	// data := model.run_trial(r)
	fmt.Println("running trial...")
	data := model.run_trial(r)

	data.WriteToJSON("data/" + filename)

	// for _, point := range data {
	// 	fmt.Println(point.time)
	// }
	// fmt.Println()
	// for _, point := range data {
	// 	fmt.Println(point.enthalpy)
	// }

	// pretty_picture(model.grid, filename, true)
	// dont add videos to the git tree by default
	// model.makeVid(FPS, "ignore/"+filename)

	return data

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
func pretty_picture(grid Grid, name string) {

	img := grid2png(grid)

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

func grid2png(grid Grid) *image.NRGBA {

	size := len(grid)
	screen_size := 960

	scale := screen_size / size

	// cropped := model.size - model.distance*2

	img := image.NewNRGBA(image.Rect(0, 0, size*scale, size*scale))

	// for y, row := range grid[model.distance : model.size-model.distance] {
	// 	for x, value := range row[model.distance : model.size-model.distance] {
	for y, row := range grid {
		for x, value := range row {

			var color color.Color = StateColor[value]

			for i := range scale {
				for j := range scale {
					img.Set(x*scale+j, y*scale+i, color)
				}
			}
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

// yeah this parallel stunt was pretty much totally useless
// the problem is the slow part is gif.Write(...), which is not parallelizable
// ...lmao
type workOrder struct {
	Grid  *Grid
	Index int
}

type workReceipt struct {
	Img   []uint8
	Index int
}

func pngWorker(orders <-chan workOrder, output chan workReceipt) {
	for {
		order, more := <-orders
		if more {
			output <- workReceipt{Img: grid2png(*order.Grid).Pix, Index: order.Index}
		} else {
			return
		}
	}
}

const WORKERS = 6

func convertGrids(model *Model, output chan []uint8) {
	// images := make([][]byte, len(model.grids))
	orders := make(chan workOrder)
	receipts := make(chan workReceipt)

	for range WORKERS {
		go pngWorker(orders, receipts)
	}

	go func() {
		for i, grid := range model.grids {
			orders <- workOrder{Index: i, Grid: &grid}
		}

		close(orders)
	}()

	buffer := make(map[int][]uint8)

	fmt.Println("converting grids...")
	next := 0

	for result := range receipts {
		// fmt.Println(next)
		buffer[result.Index] = result.Img

		for {
			img, ok := buffer[next]
			if !ok {
				break
			}

			output <- img
			delete(buffer, next)
			next++
		}
	}
}

func (m *Model) makeVid(fps int, name string) {

	start := time.Now()

	first := grid2png(m.grids[0])

	bounds := first.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// options := vidio.Options{FPS: float64(fps), Loop: 0, Delay: 1000}
	options := vidio.Options{FPS: float64(fps)}
	gif, _ := vidio.NewVideoWriter(name+".mp4", w, h, &options)
	defer gif.Close()

	// imgs := make(chan []uint8)
	// go convertGrids(m, imgs)

	// prettyTime := 0.0

	fmt.Println("making video...")
	progressLength := 30.0
	for i := range m.grids {
		// i := 0
		// for img := range imgs {

		progress := float64(i) / float64(len(m.grids))
		clear_line()
		fmt.Printf("Progress: %.1f%% ", progress*100)

		fmt.Printf("[")
		lines := int(progressLength * progress)
		for range lines {
			fmt.Printf("-")
		}
		fmt.Printf(">")
		for range int(progressLength) - lines - 1 {
			fmt.Printf(" ")
		}
		fmt.Printf("]")

		// prettyStart := time.Now()
		img := grid2png(m.grids[i]).Pix
		// img := <-imgs
		// prettyTime += time.Since(prettyStart).Seconds()
		gif.Write(img)
		// i++
	}

	elapsedTime := time.Since(start)
	fmt.Printf("\nelapsed time: %.2fs\n", elapsedTime.Seconds())
	// fmt.Println(prettyTime)
}

func makeFFTUseful(fft []complex128) ([]float64, []float64) {
	amplitudes := make([]float64, len(fft))
	phases := make([]float64, len(fft))

	for i, value := range fft {
		amplitudes[i] = math.Sqrt(real(value)*real(value) + imag(value)*imag(value))
		phases[i] = math.Atan(imag(value) / real(value))
	}

	return amplitudes, phases
}

// func (m *Model)

func makeFFTChart(filename string, data Data) {

	n := len(data) / 2
	numPoints := n / 10
	bunnyData := make([]opts.LineData, 0, numPoints)
	foxData := make([]opts.LineData, 0, numPoints)

	time := make([]opts.LineData, 0, n/10)
	badfftBunny, badfftFox := data.Fft()

	fftBunny, phaseBunny := makeFFTUseful(badfftBunny)
	fftFox, phaseFox := makeFFTUseful(badfftFox)

	csv := ""
	maxBunny := 0.0
	maxBunnyDex := 0
	maxFox := 0.0
	maxFoxDex := 0

	for i := 2; i < numPoints; i++ {

		bunny := fftBunny[i]
		fox := fftFox[i]

		value := bunny
		bunnyData = append(bunnyData, opts.LineData{Value: value})

		value = fox
		foxData = append(foxData, opts.LineData{Value: value})

		time = append(time, opts.LineData{Value: float64(i) / float64(n)})

		csv += fmt.Sprintf(
			"%f, %f, %f, %f, %f\n",
			float64(i)/float64(n),
			bunny,
			fox,
			phaseBunny[i],
			phaseFox[i],
		)
		if fftBunny[i] > maxBunny {
			maxBunny = fftBunny[i]
			maxBunnyDex = i
		}
		if fftFox[i] > maxFox {
			maxFox = fftFox[i]
			maxFoxDex = i
		}
	}

	err := os.WriteFile(filename+"-fft.csv", []byte(csv), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println(maxBunnyDex)

	fmt.Printf(
		"max bunny phase: %f\n", phaseBunny[maxBunnyDex])
	fmt.Println("max fox phase: ", phaseFox[maxFoxDex])

	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(
			opts.Initialization{Theme: "dark", Width: "1400px", Height: "700px"},
		),
		charts.WithTitleOpts(opts.Title{
			Title: "Amplitude vs. Frequency",
		}))

	// chart.SetGlobalOptions(charts.WithColorsOpts({opts.RGBColor(255, 255, 255)}))

	line.SetXAxis(time).AddSeries("Bunnies", bunnyData).AddSeries("Foxes", foxData)
	line.SetSeriesOptions(charts.WithAnimationOpts(opts.Animation{AnimationDelay: 10}))
	// options := line.RenderSnippet().Option

	// line.Renderer = newSnippetRenderer(line, line.Validate)
	f, _ := os.Create(filename + ".html")
	// json, _ := json.Marshal(options)

	line.Render(f)
	// os.WriteFile(filename+".json", []byte(options), 0664)

}

func makeRaceChart(filename string, data Data) {
	bunnys := make([]opts.LineData, 0, len(data))
	foxes := make([]opts.LineData, 0, len(data))
	time := make([]int, 0, len(data))

	for _, point := range data {
		bunnys = append(bunnys, opts.LineData{Value: point.Bunnies})
		foxes = append(foxes, opts.LineData{Value: point.Foxes})
		time = append(time, point.Time)
	}

	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(
			opts.Initialization{Theme: "dark", Width: "1400px", Height: "700px"},
		),
		charts.WithTitleOpts(opts.Title{
			Title: "Population vs. Time",
		}))

	// chart.SetGlobalOptions(charts.WithColorsOpts({opts.RGBColor(255, 255, 255)}))

	line.SetXAxis(time).AddSeries("Bunnies", bunnys).AddSeries("Foxes", foxes)
	line.SetSeriesOptions(charts.WithAnimationOpts(opts.Animation{AnimationDelay: 10}))
	options := line.RenderSnippet().Option

	// line.Renderer = newSnippetRenderer(line, line.Validate)
	f, _ := os.Create(filename + ".html")
	line.Render(f)
	// json, _ := json.Marshal(options)
	os.WriteFile(filename+".json", []byte(options), 0664)
	// snippet := line.RenderSnippet()
	// var content []byte
	//
	// // option := []byte(snippet.Option)
	// content = append(content, []byte(snippet.Element)...)
	// content = append(content, []byte(snippet.Script)...)

	// and they say go isnt a scripting language
	// err := exec.Command("xdg-open", filename+".html").Run()
	// if err != nil {
	// 	fmt.Println("couldnt open chart:", err)
	// }
	// exec.Command("hyprctl", "dispatch", "workspace", "2").Run()
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
