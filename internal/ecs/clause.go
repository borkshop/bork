package ecs

import "fmt"

// TypeClause is a logical filter for ComponentTypes.  If All is non-0, then
// Test()s true only for types that have all of those type bits set.
// Similarly if Any non-0, then Test()s true only for types that have at least
// one of those type bits set.
type TypeClause struct {
	All ComponentType
	Any ComponentType
}

func (tcl TypeClause) String() string {
	if tcl.All == 0 {
		return fmt.Sprintf("Any(%v)", tcl.Any)
	}
	if tcl.Any == 0 {
		return fmt.Sprintf("All(%v)", tcl.All)
	}
	return fmt.Sprintf("Clause(%v, %v)", tcl.All, tcl.Any)
}

// Test returns true/or false based on above logic description.
func (tcl TypeClause) Test(t ComponentType) bool {
	if tcl.All != 0 && !t.All(tcl.All) {
		return false
	}
	if tcl.Any != 0 && !t.Any(tcl.Any) {
		return false
	}
	return true
}

// AllClause matches any type; always Test()s true.
var AllClause = TypeClause{}

// Clause is a convenience constructor.
func Clause(all, any ComponentType) TypeClause { return TypeClause{all, any} }

// All is a convenience constructor.
func All(t ComponentType) TypeClause { return TypeClause{All: t} }

// Any is a convenience constructor.
func Any(t ComponentType) TypeClause { return TypeClause{Any: t} }

// Filter a list of entities under a given type clause.
func Filter(ents []Entity, tcl TypeClause) []Entity {
	i, j := 0, 0
	for ; j < len(ents); j++ {
		if tcl.Test(ents[j].Type()) {
			if j > i {
				ents[i] = ents[j]
			}
			i++
		}
	}
	for j = i; j < len(ents); j++ {
		ents[j] = NilEntity
	}
	return ents[:i]
}

// TODO: boolean logic methods?
