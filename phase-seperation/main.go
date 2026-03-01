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

type Grid [][]SiteState

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

type Model struct {
	grid  Grid
	size  int
	time  int
	ticks int
}

// func remove[T any](slice []T, s int) []T {
// 	return slices.Delete(slice, s, s+1)
// }

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

// func heart_equation_derive(x, y float64) float64 {
// 	return 3 * math.Pow(x * x + y * y - 1, 2)
//
// }

func round(x float64) int {
	return int(math.Round(x))
}

func (g Grid) raw_index(point Point) *SiteState {
	return &g[point.y][point.x]
}

func (g Grid) index(point Point) *SiteState {
	real_point := add_points(point, real_mid_point(len(g)))

	return &g[real_point.y][real_point.x]
}

func (g *Grid) is_valid_point(point Point) bool {

	radius := len(*g) / 2

	if point.x >= -radius && point.x <= radius && point.y >= -radius && point.y <= radius {
		return true
	} else {
		return false
	}
}

func (m *Model) index(point Point) *SiteState {
	return m.grid.index(point)
}

func clear_screen() {
	print("\u001b[2J")
}

func clear_line() {
	print("\u001b[2K")
	print("\r")
}

func random_step(r *rand.Rand) Point {

	value := r.Float64()

	if value < 0.25 {
		return Point{x: 1, y: 0}
	} else if value < 0.5 {
		return Point{x: -1, y: 0}
	} else if value < 0.75 {
		return Point{x: 0, y: 1}
	} else {
		return Point{x: 0, y: -1}
	}

}

func init_model(size int, ticks int, r *rand.Rand) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	// grid_type := "normal"
	// grid_type := "heart"
	heart := false

	var grid Grid
	if !heart {
		grid = gen_grid(size)

		cutoff := 0.5

		for i := range grid {
			for j := range grid {
				if r.Float64() < cutoff {
					grid[i][j] = Oil
				} else {
					grid[i][j] = Water
				}
			}
		}
		*grid.index(mid_point()) = 1
	} else {

		// heart_radius := 30.0
		// grid = gen_heart_grid(size, heart_radius)
	}
	// walkers := make([]Walker, 1)
	// walkers[0] = Walker{
	// 	location: middle,
	// 	ttl:      tau,
	// }

	model := Model{
		grid:  grid,
		size:  size,
		time:  0,
		ticks: ticks,
	}

	return model

}

func (m *Model) countNeibors(point Point) int {
	neighbors := 0
	politics := *m.grid.index(point)

	for _, step := range CARDINALS {
		new_point := add_points(point, step)
		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) == politics {
			neighbors++
		}
	}

	return neighbors
}

func (m *Model) countNonNeibors(point Point) int {
	return 4 - m.countNeibors(point)
}

func (m *Model) randomPoint(r *rand.Rand) Point {
	radius := m.size / 2

	return Point{x: r.Intn(m.size) - radius, y: r.Intn(m.size) - radius}
}

func (m *Model) shouldSwitch(point Point, r *rand.Rand) bool {
	nonNeihbors := m.countNeibors(point)

	slope := 0.25

	probability := float64(nonNeihbors) * slope

	return r.Float64() < probability
}

func (m *Model) tick(r *rand.Rand) {

	selected := m.randomPoint(r)
	state := *m.grid.index(selected)
	// start := time.Now()

	if m.shouldSwitch(selected, r) {
		var new_point Point
		for {
			step := random_step(r)
			new_point = add_points(selected, step)
			if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) == state {
				*m.index(selected) = *m.grid.index(new_point)
				*m.grid.index(new_point) = state

				break
			}
		}

		m.time++
	}

}

func (m *Model) run_trial(r *rand.Rand) Data {
	model := m

	for m.time < m.ticks {
		model.tick(r)
	}

	return make([]DataPoint, 0)

}

func run_simulation() stats.Series {
	size := 201
	num_points := 100.0

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	series := make(stats.Series, 0, int(num_points))

	for p := 0.01; p < 1.0; p += 0.01 {
		// p := p / num_points

		clear_line()
		fmt.Print("this much done: ", p*100, "%")

		model := init_model(size, 1000, r)

		data := model.run_trial(r)
		casted := data.toSeries()
		logged := logLog(casted)

		_, gradient, err := LinearRegression(logged)

		if err != nil {
			panic(err)
		}

		series = append(series, stats.Coordinate{X: p, Y: gradient})

	}

	// pretty_picture(model, "testing", 5)
	return series

}

func testing() {
	// // data := []int{0, 1, 2, 3}
	// point := Point{row: 1, col: 1}
	// data := []Point{point, point}
	//
	// // for _, idk := range data {
	// // 	idk = 67
	// // }
	//
	// // data = unordered_remove(data, 1)
	//
	// data[0].row = 2
	//
	// fmt.Println(data)
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

func logLog(series stats.Series) stats.Series {

	logged := make(stats.Series, 0, len(series))

	for _, point := range series {
		logged = append(logged, stats.Coordinate{X: math.Log(point.X), Y: math.Log(point.Y)})
	}

	return logged
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
	// parameters:
	size := 201
	ticks := int(10e3)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	model := init_model(size, ticks, r)
	// tps := 1

	// interval := 1
	// fmt.Println(interval)
	//
	// var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	// fmt.Println(delay)
	// fmt.Println(1.0 / float64(tps))

	data := model.run_trial(r)
	// for _, point := range data {
	// 	fmt.Println(point.radius)
	// }
	for _, point := range data {
		fmt.Println(point.filled)
	}

	pretty_picture(model, filename, 5)

	// pretty_picture(model, "immune", 5)

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

func pretty_picture(model Model, name string, scale int) {

	grid := model.grid

	// cropped := model.size - model.distance*2

	img := image.NewNRGBA(image.Rect(0, 0, model.size, model.size))

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
