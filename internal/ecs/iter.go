package ecs

// Iterator is an entity iterator.
type Iterator interface {
	Next() bool
	Reset()
	Count() int
	Type() ComponentType
	ID() EntityID
	Entity() Entity
}

// Iter returns a new iterator over the Core's entities which satisfy all of
// given TypeClause(s).
func (co *Core) Iter(tcls ...TypeClause) Iterator {
	switch len(tcls) {
	case 0:
		return &coreIterator{co, -1, TrueClause}
	case 1:
		return &coreIterator{co, -1, tcls[0]}
	default:
		return &coreIterator{co, -1, And(tcls...)}
	}
}

// coreIterator points into a Core's entities, iterating over them with optional
// type filter criteria.
type coreIterator struct {
	co  *Core
	i   int
	tcl TypeClause
}

// Next advances the iterator to point at the next matching entity, and
// returns true if such an entity was found; otherwise iteration is done, and
// false is returned.
func (it *coreIterator) Next() bool {
	for it.i++; it.i < len(it.co.types); it.i++ {
		t := it.co.types[it.i]
		if it.tcl.test(t) {
			return true
		}
	}
	return false
}

// Reset resets the iterator, causing it to start over.
func (it *coreIterator) Reset() { it.i = -1 }

// Count counts how many entities remain to be iterated, without advancing the
// iterator.
func (it coreIterator) Count() int {
	n := 0
	for it.Next() {
		n++
	}
	return n
}

// Type returns the type of the current entity, or NoType if iteration is
// done.
func (it coreIterator) Type() ComponentType {
	if it.i < len(it.co.types) {
		return it.co.types[it.i]
	}
	return NoType
}

// ID returns the type of the current entity, or 0 if iteration is done.
func (it coreIterator) ID() EntityID {
	if it.i < len(it.co.types) {
		return EntityID(it.i + 1)
	}
	return 0
}

// Entity returns a reference to the current entity, or NilEntity if
// iteration is done.
func (it coreIterator) Entity() Entity {
	if it.i < len(it.co.types) {
		return Entity{it.co, EntityID(it.i + 1)}
	}
	return NilEntity
}
