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

	textColor "github.com/fatih/color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	// "github.com/montanaflynn/stats"
	// "slices"
)

const END_RATIO = 0.1

type SiteState int

const (
	Empty SiteState = iota
	Filled
	// Visited
	// Immune
)

var StateColor = map[SiteState]color.NRGBA{
	Empty: {
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	},
	Filled: {
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	},
	// Visited: {
	// 	R: 0,
	// 	G: 0,
	// 	B: 255,
	// 	A: 255,
	// },
	// Immune: {
	// 	R: 0,
	// 	G: 255,
	// 	B: 0,
	// 	A: 255,
	// },
}

type Grid [][]SiteState

type Point struct {
	row int
	col int
}

var CARDINALS = []Point{
	{row: 0, col: -1},
	{row: 0, col: 1},
	{row: 1, col: 0},
	{row: -1, col: 0},
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
	size     int
	p        float64
	people   int
	infected int
	time     int
}

// func remove[T any](slice []T, s int) []T {
// 	return slices.Delete(slice, s, s+1)
// }

func unordered_remove[T any](s []T, i int) []T {
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
			case 2:
				textColor.Set(textColor.FgBlue)

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

func clear_line() {
	print("\u001b[2K")
	print("\r")
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

func init_model(size int, p float64) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	mid := (size - 1) / 2

	middle := Point{row: mid, col: mid}

	grid := gen_grid(size)

	*grid.index(middle) = 1

	// walkers := make([]Walker, 1)
	// walkers[0] = Walker{
	// 	location: middle,
	// 	ttl:      tau,
	// }

	model := Model{
		grid:     grid,
		size:     size,
		p:        p,
		people:   size * size,
		infected: 1,
	}

	return model
}

func (m *Model) random_start(r *rand.Rand) Point {
	value := int(r.Float64() * float64(m.size))

	side := r.Int31n(4)

	switch side {
	case 0:
		return Point{row: value, col: 0}
	case 1:
		return Point{row: value, col: m.size - 1}
	case 2:
		return Point{row: 0, col: value}
	case 3:
		return Point{row: m.size - 1, col: value}
	default:
		panic("erm")
	}
}

func (m *Model) countNeibors(point Point) int {
	neighbors := 0
	for _, step := range CARDINALS {
		new_point := add_points(point, step)
		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) == Filled {
			neighbors++
		}
	}

	return neighbors
}

func (m *Model) tick(r *rand.Rand) {

	// index := -1
	walker := m.random_start(r)
	// fmt.Println("starting here:", walker)
	for true {

		// index++

		// walker := &m.walkers[index]

		var new_point Point

		// start := time.Now()

		for {
			step := random_step(r)
			new_point = add_points(walker, step)
			if m.grid.is_valid_point(new_point) {
				walker = new_point
				// m.walkers[index].location = new_point
				break
			}
		}

		neihbors := m.countNeibors(new_point)
		if neihbors > 0 {
			*m.grid.index(new_point) = Filled
			m.infected++
			break
		}

	}

	m.time++

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

type Data []opts.Chart3DData

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

func main() {
	// testing()
	// return

	one_trial()
	return

	var data Data

	args := parse_args()
	//
	// if args.output == nil {
	// 	panic("no output file name specified")
	// }
	//
	// if file := args.file; file != nil {
	// 	json_data, err := os.ReadFile(*file)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	json.Unmarshal(json_data, &data)
	//
	// } else {
	// 	data = run_simulation()
	//
	// 	file, err := os.Create(*args.output + "data.png")
	//
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	defer file.Close()
	//
	// 	encoder := json.NewEncoder(file)
	// 	encoder.SetIndent("", "  ")
	// 	encoder.Encode(data)
	//
	// }

	switch *args.chart {
	case "3d":
		make_3d_chart(data)

	case "threshold":
		make_threshold_chart(data)
	case "derivative":
		make_derivative_chart(data)
	case "cross":
		cross_sections(data)

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

func cast_to_float(input []any) ([]float64, error) {
	output := make([]float64, len(input))

	for i, elem := range input {
		switch e := elem.(type) {
		case float64:
			output[i] = e

		default:
			return []float64{}, fmt.Errorf("not of type float64")
		}
	}

	return output, nil
}

// func shittyCode101(input []any) (int, int, float64, error) {
// 	floats, err := cast_to_float(input)
// 	if err != nil {
// 		return 0, 0, 0, err
// 	}
//
// 	return
//
//
// }

func formatXY(Data []opts.Chart3DData) [][]float64 {
	data := make([][]float64, 50)

	for index := range data {
		data[index] = make([]float64, 50)
	}

	for _, value := range Data {

		values, err := cast_to_float(value.Value)
		if err != nil {
			panic(err)
		}

		// fmt.Println(values)

		data[int(values[1])][int(math.Round(values[0]*50))] = values[2]

		// switch v := value.Value[0].(type) {
		// case float64:
		// 	data[row_dex][col_dex] = v
		// default:
		// 	panic("howd that get in here")
		// }
	}

	return data

}

func make_threshold_chart(Data []opts.Chart3DData) {

	data := formatXY(Data)

	threshold := make([]float64, 50)

	for p, row := range data {
		j := 0
		for j < len(row) && row[j] == 0 {
			j++
		}

		j--

		threshold[p] = float64(j) / 50

	}

	// fmt.Println(threshold)

	// for i := len(threshold) - 1; i >= 0; i-- {
	// 	fmt.Println(threshold[i])
	// }

	for _, value := range threshold {
		fmt.Println(value)
	}

}

const DATA_SIZE = 50

// MUST BE tau by p
func cast_to_echart_format(data [][]float64) []opts.Chart3DData {
	// fmt.Println(data)

	output := make([]opts.Chart3DData, 0, len(data)*len(data[0]))

	for i, row := range data {
		for j, k := range row {
			if k == 0 {
				continue
			}

			output = append(output, opts.Chart3DData{Value: []any{j, float64(i) / DATA_SIZE, k}})
		}
	}

	return output
}

func make_derivative_chart(Data []opts.Chart3DData) {
	data := formatXY(Data)

	d_tau := make([][]float64, 50)
	d_p := make([][]float64, 50)

	for i := range d_tau {
		d_tau[i] = make([]float64, 50-1)
		d_p[i] = make([]float64, 50-1)

	}

	for i := range DATA_SIZE - 1 {
		for j := range DATA_SIZE - 2 {
			if data[i][j] == 0 {
				continue
			}

			d_p[i][j] = (data[i][j+1] - data[i][j])
		}

	}

	for j := range DATA_SIZE - 1 {
		for i := range DATA_SIZE - 2 {
			d_tau[i][j] = (data[i+1][j] - data[i][j])

		}
	}

	// make_3d_chart(cast_to_echart_format(d_tau))
	make_3d_chart(cast_to_echart_format(d_p))

}

func cross_sections(Data []opts.Chart3DData) {
	data := formatXY(Data)

	tau_sections := []int{5, 10, 20}

	for _, tau := range tau_sections {
		fmt.Println("tau value:", tau)

		for i := range data[tau] {
			fmt.Println(data[tau][i])
		}
	}

	// p_sections := []int{10, 30, 49}
	//
	// for _, p := range p_sections {
	// 	fmt.Println("p value:", p)
	//
	// 	for i := range data[p] {
	// 		fmt.Println(data[i][p])
	// 		// fmt.Println(val)
	// 	}
	// }

}

func one_trial() {
	size := 201
	p := 0.2
	model := init_model(size, p)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// tps := 1

	// interval := 1
	// fmt.Println(interval)
	//
	// var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	// fmt.Println(delay)
	// fmt.Println(1.0 / float64(tps))

	i := 0
	for model.infected < model.people {
		infected_ratio := float64(model.infected) / float64(model.people)
		// fmt.Println("iter: ", i, ", infected: ", infected_ratio)
		if infected_ratio > END_RATIO {
			pretty_picture(model, "image", 5)
			break
		}

		// start := time.Now()
		model.tick(r)
		// stop := time.Now()
		// fmt.Println("took this long: ", stop.Sub(start))

		// model.grid.print_grid()
		// time.Sleep(delay)
		// clear_screen()
		i++

		// if i%interval == 0 {
		// 	make_graph(model, fmt.Sprintf("image-%d.png", i/interval), 30)
		// }

		// if i > 1000 {
		// 	break
		// }

	}

	// pretty_picture(model, "immune", 5)

}

func pretty_picture(model Model, name string, scale int) {

	grid := model.grid

	size := len(grid)

	img := image.NewNRGBA(image.Rect(0, 0, size*scale, size*scale))

	for y, row := range grid {
		for x := range row {
			for i := range scale {
				for j := range scale {

					img.Set(y*scale+i, x*scale+j, StateColor[grid[x][y]])
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

	file, err := os.Create(name + "-data.png")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	// encoder.Encode(model)
	encoder.Encode(struct {
		Time       int `json:"time"`
		Infected   int `json:"infected"`
		Population int `json:"population"`
	}{
		model.time,
		model.infected,
		model.people,
	})

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
