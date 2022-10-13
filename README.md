# Quad Trees

## Introduction

Quadtrees are one of those easy to learn concepts which present opportunities to think about a host of problems in a new light.  I found use for quadtrees when trying to optimize my [Boids](https://ronandoherty.com/projects/boids) project a little while ago.  Unoptimized, Boids is a O(n2) algorithm where for each step of the simulation each Boid must measure it's distance to every other Boid and apply forces if close enough.  Quadtrees allowed me to greatly reduce the number of proximity checks each frame, increasing the number of Boids I could add to the simulation.  

Feel free to go check out my [Boids](https://ronandoherty.com/projects/boids) project now and set the `Boid Search Algorithm` property to Quadtree and click the `Draw` checkbox to see a Quadtree in action.

In this post I will describe a **Point Quadtree** which is adept at storing and querying 2D point data.  I will also describe how to perform a box query on the Quadtree to find all points within a given rectangle.

You can find a repo with the demo code here: https://github.com/donanroherty/quadtree-go.git

## Point QuadTrees

A point quadtree is an undirected tree data structure, ie. hierarchy of connected nodes.  We call it a point quadtree because this specific implementation will be used to store 2D (x, y) point data.  Each node in the tree holds a list of points and has a point capacity.  If the capacity of a node is exceeded the node will be subdivided into exactly 4 child nodes (bottom left, top left, top right, bottom right) and the points held by the parent node will be distributed amongst it's children.  This process continues recursively until all points are allocated to a node without exceeding capacity.

The demo below should help illustrate this.  Click inside the box to add points and watch as the space is decomposed into smaller quads.  If the demo is not working for you it may be that your browser security settings are blocking web assembly.

<div align="center">
  <iframe src="https://quadtree-demo-opal.vercel.app?width=400&height=400&pts=0" width="410" height="410" frameBorder="0"></iframe>
</div>

### Implementing a quad tree

The code below is written in Go, however should be easy enough to understand people without Go experience.  Go is kinda cool that way.

```go
// pkg/quadtree/quadtree.go

package quadtree

type Point struct {
  X, Y float64
}

type QtNode struct {
  X, Y, W, H     float64 // position and size
  Cap            int     // max number of points in a node
  Pts            []Point // points in this node
  Bl, Tl, Tr, Br *QtNode // child nodes
}

// New creates a new quadtree node
func New(x float64, y float64, w float64, h float64, cap int) *QtNode {
  return &QtNode{x, y, w, h, cap, make([]Point, 0), nil, nil, nil, nil}
}

// Insert inserts a point into the quadtree
func (n *QtNode) Insert(pt Point) {
  if !n.ContainsPt(pt) {
    return
  }

  n.Pts = append(n.Pts, pt)

  if n.IsSubdivided() {
    div := n.GetDivContainingPt(pt)
    if div != nil {
      div.Insert(pt)
    }
  } else if len(n.Pts) > n.Cap {
    n.subdivide()
  }
}

// subdivide divides the node into 4 child nodes
func (n *QtNode) subdivide() {
  w, h := n.W*.5, n.H*.5
  n.Bl = New(n.X, n.Y, w, h, n.Cap)
  n.Tl = New(n.X, n.Y+h, w, h, n.Cap)
  n.Tr = New(n.X+w, n.Y+h, w, h, n.Cap)
  n.Br = New(n.X+w, n.Y, w, h, n.Cap)

  for _, pt := range n.Pts {
    div := n.GetDivContainingPt(pt)
    if div != nil {
      div.Insert(pt)
    }
  }
}

func (n *QtNode) IsSubdivided() bool {
  return n.Bl != nil && n.Tl != nil && n.Tr != nil && n.Br != nil
}

func (n *QtNode) GetDivContainingPt(pt Point) *QtNode {
  if n.Bl.ContainsPt(pt) {
    return n.Bl
  } else if n.Tl.ContainsPt(pt) {
    return n.Tl
  } else if n.Tr.ContainsPt(pt) {
    return n.Tr
  } else if n.Br.ContainsPt(pt) {
    return n.Br
  }

  return nil
}

func (n *QtNode) ContainsPt(pt Point) bool {
  return pt.X >= n.X && pt.X <= n.X+n.W && pt.Y >= n.Y && pt.Y <= n.Y+n.H
}
```

The `New` function creates a new quadtree with a given position, dimensions and capacity (`cap`).

`Insert` adds a `Point` to the current `QTNode`.  If the current node is already subdivided, insert will look at which of it's children should contain the point and calls the appropriate child's insert function.  If the node is not subdivided, `Insert` will subdivide the node if it's capacity is reached by the **Point** we just added.

`subdivide` create 4 new `QTNode` objects for the bottom left (`Bl`), top left (`Tl`), top right (`Tr`) and bottom right (`Br`) quadrants.  Next it iterates over each of the current nodes Points (`Pts`) and divides them up among the appropriate quadrants.  **_Note:_** Event though we divide the current nodes points among it's children, parent nodes always retain their on `Pts` list.  This will make querying much simpler later on as we can decide which branches of the tree we need to follow to find specific points.

`IsSubdivided`, `GetDivContainingPt` and `ContainsPt` are utility functions used in the code we've already covered.

And thats it for the basic quadtree.  Simple right?  On to querying.

## Querying a quadtree

So we've seen how we can create a point quadtree, organizing a bunch of nodes into dynamically decomposing cells.  Now we need a way to query it to make useful.  I'm going to show you my implementation of a box query.

A box query will search the quadtree for all nodes stored inside a rectangle which we define by its position, its width and its height.  We will travers the quadtree from it's root, searching nodes which overlap with the query box.

Again I'l start with an example.  In the demo below you'll see a red box move through the canvas, highlighting any points it overlaps.

<div align="center">
  <iframe src="https://quadtree-demo-opal.vercel.app?width=400&height=400&pts=50&demoQuery" width="410" height="410" frameBorder="0"></iframe>
</div>

### Box Query

The code below is all you need to box query a quad tree. `boxQuery` recursively traverses the quad tree return all points within the box.  Of course we could check each point individually for containment in the query box, but with a huge numbers of points that would be very slow.  Instead we will query the nodes of the quadtree for containment or intersection with the query box.

Here's the code and I'll talk about how it works in a moment.

```go
// pkg/quadtree/boxQuery.go

package quadtree

type Rect struct {
  L, T, R, B float64
}

func (n *QtNode) BoxQuery(qry Rect) []Point {
  var pts = []Point{}

  if len(n.Pts) == 0 {
    return pts
  }

  node := Rect{L: n.X, T: n.Y + n.H, R: n.X + n.W, B: n.Y} // node rect

  qryContainsNode := qry.L <= node.L && qry.T >= node.T && qry.R >= node.R && qry.B <= node.B
  nodeContainsQry := node.L <= qry.L && node.T >= qry.T && node.R >= qry.R && node.B <= qry.B
  qryIntersectNode := node.L <= qry.R && node.T >= qry.B && node.R >= qry.L && node.B <= qry.T

  qryContainsPt := func(pt Point) bool { return qry.L <= pt.X && qry.T >= pt.Y && qry.R >= pt.X && qry.B <= pt.Y }

  if qryContainsNode {
    pts = append(pts, n.Pts...)
  } else if qryIntersectNode || nodeContainsQry {
    if n.IsSubdivided() { // if subdivided, recurse
      pts = append(pts, n.Bl.BoxQuery(qry)...)
      pts = append(pts, n.Tl.BoxQuery(qry)...)
      pts = append(pts, n.Tr.BoxQuery(qry)...)
      pts = append(pts, n.Br.BoxQuery(qry)...)
      return pts
    } else {
      // check each point for inclusion in qr
      for _, pt := range n.Pts {
        if qryContainsPt(pt) {
          pts = append(pts, pt)
        }
      }
    }
  }

  return pts
}

```

If `qryContainsNode` is true then we can return all points in the current node, no need to check the children as they will be contained also.

If `qryIntersectNode` or `nodeContainsQry` then the query box overlaps the bounds of the node and may or may not contain it's points.  If the node is subdivided then we will recurse to the child nodes and perform the same intersections test on them, appending any resulting points to `pts`.

Finally we will have recursed to a node that is not fully contained by the query box and is not subdivided. Now we can brute force it using the inline function `qryContainsPt`, checking each point in the nodes `Pts` list for containment in the query box.

`BoxQuery` allows us to substitute potentially expensive point containment tests with a handful of rect intersection tests.  This is a huge performance boost when dealing with large numbers of points.

## Conclusion

That's about it for this post.  This was more a note to self than anything else, as are all my posts because I don't expect anyone actually reads them.  I really just do this as a way to practice communicating about code and engraining knowledge.  If you did read this, thanks!  I hope you found it useful and maybe send me a message on my contact page if you have any suggestions or comments.

☮
