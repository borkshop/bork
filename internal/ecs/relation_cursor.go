package ecs

// Cursor iterates through a Relation.
type Cursor interface {
	Scan() bool
	Count() int
	Entity() Entity
	R() RelationType
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

	it    Iterator
	where func(r RelationType, ent, a, b Entity) bool

	ent Entity
	a   Entity
	r   RelationType
	b   Entity
}

func (cur iterCursor) Count() int {
	if cur.where == nil {
		return cur.it.Count()
	}

	n := 0
	it := cur.it
	for it.Next() {
		ent := it.Entity()
		i := ent.ID() - 1
		r := RelationType(cur.rel.types[i] & ^relType)
		a := cur.rel.aCore.Ref(cur.rel.aids[i])
		b := cur.rel.aCore.Ref(cur.rel.bids[i])
		if cur.where(r, ent, a, b) {
			n++
		}
	}
	return n
}

func (cur *iterCursor) Scan() bool {
	for cur.it.Next() {
		cur.ent = cur.it.Entity()
		i := cur.ent.ID() - 1
		cur.r = RelationType(cur.rel.types[i] & ^relType)
		cur.a = cur.rel.aCore.Ref(cur.rel.aids[i])
		cur.b = cur.rel.aCore.Ref(cur.rel.bids[i])
		if cur.where == nil || cur.where(cur.r, cur.ent, cur.a, cur.b) {
			return true
		}
	}
	cur.ent = NilEntity
	cur.r = 0
	cur.a = NilEntity
	cur.b = NilEntity
	return false
}

func (cur iterCursor) Entity() Entity  { return cur.ent }
func (cur iterCursor) R() RelationType { return cur.r }
func (cur iterCursor) A() Entity       { return cur.a }
func (cur iterCursor) B() Entity       { return cur.b }
