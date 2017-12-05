package ecs

// TraversalMode represents a graph traversal mode.
type TraversalMode uint8

const (
	// TraverseDFS is Depth First Search traversal, starting from all matching
	// roots.
	TraverseDFS TraversalMode = 1 << iota
	traverseCo

	// TraverseCoDFS is Reversed Depth First Search traversal, starting from
	// all matching leaves.
	TraverseCoDFS = traverseCo | TraverseDFS

	// TODO: other modes like BFS
)

// Traverse returns a new graph travers for the given type clause and mode.
func (G *Graph) Traverse(tcl TypeClause, mode TraversalMode) GraphTraverser {
	switch mode {
	case TraverseDFS, TraverseCoDFS:
		return &dfsTraverser{
			g:    G,
			tcl:  tcl,
			mode: mode,
		}

	default:
		panic("invalid graph traversal mode")
	}
}

// GraphTraverser traverses a graph in some order.
type GraphTraverser interface {
	Init(seed ...EntityID)
	Traverse() bool
	G() *Graph
	Edge() Entity
	Node() Entity
}

type dfsTraverser struct {
	g    *Graph
	tcl  TypeClause
	mode TraversalMode
	seen map[EntityID]struct{}
	edge EntityID
	node EntityID
	curs []Cursor
	q    []EntityID
}

func (gt *dfsTraverser) G() *Graph    { return gt.g }
func (gt *dfsTraverser) Edge() Entity { return gt.g.Ref(gt.edge) }
func (gt *dfsTraverser) Node() Entity { return gt.g.aCore.Ref(gt.node) }
func (gt *dfsTraverser) Traverse() bool {
	for gt.traverse() {
		if _, seen := gt.seen[gt.node]; !seen {
			gt.seen[gt.node] = struct{}{}
			return true
		}
	}
	return false
}

func (gt *dfsTraverser) traverse() bool {
	if gt.node != 0 {
		var cur Cursor
		if gt.mode&traverseCo == 0 {
			cur = gt.g.Select(gt.tcl, InA(gt.node))
		} else {
			cur = gt.g.Select(gt.tcl, InB(gt.node))
		}
		if cur.Scan() {
			gt.curs = append(gt.curs, cur)
			gt.setState(cur)
			return true
		}
		for i := len(gt.curs) - 1; i >= 0; i-- {
			if cur := gt.curs[i]; cur.Scan() {
				gt.setState(cur)
				return true
			}
			gt.curs = gt.curs[:i]
		}
		gt.edge = 0
		gt.node = 0
	}
	if i := len(gt.q) - 1; i >= 0 {
		gt.node = gt.q[i]
		gt.q = gt.q[:i]
		return true
	}
	return false
}

func (gt *dfsTraverser) setState(cur Cursor) {
	gt.edge = cur.R().id
	if gt.mode&traverseCo == 0 {
		gt.node = cur.B().id
	} else {
		gt.node = cur.A().id
	}
}

func (gt *dfsTraverser) Init(seed ...EntityID) {
	if len(gt.seen) > 0 {
		for id := range gt.seen {
			delete(gt.seen, id)
		}
	} else {
		// TODO: shave down this estimate?
		gt.seen = make(map[EntityID]struct{}, gt.g.Len())
	}
	gt.edge = 0
	gt.node = 0
	gt.q = gt.q[:0]

	if len(seed) > 0 {
		gt.q = append(gt.q, seed...)
	} else {
		var (
			triset map[EntityID]bool
			n      int
		)
		if gt.mode&traverseCo == 0 {
			triset, n = gt.g.roots(gt.tcl, nil)
		} else {
			triset, n = gt.g.leaves(gt.tcl, nil)
		}
		if n <= 0 {
			return
		}

		if cap(gt.q) < n {
			gt.q = make([]EntityID, 0, n)
		}
		for id, in := range triset {
			if in {
				gt.q = append(gt.q, id)
			}
		}
	}
}
