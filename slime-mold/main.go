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
	"slices"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/montanaflynn/stats"
	// "slices"
)

const END_RATIO = 0.1
const Size = 51

const SACRIFICE = 0.01
const SELFISH = 0.9

const (
	FPS = 60.0
)

var WEIGHT_VECTOR = Vector{x: -1, y: 1}

type SiteState int

const (
	Empty SiteState = iota
	Filled
	Origin SiteState = -2
	Active SiteState = -1
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
	Origin: {
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	},
	Active: {
		R: 0,
		G: 255,
		B: 0,
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

type Vector struct {
	x float64
	y float64
}

var UNITS = []Vector{
	{x: 0, y: -1},
	{x: 0, y: 1},
	{x: 1, y: 0},
	{x: -1, y: 0},
}

func (v *Vector) scale(r float64) {
	v.x *= r
	v.y *= r
}

func (v *Vector) magnitude() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v *Vector) normalize() {
	magnitude := v.magnitude()

	v.scale(1 / magnitude)
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

	// for _, p := range permutations {
	// 	fmt.Println(p)
	// }

	// fmt.Println(permutations)

	return permutations

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

type Walker struct {
	location Point
	ttl      int
}

type Model struct {
	grid  Grid
	grids []Grid
	size  int
	// p        float64
	// people   int
	// infected int
	radius int
	time   int
	// distance int
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
		point := Point{x: int(math.Round(x * radius)), y: int(math.Round(y * radius))}

		*grid.index(point) = Filled

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
	return &g[point.y][point.x]
}

func (g Grid) index(point Point) *SiteState {
	point.y *= -1
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

func init_model(size int, p float64, distance int) Model {

	if size%2 == 0 {
		panic("grid size must be odd you doofus")
	}

	if distance <= 0 {
		panic("spawning distacne must be non-negative")
	}

	// grid_type := "normal"
	// grid_type := "heart"
	heart := false

	var grid Grid
	if !heart {
		grid = gen_grid(size)
		*grid.index(mid_point()) = Filled
	} else {

		heart_radius := 30.0
		grid = gen_heart_grid(size, heart_radius)
	}

	// walkers := make([]Walker, 1)
	// walkers[0] = Walker{
	// 	location: middle,
	// 	ttl:      tau,
	// }

	model := Model{
		grid:  grid,
		grids: make([]Grid, 0, 100),
		size:  size,
		time:  0,
		// p:        p,
		// people:   size * size,
		// infected: 1,
		// distance: distance,
	}

	return model

}

func (m *Model) origin() Point {
	return Point{x: 0, y: 0}
}

// func (m *Model) random_start(r *rand.Rand) Point {
// 	value := int(r.Float64() * float64(m.size))
// 	value -= m.size / 2
// 	spawn_radius := m.radius + m.distance
//
// 	side := r.Int31n(4)
//
// 	var point Point
//
// 	switch side {
// 	case 0:
// 		point = Point{x: value, y: -spawn_radius}
// 	case 1:
// 		point = Point{x: value, y: spawn_radius}
// 	case 2:
// 		point = Point{x: -spawn_radius, y: value}
// 	case 3:
// 		point = Point{x: spawn_radius, y: value}
// 	default:
// 		panic("erm")
// 	}
//
// 	return point
// }

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

	radius := (m.size - 1) / 2
	if point.x == radius || point.x == -radius || point.y == radius || point.y == -radius {
		return true
	} else {
		return false
	}

}

func (m *Model) countOnRadius(radius int) int {
	count := 0

	if radius == 0 {
		return 1
	}

	for i := range radius*2 + 1 {
		i -= radius

		if *m.grid.index(Point{x: i, y: -radius}) > 0 {
			count++
		}
		if *m.grid.index(Point{x: i, y: radius}) > 0 {
			count++
		}
	}
	// dont wanna double count the corners
	for i := range radius*2 - 1 {
		i -= radius

		if *m.grid.index(Point{x: -radius, y: i}) > 0 {
			count++
		}
		if *m.grid.index(Point{x: radius, y: i}) > 0 {
			count++
		}
	}

	return count
}

func dotProduct(v1, v2 Vector) float64 {
	return v1.x*v2.x + v1.y*v2.y
}

func selectProbability(probs []float64, r *rand.Rand) int {

	sumProbs := 0.0

	for _, value := range probs {
		sumProbs += value
	}

	selection := r.Float64() * sumProbs

	runningProb := sumProbs

	i := len(probs) - 1

	for i > 0 {

		runningProb -= probs[i]
		if selection > runningProb {
			// fmt.Println(selection, prob)
			return i
		}
		i--
	}
	return 0
}

func weightedDirection(weight Vector) []float64 {

	probabilites := make([]float64, 4)

	for i, unit := range UNITS {
		dot := dotProduct(weight, unit)

		probabilites[i] = math.Exp(dot)
	}

	return probabilites
	// panic("ahhhhh")

}

func (m *Model) tick(r *rand.Rand) bool {

	walker := m.origin()

	var new_point Point

	for true {

		// start := time.Now()

		// step := random_step(r)

		// index := r.Int63n(4)
		// if index == 3 {
		// 	continue
		// }
		// direction := CARDINALS[index]
		probs := weightedDirection(WEIGHT_VECTOR)

		for i := range probs {
			if *m.grid.index(new_point) == Empty {
				probs[i] *= SELFISH
			} else {
				probs[i] *= SACRIFICE
			}
		}

		selection := CARDINALS[selectProbability(probs, r)]

		new_point = add_points(walker, selection)

		walker = new_point

		if *m.grid.index(walker) == Empty {
			// fmt.Println("hi")
			*m.grid.index(walker) = Filled

			if m.onPerimeter(new_point) {
				return true
			} else {
				return false
			}
		}
	}

	panic("shouldnt reach this")
}

// func (m *Model) alternate_tick(r *rand.Rand) bool {
//
// 	walker := m.origin()
//
// 	var new_point Point
//
// 	for true {
//
// 		// start := time.Now()
//
// 		// step := random_step(r)
//
// 		for {
// 			// index := r.Int63n(4)
// 			// if index == 3 {
// 			// 	continue
// 			// }
// 			// direction := CARDINALS[index]
// 			direction := weightedDirection(WEIGHT_VECTOR)
//
// 			new_point = add_points(walker, mid_point())
//
// 			if *m.grid.index(new_point) == Filled {
// 				if r.Float64() < SELFISH {
// 					walker = new_point
// 					break
// 				}
// 			} else {
// 				if r.Float64() < SACRIFICE {
//
// 					*m.grid.index(new_point) = Filled
//
// 					if m.onPerimeter(new_point) {
// 						return true
// 					} else {
// 						return false
// 					}
// 				}
// 			}
//
// 		}
// 	}
//
// 	panic("shouldnt reach this")
// }

func (m *Model) run_trial(r *rand.Rand) Data {
	model := m

	for m.time < int(2e3) {
		// fmt.Println(m.time)
		end := model.tick(r)

		copied := make(Grid, m.size)
		for i := range copied {
			copied[i] = slices.Clone(m.grid[i])
		}
		// *copied.index(new_point) = Active

		m.grids = append(m.grids, copied)

		m.time++
		// end := model.different_tick(r)

		// fmt.Println("ticked me off")

		if end {
			break
		}
	}

	data := make(Data, 0, model.radius)
	// running_total := 1
	//
	// for r := 1; r < model.radius; r++ {
	// 	running_total += model.countOnRadius(r)
	// 	data = append(data, DataPoint{radius: r, filled: running_total})
	// }
	return data

}

func run_simulation() stats.Series {
	distance := 20
	num_points := 100.0

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	series := make(stats.Series, 0, int(num_points))

	for p := 0.01; p < 1.0; p += 0.01 {
		// p := p / num_points

		clear_line()
		fmt.Print("this much done: ", p*100, "%")

		model := init_model(Size, p, distance)

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

	// for
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
		series = append(
			series,
			stats.Coordinate{X: float64(point.radius), Y: float64(point.filled)},
		)
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

	WEIGHT_VECTOR.normalize()
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
	size := 101
	p := 0.1
	distance := 20
	model := init_model(size, p, distance)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// tps := 1

	// interval := 1
	// fmt.Println(interval)
	//
	// var delay time.Duration = time.Duration(1.0 / float64(tps) * math.Pow10(9))
	// fmt.Println(delay)
	// fmt.Println(1.0 / float64(tps))

	_ = model.run_trial(r)

	model.makeVid(filename)
	// for _, point := range data {
	// 	fmt.Println(point.radius)
	// }
	// for _, point := range data {
	// 	fmt.Println(point.filled)
	// }
	// casted := data.toSeries()
	// logged := logLog(casted)
	//
	// _, gradient, err := LinearRegression(logged)
	// fmt.Println(gradient)
	//
	// if err != nil {
	// 	panic(err)
	// }

	pretty_picture(model, filename, 10)

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

	*grid.index(model.origin()) = Origin

	// cropped := model.size - model.distance*2
	cropped := model.size

	img := image.NewNRGBA(image.Rect(0, 0, cropped*scale, cropped*scale))

	// for y, row := range grid[model.distance : model.size-model.distance] {
	// 	for x, value := range row[model.distance : model.size-model.distance] {
	for y, row := range grid {
		for x, value := range row {

			var color = StateColor[value]

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
}

func grid2png(grid Grid) *image.NRGBA {

	size := len(grid)
	screen_size := 960

	scale := screen_size / size

	*grid.index(mid_point()) = Origin

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
}

func (m *Model) makeVid(name string) {

	first := grid2png(m.grids[0])

	bounds := first.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	// options := vidio.Options{FPS: float64(fps), Loop: 0, Delay: 1000}
	options := vidio.Options{FPS: float64(FPS)}
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
}
