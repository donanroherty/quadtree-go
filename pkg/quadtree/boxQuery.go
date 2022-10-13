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
