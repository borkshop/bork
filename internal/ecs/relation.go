package ecs

// RelationFlags specifies options for the A or B dimension in a Relation.
type RelationFlags uint32

const (
	// RelationCascadeDestroy causes destruction of an entity relation to
	// destroy related entities within the flagged dimension.
	RelationCascadeDestroy RelationFlags = 1 << iota

	// RelationRestrictDeletes TODO: cannot abort a destroy at present
)

// Relation contains entities that represent relations between entities in two
// (maybe different) Cores. Users may attach arbitrary data to these relations
// the same way you would with Core.
type Relation struct {
	Core
	aCore, bCore *Core
	aFlag, bFlag RelationFlags
	aids         []EntityID
	bids         []EntityID
}

// NewRelation creates a new relation for the given Core systems.
func NewRelation(
	aCore *Core, aFlags RelationFlags,
	bCore *Core, bFlags RelationFlags,
) *Relation {
	rel := &Relation{}
	rel.Init(aCore, aFlags, bCore, bFlags)
	return rel
}

// Init initializes the entity relation; useful for embedding.
func (rel *Relation) Init(
	aCore *Core, aFlags RelationFlags,
	bCore *Core, bFlags RelationFlags,
) {
	rel.aCore, rel.aFlag = aCore, aFlags
	rel.bCore, rel.bFlag = bCore, bFlags
	rel.RegisterAllocator(NoType, rel.allocRel)
	rel.RegisterDestroyer(NoType, rel.destroyRel)
	rel.aCore.RegisterDestroyer(NoType, rel.destroyFromA)
	if rel.aCore != rel.bCore {
		rel.bCore.RegisterDestroyer(NoType, rel.destroyFromB)
	}
}

// A returns a reference to the A-side entity for the given relation entity.
func (rel *Relation) A(ent Entity) Entity {
	return rel.aCore.Ref(rel.aids[rel.Deref(ent)-1])
}

// B returns a reference to the B-side entity for the given relation entity.
func (rel *Relation) B(ent Entity) Entity {
	return rel.bCore.Ref(rel.bids[rel.Deref(ent)-1])
}

func (rel *Relation) allocRel(id EntityID, t ComponentType) {
	rel.aids = append(rel.aids, 0)
	rel.bids = append(rel.bids, 0)
}

func (rel *Relation) destroyRel(id EntityID, t ComponentType) {
	i := int(id) - 1
	if aid := rel.aids[i]; aid != 0 {
		if rel.aFlag&RelationCascadeDestroy != 0 {
			rel.aCore.setType(aid, NoType)
		}
		rel.aids[i] = 0
	}
	if bid := rel.bids[i]; bid != 0 {
		if rel.bFlag&RelationCascadeDestroy != 0 {
			rel.bCore.setType(bid, NoType)
		}
		rel.bids[i] = 0
	}
}

func (rel *Relation) destroyFromA(aid EntityID, t ComponentType) {
	for i := range rel.types {
		if rel.aids[i] == aid {
			rel.setType(EntityID(i+1), NoType)
		}
	}
}

func (rel *Relation) destroyFromB(bid EntityID, t ComponentType) {
	for i := range rel.types {
		if rel.bids[i] == bid {
			rel.setType(EntityID(i+1), NoType)
		}
	}
}

// Select creates a cursor with the given options applied. If none are given,
// TrueClause is used; so the default is basically "select all".
func (rel *Relation) Select(opts ...CursorOpt) Cursor {
	if len(opts) == 0 {
		return TrueClause.apply(rel, nil)
	}
	cur := opts[0].apply(rel, nil)
	for _, opt := range opts[1:] {
		cur = opt.apply(rel, cur)
	}
	return cur
}

// Upsert updates any relations that the given cursor iterates, and may insert
// new ones.
//
// If the each function is nil, all matched relations are destroyed.
//
// If the cursor is nil, then each is called exactly once with an empty cursor
// that it should use to create new relations.
func (rel *Relation) Upsert(cur Cursor, each func(*UpsertCursor)) (n, m int) {
	if each == nil {
		for cur.Scan() {
			rel.setType(cur.R().ID(), NoType)
			m++
		}
		return n, m
	}

	uc := UpsertCursor{rel: rel, Cursor: cur}
	if cur == nil {
		each(&uc)
		return uc.n, 0
	}
	for uc.Scan() {
		each(&uc)
	}
	if uc.n == 0 {
		each(&uc)
	}
	return uc.n, m
}

// UpsertCursor allows inserting, updating, and deleting relations.
type UpsertCursor struct {
	Cursor
	rel  *Relation
	last Entity
	any  bool
	n    int
}

// Scan advances the underlying cursor; but first, it destroys the last scanned
// relation if no updated record was emitted.
func (uc *UpsertCursor) Scan() bool {
	if uc.last != NilEntity && uc.any {
		uc.last.Destroy()
	}
	uc.any = false
	if uc.Cursor.Scan() {
		uc.last = uc.R()
		return true
	}
	uc.last = NilEntity
	return false
}

// Emit a record, replacing the current, or inserting a new one if the current
// record has already been updated.
func (uc *UpsertCursor) Emit(er ComponentType, ea, eb Entity) Entity {
	if uc.any {
		return uc.Create(er, ea, eb)
	}
	uc.any = true
	rel := uc.R()
	if rel == NilEntity {
		return NilEntity
	}
	if er == NoType || ea == NilEntity || eb == NilEntity {
		rel.Destroy()
		return NilEntity
	}
	if er != rel.Type() {
		rel.SetType(ComponentType(er))
	}
	i := rel.ID() - 1
	if ea != uc.A() {
		uc.rel.aids[i] = ea.ID()
	}
	if eb != uc.B() {
		uc.rel.bids[i] = eb.ID()
	}
	uc.n++
	return rel
}

// Create a new relation, ignoring the current; when bulk loading data (no
// underlying Cursor), this is the prefered method.
func (uc *UpsertCursor) Create(r ComponentType, a, b Entity) Entity {
	if a == NilEntity || b == NilEntity {
		return NilEntity
	}
	aid := uc.rel.aCore.Deref(a)
	bid := uc.rel.bCore.Deref(b)
	rel := uc.rel.AddEntity(ComponentType(r))
	i := int(rel.ID()) - 1
	uc.rel.aids[i] = aid
	uc.rel.bids[i] = bid
	uc.n++
	return rel
}
