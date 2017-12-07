package ecs

// CursorOpt specifies which relational entites are scanned by the cursor.
// TypeClause implements CursorOpt, and is the most basic one.
type CursorOpt interface {
	apply(*Relation, Cursor) Cursor
}

func (cc constClause) apply(rel *Relation, cur Cursor) Cursor {
	return typeClauseOpt{cc}.apply(rel, cur)
}
func (t allClause) apply(rel *Relation, cur Cursor) Cursor    { return typeClauseOpt{t}.apply(rel, cur) }
func (t anyClause) apply(rel *Relation, cur Cursor) Cursor    { return typeClauseOpt{t}.apply(rel, cur) }
func (t notAllClause) apply(rel *Relation, cur Cursor) Cursor { return typeClauseOpt{t}.apply(rel, cur) }
func (t notAnyClause) apply(rel *Relation, cur Cursor) Cursor { return typeClauseOpt{t}.apply(rel, cur) }
func (tcls andClause) apply(rel *Relation, cur Cursor) Cursor {
	return typeClauseOpt{tcls}.apply(rel, cur)
}
func (tcls orClause) apply(rel *Relation, cur Cursor) Cursor {
	return typeClauseOpt{tcls}.apply(rel, cur)
}

type typeClauseOpt struct{ TypeClause }

func (tco typeClauseOpt) apply(rel *Relation, cur Cursor) Cursor {
	if cur == nil {
		return &iterCursor{
			rel: rel,
			it:  rel.Iter(tco.TypeClause),
		}
	}

	if nc := tco.justApply(rel, cur); nc != nil {
		return nc
	}

	return filterCursor{Cursor: cur}.with(tco.filter)
}

func (tco typeClauseOpt) filter(cur Cursor) bool { return tco.test(cur.R().Type()) }

func (tco typeClauseOpt) justApply(rel *Relation, cur Cursor) Cursor {
	switch impl := cur.(type) {
	case *iterCursor:
		impl.it.tcl = and(impl.it.tcl, tco.TypeClause)
		return impl

	case *iterFilterCursor:
		impl.it.tcl = and(impl.it.tcl, tco.TypeClause)
		return impl

	case filterCursor:
		if nc := tco.justApply(rel, impl.Cursor); nc != nil {
			impl.Cursor = nc
			return impl
		}

	}
	return nil
}

// Cursor iterates through a Relation.
type Cursor interface {
	Scan() bool
	Count() int
	R() Entity
	A() Entity
	B() Entity
}

// InA returns a cursor option that limits the cursor to relations involving
// one or more given A-side entities.
func InA(ids ...EntityID) CursorOpt {
	// TODO: if ids is big enough, build a set
	return lookupAOpt(ids)
}

// InB returns a cursor option that limits the cursor to relations involving
// one or more given A-side entities.
func InB(ids ...EntityID) CursorOpt {
	// TODO: if ids is big enough, build a set
	return lookupBOpt(ids)
}

type lookupAOpt []EntityID
type lookupBOpt []EntityID

// TODO indexing support beyond a filter predicate
func (lko lookupAOpt) apply(rel *Relation, cur Cursor) Cursor {
	return Filter(lko.filter).apply(rel, cur)
}
func (lko lookupBOpt) apply(rel *Relation, cur Cursor) Cursor {
	return Filter(lko.filter).apply(rel, cur)
}

func (lko lookupAOpt) filter(cur Cursor) bool {
	id := cur.A().ID()
	for _, qid := range lko {
		if qid == id {
			return true
		}
	}
	return false
}
func (lko lookupBOpt) filter(cur Cursor) bool {
	id := cur.B().ID()
	for _, qid := range lko {
		if qid == id {
			return true
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

// Filter is a CursorOpt that applies a filtering
// predicate function to a cursor. NOTE this the
// returned Cursor's Count() method may ignore the
// filter, drastically over-counting.
type Filter func(Cursor) bool

func (f Filter) apply(rel *Relation, cur Cursor) Cursor {
	if cur == nil {
		return f.apply(rel, TrueClause.apply(rel, nil))
	}

	switch impl := cur.(type) {
	case *iterCursor:
		return iterFilterCursor{iterCursor: *impl}.with(f)

	case *iterFilterCursor:
		return impl.with(f)

	case filterCursor:
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
