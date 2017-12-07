package ecs

import (
	"fmt"
	"strings"
)

// TypeClause is a logical filter for ComponentTypes.
type TypeClause interface {
	CursorOpt
	test(ComponentType) bool
	// TODO this is a convenient place to start, but to make it perform, we'll
	// need to compile to a tighter for linear scan and/or to proper planning
	// wrt available indexing.
}

var (
	// TrueClause matches any type.
	TrueClause TypeClause = constClause(true)

	// FalseClause matches no type.
	FalseClause TypeClause = constClause(false)
)

// All return a clause that matches only if all of the type bits are set.  If
// the type is NoType, the clause never matches (always returns false).
func (t ComponentType) All() TypeClause {
	if t == NoType {
		return FalseClause
	}
	return allClause(t)
}

// Any return a clause that matches only if at least one of the type bits is
// set. If the type is NoType, the clause always matches (always returns true).
func (t ComponentType) Any() TypeClause {
	if t == NoType {
		return TrueClause
	}
	return anyClause(t)
}

// NotAll return a clause that matches only if at least one of the type bits is not set.
func (t ComponentType) NotAll() TypeClause { return notAllClause(t) }

// NotAny return a clause that matches only if none of the type bits are not set.
func (t ComponentType) NotAny() TypeClause { return notAnyClause(t) }

type constClause bool
type allClause ComponentType
type anyClause ComponentType
type notAllClause ComponentType
type notAnyClause ComponentType
type andClause []TypeClause
type orClause []TypeClause

func (cc constClause) String() string { return fmt.Sprintf("%v", bool(cc)) }
func (t allClause) String() string    { return fmt.Sprintf("All(%v)", ComponentType(t)) }
func (t anyClause) String() string    { return fmt.Sprintf("Any(%v)", ComponentType(t)) }
func (t notAllClause) String() string { return fmt.Sprintf("NotAll(%v)", ComponentType(t)) }
func (t notAnyClause) String() string { return fmt.Sprintf("NotAny(%v)", ComponentType(t)) }
func (tcls andClause) String() string {
	ss := make([]string, len(tcls))
	for i := range tcls {
		ss = append(ss, fmt.Sprint(tcls[i]))
	}
	return fmt.Sprintf("And(%s)", strings.Join(ss, " "))
}
func (tcls orClause) String() string {
	ss := make([]string, len(tcls))
	for i := range tcls {
		ss = append(ss, fmt.Sprint(tcls[i]))
	}
	return fmt.Sprintf("Or(%s)", strings.Join(ss, " "))
}

func (cc constClause) test(ComponentType) bool    { return bool(cc) }
func (t allClause) test(ot ComponentType) bool    { return ot&ComponentType(t) == ComponentType(t) }
func (t anyClause) test(ot ComponentType) bool    { return ot&ComponentType(t) != 0 }
func (t notAllClause) test(ot ComponentType) bool { return ot&ComponentType(t) != ComponentType(t) }
func (t notAnyClause) test(ot ComponentType) bool { return ot&ComponentType(t) == 0 }
func (tcls andClause) test(ot ComponentType) bool {
	for i := range tcls {
		if !tcls[i].test(ot) {
			return false
		}
	}
	return true
}
func (tcls orClause) test(ot ComponentType) bool {
	for i := range tcls {
		if tcls[i].test(ot) {
			return true
		}
	}
	return false
}

// And returns a type clause that matches only if all of its component clauses
// match.
func And(tcls ...TypeClause) TypeClause {
	if len(tcls) == 0 {
		return FalseClause
	}
	for len(tcls) > 1 {
		tcl := and(tcls[0], tcls[1])
		if tcl == nil {
			tcl = andClause{tcls[0], tcls[1]}
		}
		tcls = tcls[1:]
		tcls[0] = tcl
	}
	return tcls[0]
}

// Or returns a type clause that matches only if any of its component clauses
// match.
func Or(tcls ...TypeClause) TypeClause {
	if len(tcls) == 0 {
		return TrueClause
	}
	for len(tcls) > 1 {
		tcl := or(tcls[0], tcls[1])
		if tcl == nil {
			tcl = orClause{tcls[0], tcls[1]}
		}
		tcls = tcls[1:]
		tcls[0] = tcl
	}
	return tcls[0]
}

func and(a, b TypeClause) TypeClause {
	if _, ok := b.(constClause); ok {
		return and(b, a)
	}
	if cc, ok := a.(constClause); ok {
		if bool(cc) {
			return b
		}
		return FalseClause
	}

	// all(a) && all(b) = all(a|b)
	if allA, ok := a.(allClause); ok {
		if allB, ok := b.(allClause); ok {
			return allClause(allA | allB)
		}
	}

	// notAny(a) && notAny(b) = notAny(a|b)
	if notAnyA, ok := a.(notAnyClause); ok {
		if notAnyB, ok := b.(notAnyClause); ok {
			return notAnyClause(notAnyA | notAnyB)
		}
	}

	if tclsA, ok := a.(andClause); ok {
		r := append(andClause(nil), tclsA...)
		tclsB, ok := b.(andClause)
		if !ok {
			tclsB = andClause{b}
		}
		for _, b := range tclsB {
			for i := range r {
				if tcl := and(r[i], b); tcl != nil {
					r[i] = tcl
					continue
				}
			}
			r = append(r, b)
		}
		return r
	}

	return nil
}

func or(a, b TypeClause) TypeClause {
	if _, ok := b.(constClause); ok {
		return or(b, a)
	}
	if cc, ok := a.(constClause); ok {
		if bool(cc) {
			return TrueClause
		}
		return b
	}

	// any(a) || any(b) = any(a&b)
	if anyA, ok := a.(anyClause); ok {
		if anyB, ok := b.(anyClause); ok {
			return anyClause(anyA & anyB)
		}
	}

	// notAll(a) || notAll(b) = notAll(a&b)
	if notAllA, ok := a.(notAllClause); ok {
		if notAllB, ok := b.(notAllClause); ok {
			return notAllClause(notAllA & notAllB)
		}
	}

	if tclsA, ok := a.(orClause); ok {
		r := append(orClause(nil), tclsA...)
		tclsB, ok := b.(orClause)
		if !ok {
			tclsB = orClause{b}
		}
		for _, b := range tclsB {
			for i := range r {
				if tcl := or(r[i], b); tcl != nil {
					r[i] = tcl
					continue
				}
			}
			r = append(r, b)
		}
		return r
	}

	return nil
}

// Not returns a type clause that matches only the given clause does not match.
func Not(tcl TypeClause) TypeClause {
	switch val := tcl.(type) {
	case constClause:
		if bool(val) {
			return FalseClause
		}
		return TrueClause
	case allClause:
		return notAllClause(val)
	case anyClause:
		return notAnyClause(val)
	case notAllClause:
		return allClause(val)
	case notAnyClause:
		return anyClause(val)
	case andClause:
		nor := make(orClause, len(val))
		for i := range val {
			nor[i] = Not(val[i])
		}
		return nor
	case orClause:
		nand := make(andClause, len(val))
		for i := range val {
			nand[i] = Not(val[i])
		}
		return nand
	default:
		panic(fmt.Sprintf("unsupported TypeClause %T", tcl))
	}
}
