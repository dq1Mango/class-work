package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"time"

	textColor "github.com/fatih/color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/montanaflynn/stats"
	// "slices"
)

// some clever stack overflow code i found
type SiteState int

const (
	Susceptible SiteState = iota
	Removed
	Visited
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
	Visited: {
		R: 0,
		G: 0,
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

type Walkers struct {
	queue []*[]Walker
	alive int
}

func make_walkers(tau int) Walkers {
	walkers := make([]*[]Walker, tau+1)

	for i := range walkers {
		tmp := make([]Walker, 0)
		walkers[i] = &tmp
	}

	return Walkers{queue: walkers, alive: 0}
}

func (w *Walkers) add_walker(walker Walker) {
	q := w.queue
	last := len(q) - 1
	*q[last] = append(*q[last], walker)

	w.alive += 1

}

func (w *Walkers) tick() {

	q := w.queue
	w.alive -= len(*q[0])

	for i := range len(q) - 1 {
		q[i] = q[i+1]
	}

	tmp := make([]Walker, 0)

	q[len(q)-1] = &tmp
}

func (w *Walkers) Iterate() []*Walker {
	q := w.queue

	size := 0
	for i := range len(q) - 1 {
		size += len(*q[i])
	}

	flattened := make([]*Walker, 0, size)

	for i := range len(q) - 1 {
		for j := range *q[i] {
			flattened = append(flattened, &(*q[i])[j])
		}
	}

	// fmt.Println("", flattened)
	return flattened

}

type Model struct {
	grid     Grid
	walkers  Walkers
	p        float64
	tau      int
	people   int
	infected int
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

	// walkers := make([]Walker, 1)
	// walkers[0] = Walker{
	// 	location: middle,
	// 	ttl:      tau,
	// }

	walkers := make_walkers(tau)
	for range 2 {
		walkers.add_walker(Walker{
			location: middle,
			ttl:      tau,
		})
	}
	walkers.tick()

	model := Model{
		grid:     grid,
		walkers:  walkers,
		p:        p,
		people:   size * size,
		infected: 1,
	}

	return model
}

func (m *Model) add_walker(walker Walker) {
	m.walkers.add_walker(walker)
	// m.walkers = append(m.walkers, walker)
}

func (m *Model) tick(r *rand.Rand) {

	index := -1
	// for index < len(m.walkers)-1 {
	for _, walker := range m.walkers.Iterate() {
		index++

		// walker := &m.walkers[index]

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

		// start = time.Now()
		// r := rand.New(rand.NewSource(time.Now().UnixNano()))

		state := m.grid.index(new_point)
		if *state == Susceptible || *state == Visited {
			value := r.Float64()
			if value < m.p {
				*state = Removed
				// walkers_to_add = append(walkers_to_add, Walker{location: new_point, ttl: m.tau})
				m.walkers.add_walker(Walker{location: new_point, ttl: m.tau})
				m.infected++
			} else {
				*state = Visited
			}
		}

		// stop = time.Now()

		// fmt.Println("maybe spreading took this long: ", stop.Sub(start))

	}

	m.walkers.tick()

}

func testing() {
	// data := []int{0, 1, 2, 3}
	point := Point{row: 1, col: 1}
	data := []Point{point, point}

	// for _, idk := range data {
	// 	idk = 67
	// }

	// data = unordered_remove(data, 1)

	data[0].row = 2

	fmt.Println(data)
}

func average_spread_rate(size int, p float64, tau int, r *rand.Rand) float64 {
	model := init_model(size, p, tau)
	infected_stop := int(0.2 * float64(model.people))
	i := 0

	for model.infected < infected_stop {
		// infected_ratio := float64(model.infected) / float64(model.people)

		if model.walkers.alive == 0 {
			// fmt.Println("Failed to spread")
			return 0
			// break
		}

		// start := time.Now()
		model.tick(r)
		// stop := time.Now()
		// fmt.Println("took this long: ", stop.Sub(start))
		i++
	}

	return float64(model.infected) / float64(i)

}

type Data []opts.Chart3DData

func run_simulation() Data {
	size := 201
	// p := 0.5
	// tau := 5

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	tau_values := 20
	trials := 11
	num_points := tau_values * tau_values
	data := make(Data, 0, num_points)
	// average := 0.0

	for P := range tau_values {
		p := float64(P) / float64(tau_values)
		for tau := range tau_values {
			fmt.Println("This much done: ", float64(P*tau_values+tau)/float64(num_points), "%")

			values := make([]float64, 0, trials)
			for range trials {
				values = append(values, average_spread_rate(size, p, tau, r))
			}

			median, err := stats.Median(values)
			if err != nil {
				panic(fmt.Sprint("ahhhh", err))
			}

			data = append(data, opts.Chart3DData{
				Value: []interface{}{p, float64(tau), median}})
			// Value: []interface{}{1, 2, 3},
		}
	}

	return data

}

type Arguments struct {
	file    string
	is_file bool
}

func parse_args() Arguments {
	args := os.Args[1:]
	if len(args) >= 2 {
		if args[0] == "--file" || args[0] == "-f" {
			return Arguments{
				file:    args[1],
				is_file: true,
			}
		}
	}

	return Arguments{
		file:    "",
		is_file: false,
	}
}

func main() {
	// testing()
	// return

	var data Data

	args := parse_args()

	if args.is_file {
		json_data, err := os.ReadFile(args.file)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(json_data, &data)
	} else {
		data = run_simulation()

	}

	make_chart(data)

	file, err := os.Create("data.json")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)

}

func make_chart(data []opts.Chart3DData) {
	surface := charts.NewSurface3D()

	surface.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "3D Surface Example",
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Max:       1,
			Min:       -1,
			Dimension: "x",
		}),
	)

	surface.AddSeries("surface", data)
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

func one_trial(model Model, r *rand.Rand) {
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
		if infected_ratio > 0.2 {
			pretty_picture(model, "image.png", 5)
			break
		}

		if model.walkers.alive == 0 {
			fmt.Println("Failed to spread")
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

	for _, walker := range model.walkers.Iterate() {
		for i := range scale {
			for j := range scale {
				img.Set(walker.location.col*scale+i, walker.location.row*scale+j, color.NRGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 255,
				})
			}
		}
	}

	f, err := os.Create(name)
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
