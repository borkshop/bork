package ecs

// Cursor iterates through a Relation.
type Cursor interface {
	Scan() bool
	Count() int
	R() Entity
	A() Entity
	B() Entity
}

func (rel *Relation) scanLookup(tcl TypeClause, co bool, qids []EntityID) Cursor {
	// TODO: if qids is big enough, build a set first
	if co {
		return &coScanCursor{
			qids: qids,
			iterCursor: iterCursor{
				rel: rel,
				it:  rel.Iter(tcl),
			},
		}
	}
	return &scanCursor{
		qids: qids,
		iterCursor: iterCursor{
			rel: rel,
			it:  rel.Iter(tcl),
		},
	}
}

type scanCursor struct {
	iterCursor
	qids []EntityID
}

func (sc *scanCursor) Scan() bool {
	for sc.iterCursor.Scan() {
		id := sc.iterCursor.a.ID()
		for _, qid := range sc.qids {
			if qid == id {
				return true
			}
		}
	}
	return false
}

type coScanCursor scanCursor

func (csc *coScanCursor) Scan() bool {
	for csc.iterCursor.Scan() {
		id := csc.iterCursor.b.ID()
		for _, qid := range csc.qids {
			if qid == id {
				return true
			}
		}
	}
	return false
}

type iterCursor struct {
	rel *Relation
	it  Iterator
	r   Entity
	a   Entity
	b   Entity
}

func (cur iterCursor) Count() int {
	return cur.it.Count()
}

func (cur *iterCursor) Scan() bool {
	if cur.it.Next() {
		i := cur.it.ID() - 1
		cur.r = cur.it.Entity()
		cur.a = cur.rel.aCore.Ref(cur.rel.aids[i])
		cur.b = cur.rel.aCore.Ref(cur.rel.bids[i])
		return true
	}
	cur.r = NilEntity
	cur.a = NilEntity
	cur.b = NilEntity
	return false
}

func (cur iterCursor) R() Entity { return cur.r }
func (cur iterCursor) A() Entity { return cur.a }
func (cur iterCursor) B() Entity { return cur.b }

// FilterCursor returns a cursor filtered by the given
// predicate function; its Count() method may ignore the
// filter, drastically over-counting.
func FilterCursor(cur Cursor, f func(Cursor) bool) Cursor {
	switch impl := cur.(type) {
	case (*iterCursor):
		return iterFilterCursor{iterCursor: *impl}.with(f)

	case (*iterFilterCursor):
		return impl.with(f)

	case (filterCursor):
		return impl.with(f)

	default:
		return filterCursor{Cursor: cur}.with(f)
	}
}

type iterFilterCursor struct {
	iterCursor
	fs []func(Cursor) bool
}

type filterCursor struct {
	Cursor
	fs []func(Cursor) bool
}

func (ifc iterFilterCursor) with(f func(Cursor) bool) Cursor {
	ifc.fs = append(ifc.fs[:len(ifc.fs):len(ifc.fs)], f)
	return &ifc
}

func (fc filterCursor) with(f func(Cursor) bool) Cursor {
	fc.fs = append(fc.fs[:len(fc.fs):len(fc.fs)], f)
	return fc
}

func (ifc *iterFilterCursor) Scan() bool {
scan:
	for ifc.iterCursor.Scan() {
		for _, f := range ifc.fs {
			if !f(ifc) {
				continue scan
			}
		}
		return true
	}
	return false
}

func (ifc iterFilterCursor) Count() (n int) {
scan:
	for ifc.iterCursor.Scan() {
		for _, f := range ifc.fs {
			if !f(&ifc) {
				continue scan
			}
		}
		n++
	}
	return n
}

func (fc filterCursor) Scan() bool {
scan:
	for fc.Cursor.Scan() {
		for _, f := range fc.fs {
			if !f(fc) {
				continue scan
			}
		}
		return true
	}
	return false
}

// TODO obvious overcount func (fc filterCursor) Count() int
