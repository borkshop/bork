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

// Cursor returns a cursor that will scan over relations with given type and
// that meet the optional where clause.
func (rel *Relation) Cursor(
	tcl TypeClause,
	where func(r ComponentType, ent, a, b Entity) bool,
) Cursor {
	it := rel.Iter(tcl)
	return &iterCursor{rel: rel, it: it, where: where}
}

// LookupA returns a Cursor that will iterate over relations involving one or
// more given A entities.
func (rel *Relation) LookupA(tcl TypeClause, ids ...EntityID) Cursor {
	// TODO: indexing
	return rel.scanLookup(tcl, false, ids)
}

// LookupB returns a Cursor that will iterate over relations involving one or
// more given B entities.
func (rel *Relation) LookupB(tcl TypeClause, ids ...EntityID) Cursor {
	// TODO: indexing
	return rel.scanLookup(tcl, true, ids)
}

// Insert relations under the given type clause. TODO: constraints, indices,
// etc.
func (rel *Relation) Insert(r ComponentType, a, b Entity) Entity {
	return rel.insert(r, a, b)
}

// InsertMany allows a function to insert many relations without incurring
// indexing cost; indexing is deferred until the with function returns, at
// which point indices are fixed.
func (rel *Relation) InsertMany(with func(func(r ComponentType, a, b Entity) Entity)) {
	with(rel.insert)
}

func (rel *Relation) insert(r ComponentType, a, b Entity) Entity {
	aid := rel.aCore.Deref(a)
	bid := rel.bCore.Deref(b)
	ent := rel.AddEntity(ComponentType(r))
	i := int(ent.ID()) - 1
	rel.aids[i] = aid
	rel.bids[i] = bid
	return ent
}

// Delete all relations matching the given type clause and optional where
// function; this is like Update with a set function that zeros the relation,
// but marginally faster / simpler.
func (rel *Relation) Delete(
	tcl TypeClause,
	where func(r ComponentType, ent, a, b Entity) bool,
) {
	for cur := rel.Cursor(tcl, where); cur.Scan(); {
		rel.setType(cur.Entity().ID(), NoType)
	}
}

// Upsert updates and/or inserts relations:
//
// The `each` function may call `emit` 0 or more times for each relation
// entity; `emit` will return a, maybe newly inserted, entity for the given
// `r`, `a`, `b` triple that its given.
//
// If emit isn't called for an entity, then it is destroyed. The first time
// `emit` is called the entity is updated; thereafter a new entity is inserted.
//
// If no relations matched, then the each function is called exactly once with
// NilEntity for r, a, and b.
//
// It returns the total number of `emit()`ed relations, and the total number of
// matched relations.
func (rel *Relation) Upsert(
	tcl TypeClause,
	where func(r ComponentType, ent, a, b Entity) bool,
	each func(
		r ComponentType, ent, a, b Entity,
		emit func(r ComponentType, a, b Entity) Entity,
	),
) (n, m int) {
	for cur := rel.Cursor(tcl, where); cur.Scan(); {
		ent, any := cur.Entity(), false
		each(cur.R(), ent, cur.A(), cur.B(), func(er ComponentType, ea, eb Entity) Entity {
			if any {
				if ea == NilEntity || eb == NilEntity {
					return NilEntity
				}
				n++
				return rel.insert(er, ea, eb)
			}
			any = true
			if rel.doUpdate(ent, cur.R(), cur.A(), cur.B(), er, ea, eb) {
				n++
			}
			return ent
		})
		if !any {
			ent.Destroy()
		}
	}
	if n == 0 {
		each(0, NilEntity, NilEntity, NilEntity, func(er ComponentType, ea, eb Entity) Entity {
			if ea == NilEntity || eb == NilEntity {
				return NilEntity
			}
			n++
			return rel.insert(er, ea, eb)
		})
	}
	return n, m
}

func (rel *Relation) doUpdate(
	ent Entity,
	or ComponentType, oa, ob Entity,
	nr ComponentType, na, nb Entity,
) bool {
	if nr == NoType || na == NilEntity || nb == NilEntity {
		ent.Destroy()
		return false
	}
	if ent.Type() == NoType {
		return false
	}
	if nr != or {
		ent.SetType(ComponentType(nr))
	}
	i := ent.ID() - 1
	if na != oa {
		rel.aids[i] = na.ID()
	}
	if nb != ob {
		rel.bids[i] = nb.ID()
	}
	return true
}
