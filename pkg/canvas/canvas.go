package canvas

import (
	"fmt"
	"math"
	"quadtree-wasm/pkg/quadtree"
	"syscall/js"
)

type Canvas struct {
	*js.Value
}

type Context2D struct {
	*js.Value
}

func New(el *js.Value) *Canvas {
	return &Canvas{el}
}

func (c *Canvas) DrawQuadTree(qt *quadtree.QtNode) {
	c.DrawRect(qt.X, qt.Y, qt.W, qt.H, "darkgray", true)
	if qt.Bl != nil {
		c.DrawQuadTree(qt.Bl)
	}
	if qt.Tl != nil {
		c.DrawQuadTree(qt.Tl)
	}
	if qt.Tr != nil {
		c.DrawQuadTree(qt.Tr)
	}
	if qt.Br != nil {
		c.DrawQuadTree(qt.Br)
	}
}

func (c *Canvas) DrawPt(x, y, rad float64, color string) {
	ctx := c.GetContext2D()
	ctx.Call("beginPath")
	ctx.Call("arc", x, y, rad, 0, math.Pi*2)
	ctx.Set("fillStyle", color)
	ctx.Set("strokeStyle", color)
	ctx.Call("fill")
	ctx.Call("stroke")
}

func (c *Canvas) DrawRect(x, y, w, h float64, color string, stroke bool) {
	ctx := c.GetContext2D()
	ctx.Call("beginPath")
	ctx.Call("rect", x, y, w, h)
	ctx.Set("strokeStyle", color)
	if stroke {
		ctx.Call("stroke")
	}
}

func (c *Canvas) DrawLine(x1, y1, x2, y2 float64, color string) {
	ctx := c.GetContext2D()
	ctx.Call("beginPath")
	ctx.Call("moveTo", x1, y1)
	ctx.Call("lineTo", x2, y2)
	ctx.Set("strokeStyle", color)
	ctx.Call("stroke")
}

func (c *Canvas) GetContext2D() *Context2D {
	ctx := c.Call("getContext", "2d")
	return &Context2D{&ctx}
}

func (c *Canvas) ScaleCanvasByDPI() {
	w := c.Get("clientWidth").Float()
	h := c.Get("clientHeight").Float()
	dpi := js.Global().Get("devicePixelRatio")

	c.Set("width", w*dpi.Float())
	c.Set("height", h*dpi.Float())

	c.Get("style").Set("width", fmt.Sprintf("%fpx", w))
	c.Get("style").Set("height", fmt.Sprintf("%fpx", h))

	ctx := c.GetContext2D()
	ctx.Call("scale", dpi.Float(), dpi.Float())

	c.Get("style").Set("transform", "scaleY(-1)")
}

func (c *Canvas) Clear() {
	ctx := c.GetContext2D()
	bcr := c.GetBoundingClientRect()
	ctx.Call("clearRect", 0, 0, bcr.Width, bcr.Height)
}

func (c *Canvas) GetBoundingClientRect() struct{ X, Y, Width, Height, Top, Right, Bottom, Left float64 } {
	bcrJS := c.Call("getBoundingClientRect")

	return struct{ X, Y, Width, Height, Top, Right, Bottom, Left float64 }{
		bcrJS.Get("x").Float(),
		bcrJS.Get("y").Float(),
		bcrJS.Get("width").Float(),
		bcrJS.Get("height").Float(),
		bcrJS.Get("top").Float(),
		bcrJS.Get("right").Float(),
		bcrJS.Get("bottom").Float(),
		bcrJS.Get("left").Float(),
	}
}
