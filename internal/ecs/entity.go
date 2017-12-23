package ecs

import "fmt"

// Entity is a reference to an entity in a Core
type Entity struct {
	co *Core
	id EntityID
}

// NilEntity is the zero of Entity, representing "no entity, in no Core".
var NilEntity = Entity{}

func (ent Entity) String() string {
	if ent.co == nil {
		return fmt.Sprintf("Nil<>[%v]", ent.id)
	}
	if ent.co == nil {
		return fmt.Sprintf("%p<>[%v]", ent.co, ent.id)
	}
	return fmt.Sprintf("%p<%v>[%v]",
		ent.co,
		ent.co.types[ent.id-1],
		ent.id,
	)
}

// Type returns the type of the referenced entity, or NoType if the reference
// is empty.
func (ent Entity) Type() ComponentType {
	if ent.co == nil || ent.id == 0 {
		return NoType
	}
	return ent.co.types[ent.id-1]
}

// ID returns the ID of the referenced entity; it SHOULD only be called in a
// context where the caller is sure of ownership; when in doubt, use
// Core.Deref(ent) instead.
func (ent Entity) ID() EntityID {
	if ent.co == nil {
		return 0
	}
	return ent.id
}

// Deref unpacks an Entity reference, returning its ID; it panics if the Core
// doesn't own the Entity.
func (co *Core) Deref(e Entity) EntityID {
	if e.co == co {
		return e.id
	} else if e.co == nil {
		panic("nil entity")
	} else {
		panic("foreign entity")
	}
}

// Ref returns an Entity reference to the given ID; it is valid to return a
// reference to the zero entity, to represent "no entity, in this Core" (e.g.
// will Deref() to 0 EntityID).
func (co *Core) Ref(id EntityID) Entity {
	if id == 0 {
		return NilEntity
	}
	return Entity{co, id}
}

// AddEntity adds an entity to a core, returning an Entity reference; it MAY
// re-use a previously-used but since-destroyed entity (one whose type is still
// NoType). MAY invokes all allocators to make space for more entities (will do
// so if Cap() == Len()).
func (co *Core) AddEntity(nt ComponentType) Entity {
	ent := Entity{co, co.allocate()}
	co.SetType(ent.id, nt)
	return ent
}

// Add sets bits in the entity's type, calling any creators that are newly
// satisfied by the new type.
func (ent Entity) Add(t ComponentType) {
	if ent.co != nil && ent.id > 0 {
		old := ent.co.types[ent.id-1]
		ent.co.SetType(ent.id, old|t)
	}
}

// Delete clears bits in the entity's type, calling any destroyers that are no
// longer satisfied by the new type (which may be NoType).
func (ent Entity) Delete(t ComponentType) {
	if ent.co != nil && ent.id > 0 {
		old := ent.co.types[ent.id-1]
		ent.co.SetType(ent.id, old & ^t)
	}
}

// Destroy sets the entity's type to NoType, invoking any destroyers that match
// the prior type.`
func (ent Entity) Destroy() {
	if ent.co != nil && ent.id > 0 {
		ent.co.SetType(ent.id, NoType)
	}
}

// SetType sets the entity's type; may invoke creators and destroyers as
// appropriate.
func (ent Entity) SetType(t ComponentType) {
	if ent.co != nil && ent.id > 0 {
		ent.co.SetType(ent.id, t)
	}
}

func (co *Core) allocate() EntityID {
	if co.free > 0 {
		for i := 0; i < len(co.types); i++ {
			if co.types[i] == NoType {
				co.free--
				return EntityID(i + 1)
			}
		}
	}
	id := EntityID(len(co.types) + 1)
	co.types = append(co.types, NoType)
	for _, ef := range co.allocators {
		ef.f(id, NoType)
	}
	return id
}

// Type returns the entity's type.
func (co *Core) Type(id EntityID) ComponentType { return co.types[id-1] }

// SetType changes an entity's type, calling any relevant lifecycle functions.
func (co *Core) SetType(id EntityID, new ComponentType) {
	i := id - 1
	old := co.types[i]
	if old == new {
		return
	}
	co.types[i] = new
	if old == NoType {
		for _, ef := range co.creators {
			if ef.t == NoType {
				ef.f(id, new)
				new = co.types[i]
			}
		}
	}
	if new & ^old != 0 {
		for _, ef := range co.creators {
			if new.HasAll(ef.t) && !old.HasAll(ef.t) {
				ef.f(id, new)
				new = co.types[i]
			}
		}
	}
	if old & ^new != 0 {
		for _, ef := range co.destroyers {
			if old.HasAll(ef.t) && !new.HasAll(ef.t) {
				ef.f(id, new)
				new = co.types[i]
			}
		}
	}
	if new == NoType {
		for _, ef := range co.destroyers {
			if ef.t == NoType {
				ef.f(id, new)
				new = co.types[i]
			}
		}
		co.free++
	}
}
