package ecs

// Iter returns a new iterator over the Core's entities which satisfy the
// given TypeClause.
func (co *Core) Iter(tcl TypeClause) Iterator { return Iterator{co, -1, tcl} }

// Iterator points into a Core's entities, iterating over them with optional
// type filter criteria.
type Iterator struct {
	co  *Core
	i   int
	tcl TypeClause
}

// Next advances the iterator to point at the next matching entity, and
// returns true if such an entity was found; otherwise iteration is done, and
// false is returned.
func (it *Iterator) Next() bool {
	for it.i++; it.i < len(it.co.types); it.i++ {
		t := it.co.types[it.i]
		if it.tcl.Test(t) {
			return true
		}
	}
	return false
}

// Reset resets the iterator, causing it to start over.
func (it *Iterator) Reset() { it.i = -1 }

// Count counts how many entities remain to be iterated, without advancing the
// iterator.
func (it Iterator) Count() int {
	n := 0
	for it.Next() {
		n++
	}
	return n
}

// Type returns the type of the current entity, or NoType if iteration is
// done.
func (it Iterator) Type() ComponentType {
	if it.i < len(it.co.types) {
		return it.co.types[it.i]
	}
	return NoType
}

// ID returns the type of the current entity, or 0 if iteration is done.
func (it Iterator) ID() EntityID {
	if it.i < len(it.co.types) {
		return EntityID(it.i + 1)
	}
	return 0
}

// Entity returns a reference to the current entity, or NilEntity if
// iteration is done.
func (it Iterator) Entity() Entity {
	if it.i < len(it.co.types) {
		return Entity{it.co, EntityID(it.i + 1)}
	}
	return NilEntity
}
