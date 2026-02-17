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
		G: 0,
		B: 0,
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

func mid_point(size int) Point {
	// mid := (size - 1) / 2

	return Point{row: 0, col: 0}
}

func real_mid_point(size int) Point {

	mid := (size - 1) / 2

	return Point{row: mid, col: mid}
}

func add_points(p1, p2 Point) Point {
	return Point{row: p1.row + p2.row, col: p1.col + p2.col}
}

func (p *Point) add(p2 Point) {
	p.row += p2.row
	p.col += p2.col
}

func (p *Point) flip() {
	p.row *= -1
	p.col *= -1
}

func abs(x int) int {
	return max(x, x*-1)
}

func (p *Point) radius() int {
	return max(abs(p.row), abs(p.col))
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
	radius   int
	distance int
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

func heart_equation(t float64) (float64, float64) {
	x := 16 * math.Pow(math.Sin(t), 3)
	y := 13*math.Cos(t) - 5*math.Cos(2*t) - 2*math.Cos(3*t) - math.Cos(4*t)

	y /= -15
	x /= 15

	return x, y
}

func gen_heart_grid(size int, radius float64) Grid {

	if int(radius) > size/2 {
		panic(fmt.Errorf("ahhhhh: radius too large for grid size"))
	}

	grid := gen_grid(size)

	points := 1000

	for i := range points {
		t := float64(i) / float64(points) * 2 * math.Pi

		x, y := heart_equation(t)
		point := Point{col: int(math.Round(x * radius)), row: int(math.Round(y * radius))}

		real_point := add_points(point, mid_point(size))

		*grid.index(real_point) = Filled

	}

	return grid

	// heart := [][]int {
	// 	{0, 0, 1, 0, 0},
	// 	{0, 1, 0, 1, 0},
	// 	{1, 0, 0, 0, 1},
	// 	{0, 1, 0, 1, 0},
	// 	{0, 0, 1, 0, 0},
	// }

}

func (g Grid) raw_index(point Point) *SiteState {
	return &g[point.row][point.col]
}

func (g Grid) index(point Point) *SiteState {
	real_point := add_points(point, real_mid_point(len(g)))

	return &g[real_point.row][real_point.col]
}

func (g *Grid) is_valid_point(point Point) bool {
	radius := len(*g) / 2

	if point.row >= -radius && point.row <= radius && point.col >= -radius && point.col <= radius {
		return true
	} else {
		return false
	}

	// if point.row >= 0 && point.row < size && point.col >= 0 && point.col < size {
	// 	return true
	// } else {
	// 	return false
	// }
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

func init_model(size int, p float64, distance int) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	if distance <= 0 {
		panic("spawning distacne must be non-negative")
	}

	// grid := gen_grid(size)
	// *grid.index(mid_point(size)) = 1

	heart_radius := 30.0
	grid := gen_heart_grid(size, heart_radius)
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
		radius:   0,
		distance: 10,
	}

	return model

}

func (m *Model) random_start(r *rand.Rand) Point {
	value := int(r.Float64() * float64(m.size))
	value -= m.size / 2
	spawn_radius := m.radius + m.distance

	side := r.Int31n(4)

	var point Point

	switch side {
	case 0:
		point = Point{row: value, col: -spawn_radius}
	case 1:
		point = Point{row: value, col: spawn_radius}
	case 2:
		point = Point{row: -spawn_radius, col: value}
	case 3:
		point = Point{row: spawn_radius, col: value}
	default:
		panic("erm")
	}

	return point
}

func (m *Model) countNeibors(point Point) int {
	neighbors := 0
	for _, step := range CARDINALS {
		new_point := add_points(point, step)
		if m.grid.is_valid_point(new_point) && *m.grid.index(new_point) > 0 {
			neighbors++
		}
	}

	return neighbors
}

func (m *Model) onPerimeter(point Point) bool {
	radius := m.size / 2
	if point.row == -radius || point.row == radius || point.col == -radius || point.col == radius {
		return true
	} else {
		return false
	}
}

func (m *Model) tick(r *rand.Rand) bool {

	// index := -1
	walker := m.random_start(r)
	var new_point Point
	// fmt.Println("starting here:", walker)
	for true {

		// index++

		// walker := &m.walkers[index]

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
		for range neihbors {
			if r.Float64() <= m.p {
				*m.grid.index(new_point) = SiteState(m.infected)
				// *m.grid.index(new_point) = Filled
				m.infected++

				if new_point.radius() > m.radius {
					m.radius = new_point.radius()
				}

				if m.radius+m.distance >= m.size/2 {
					return true
				} else {
					return false
				}

				if m.onPerimeter(new_point) {
					return true
				} else {
					return false
				}

			}
		}

	}
	fmt.Println("shouldnt reach this")
	return true

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

func one_trial() {
	size := 201
	p := 1.0
	distance := 10
	model := init_model(size, p, distance)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// tps := 1

	// interval := 1
	// fmt.Println(interval)
	//
	// var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	// fmt.Println(delay)
	// fmt.Println(1.0 / float64(tps))

	for true {
		end := model.tick(r)

		if end {
			break
		}
	}
	pretty_picture(model, "image", 5)

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

	size := len(grid)

	img := image.NewNRGBA(image.Rect(0, 0, size*scale, size*scale))

	for y, row := range grid {
		for x := range row {

			var color color.Color
			if grid[x][y] == Empty {
				color = StateColor[Empty]
			} else {
				// color = StateColor[Filled]
				color = calc_color(float64(grid[x][y]) / float64(model.infected))
			}

			for i := range scale {
				for j := range scale {
					img.Set(y*scale+i, x*scale+j, color)
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
		P          float64 `json:"p"`
		Size       int     `json:"size"`
		Infected   int     `json:"infected"`
		Population int     `json:"population"`
	}{
		model.p,
		model.size,
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
