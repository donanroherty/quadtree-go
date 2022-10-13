package quadtree

// Point is a point in 2D space
type Point struct {
	X, Y float64
}

// QtNode is a quadtree node
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

// IsSubdivided returns true if the node is subdivided
func (n *QtNode) IsSubdivided() bool {
	return n.Bl != nil && n.Tl != nil && n.Tr != nil && n.Br != nil
}

// GetDivContainingPt returns the subdivision containing the point
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

// ContainsPt returns true if the node contains the point
func (n *QtNode) ContainsPt(pt Point) bool {
	return pt.X >= n.X && pt.X <= n.X+n.W && pt.Y >= n.Y && pt.Y <= n.Y+n.H
}
