package ecs

import "sort"

// AddAIndex adds an index for A-side entity IDs.
func (rel *Relation) AddAIndex() {
	rel.aix = make([]int, len(rel.aids))
	for i := range rel.aids {
		rel.aix[i] = i
	}
	sortit(len(rel.aix), rel.aixLess, rel.aixSwap)
}

// AddBIndex adds an index for B-side entity IDs.
func (rel *Relation) AddBIndex() {
	rel.bix = make([]int, len(rel.bids))
	for i := range rel.bids {
		rel.bix[i] = i
	}
	sortit(len(rel.bix), rel.bixLess, rel.bixSwap)
}

func (rel *Relation) indexLookup(
	tcl TypeClause,
	qids, aids []EntityID,
	aix []int,
) Cursor {
	return &indexCursor{
		rel:  rel,
		tcl:  tcl,
		qids: qids,
		ids:  aids,
		ix:   aix,
	}
}

func (rel Relation) aixLess(i, j int) bool { return rel.aids[rel.aix[i]] < rel.aids[rel.aix[j]] }
func (rel Relation) bixLess(i, j int) bool { return rel.bids[rel.bix[i]] < rel.bids[rel.bix[j]] }

func (rel Relation) aixSwap(i, j int) { rel.aix[i], rel.aix[j] = rel.aix[j], rel.aix[i] }
func (rel Relation) bixSwap(i, j int) { rel.bix[i], rel.bix[j] = rel.bix[j], rel.bix[i] }

func (rel *Relation) deferIndexing() func() {
	if rel.aix == nil && rel.bix == nil {
		return nil
	}
	rel.fix = true
	return func() {
		if rel.aix != nil {
			sortit(len(rel.aix), rel.aixLess, rel.aixSwap)
		}
		if rel.bix != nil {
			sortit(len(rel.aix), rel.aixLess, rel.aixSwap)
		}
		rel.fix = false
	}
}

type tmpSort struct {
	n    int
	less func(i, j int) bool
	swap func(i, j int)
}

func (ts tmpSort) Len() int           { return ts.n }
func (ts tmpSort) Less(i, j int) bool { return ts.less(i, j) }
func (ts tmpSort) Swap(i, j int)      { ts.swap(i, j) }

func fix(
	i, n int,
	less func(i, j int) bool,
	swap func(i, j int),
) {
	// TODO: something more minimal, since we assume sorted order but for [i]
	sortit(n, less, swap)
}

func sortit(
	n int,
	less func(i, j int) bool,
	swap func(i, j int),
) {
	sort.Sort(tmpSort{n, less, swap})
}

type indexCursor struct {
	rel       *Relation
	tcl       TypeClause
	qidi      int
	qids, ids []EntityID
	ix        []int

	r             RelationType
	ent, aid, bid EntityID
}

func (ixc *indexCursor) scan() bool {
	for ixc.qidi < len(ixc.qids) {
		ixc.ent = ixc.qids[ixc.qidi]
		ixc.qidi++
		for i := sort.Search(len(ixc.ix), ixc.search); i < len(ixc.ix) && ixc.ids[ixc.ix[i]] == ixc.ent; i++ {
			if j := ixc.ix[i]; ixc.tcl.Test(ixc.rel.types[j]) {
				return true
			}
		}
	}
	return false
}

func (ixc indexCursor) search(i int) bool {
	return ixc.ids[ixc.ix[i]] >= ixc.ent
}

func (ixc *indexCursor) Scan() bool {
	if ixc.scan() {
		ixc.r = RelationType(ixc.rel.types[ixc.ent-1] & ^relType)
		return true
	}
	return false
}

func (ixc indexCursor) Count() int {
	n := 0
	for ixc.scan() {
		n++
	}
	return n
}

func (ixc indexCursor) Entity() Entity  { return ixc.rel.Ref(ixc.ent) }
func (ixc indexCursor) R() RelationType { return ixc.r }
func (ixc indexCursor) A() Entity       { return ixc.rel.aCore.Ref(ixc.aid) }
func (ixc indexCursor) B() Entity       { return ixc.rel.bCore.Ref(ixc.bid) }
