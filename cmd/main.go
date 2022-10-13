package main

import (
	"math"
	"math/rand"
	"quadtree-wasm/pkg/canvas"
	"quadtree-wasm/pkg/quadtree"
	"strconv"
	"syscall/js"
)

type App struct {
	cvs       *canvas.Canvas
	qt        *quadtree.QtNode
	demoQuery bool
}

func main() {
	c := make(chan struct{})

	app := App{}

	js.Global().Set("initQuadtreeDemo", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		canvasEl := args[0]
		width := args[1].Int()
		height := args[2].Int()
		demoQuery := args[3].Bool()
		initialPts := args[4].Int()

		app.init(canvasEl, width, height, demoQuery, initialPts)
		return js.Undefined()
	}))

	js.Global().Set("clearPts", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.buildScene(0)
		return js.Undefined()
	}))

	<-c
}

func (app *App) init(canvasEl js.Value, width, height int, demoQuery bool, initialPts int) {
	app.demoQuery = demoQuery // enable query animation

	{ // setup canvas
		app.cvs = canvas.New(&canvasEl)
		app.cvs.Call("setAttribute", "width", strconv.Itoa(width))
		app.cvs.Call("setAttribute", "height", strconv.Itoa(height))
		app.cvs.Set("width", width)
		app.cvs.Set("height", height)
		app.cvs.ScaleCanvasByDPI() // scale canvas by dpi for retina displays
	}

	{ // create update loop
		var l func(this js.Value, args []js.Value) interface{}

		l = func(this js.Value, args []js.Value) interface{} {
			timestampJS := args[0]
			app.loop(timestampJS.Float())
			js.Global().Call("requestAnimationFrame", js.FuncOf(l))
			return 0
		}
		js.Global().Call("requestAnimationFrame", js.FuncOf(l))
	}

	//handle click event
	{
		app.cvs.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) any {
			app.handleClick(args[0])
			return nil
		}))
	}

	app.buildScene(initialPts)
}

func (app *App) buildScene(initialPts int) {
	bcr := app.cvs.GetBoundingClientRect()
	app.qt = quadtree.New(0, 0, bcr.Width, bcr.Height, 2)

	// add points

	// seed := time.Now().UnixMilli()
	rand.Seed(1662992361055)

	for i := 0; i < initialPts; i++ {
		pt := quadtree.Point{X: rand.Float64() * bcr.Width, Y: rand.Float64() * bcr.Height}
		app.qt.Insert(pt)
	}
}

func (app *App) loop(ts float64) {
	app.cvs.Clear()

	// draw pts
	{
		for _, pt := range app.qt.Pts {
			app.cvs.DrawPt(pt.X, pt.Y, 4, "#484848")
		}
	}

	app.cvs.DrawQuadTree(app.qt)

	if app.demoQuery {
		bcr := app.cvs.GetBoundingClientRect()

		w, h := 100.0, 100.0

		durationX := 24000
		durationY := 12000
		tX := math.Mod(ts, float64(durationX)) / float64(durationX)
		tY := math.Mod(ts, float64(durationX)) / float64(durationY)
		cx := bcr.Width * .5
		cy := bcr.Height * .5
		rX := 200.0
		rY := 150.0

		x := cx + math.Cos(math.Pi*2*tX)*rX
		y := cy + math.Sin(math.Pi*2*tY)*rY

		qry := quadtree.Rect{L: x - w*.5, T: y + h*.5, R: x + w*.5, B: y - h*.5}

		qryRes := app.qt.BoxQuery(qry)

		// draw query
		{
			app.cvs.DrawRect(x-w*.5, y-h*.5, w, h, "red", true)

			for _, pt := range qryRes {
				// if pt is inside query rect, draw it in red
				app.cvs.DrawPt(pt.X, pt.Y, 4, "red")
			}
		}
	}
}

func (app *App) handleClick(evt js.Value) interface{} {
	bcr := app.cvs.GetBoundingClientRect()
	x := evt.Get("clientX").Float() - bcr.Left
	y := (evt.Get("clientY").Float() - bcr.Top - bcr.Height) * -1 // flip y
	app.qt.Insert(quadtree.Point{X: x, Y: y})
	return nil
}

type Canvas struct {
	*js.Value
}

type Context2D struct {
	*js.Value
}

func New(el *js.Value) *Canvas {
	return &Canvas{el}
}
