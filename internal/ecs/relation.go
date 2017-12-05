package ecs

const relType ComponentType = 1 << (63 - iota)

// Relation contains entities that represent relations between entities in two
// (maybe different) Cores. Users may attach arbitrary data to these relations
// the same way you would with Core.
//
// NOTE: the high Type bit (bit 64) is reserved.
type Relation struct {
	Core
	aCore, bCore *Core
	aids         []EntityID
	bids         []EntityID
}

// NewRelation creates a new relation for the given Core systems.
func NewRelation(
	aCore *Core,
	bCore *Core,
) *Relation {
	rel := &Relation{}
	rel.Init(aCore, bCore)
	return rel
}

// Init initializes the entity relation; useful for embedding.
func (rel *Relation) Init(
	aCore *Core,
	bCore *Core,
) {
	rel.aCore = aCore
	rel.bCore = bCore
	rel.RegisterAllocator(relType, rel.allocRel)
	rel.RegisterDestroyer(relType, rel.destroyRel)
}

// A returns a reference to the A-side entity for the given relation entity.
func (rel *Relation) A(ent Entity) Entity {
	if ent.Type().HasAll(relType) {
		return rel.aCore.Ref(rel.aids[rel.Deref(ent)-1])
	}
	return NilEntity
}

// B returns a reference to the B-side entity for the given relation entity.
func (rel *Relation) B(ent Entity) Entity {
	if ent.Type().HasAll(relType) {
		return rel.bCore.Ref(rel.bids[rel.Deref(ent)-1])
	}
	return NilEntity
}

func (rel *Relation) allocRel(id EntityID, t ComponentType) {
	rel.aids = append(rel.aids, 0)
	rel.bids = append(rel.bids, 0)
}

func (rel *Relation) destroyRel(id EntityID, t ComponentType) {
	i := int(id) - 1
	if aid := rel.aids[i]; aid != 0 {
		rel.aids[i] = 0
	}
	if bid := rel.bids[i]; bid != 0 {
		rel.bids[i] = 0
	}
}

// RelationType specified the type of a relation, it's basically a
// ComponentType where the highest bit is reserved.
type RelationType uint64

// NoRelType is the RelationType equivalent of NoType.
const NoRelType RelationType = 0

// All returns true only if all of the masked type bits are set.
func (t RelationType) All(mask RelationType) bool { return t&mask == mask }

// Any returns true only if at least one of the masked type bits is set.
func (t RelationType) Any(mask RelationType) bool { return t&mask != 0 }
