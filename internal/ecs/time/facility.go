package ecsTime

import (
	"math"

	"github.com/borkshop/bork/internal/ecs"
)

// Facility implements a timer facility attached to an ecs.Core.
type Facility struct {
	// NOTE: dense storage strategy, may want to explicate that
	// eventually and provide alternate
	core *ecs.Core
	t    ecs.ComponentType

	now Time

	// TODO: shift to a heap rather than iterating and decrementing every time;
	// maybe introspectively after len(timers) passes a certain point?
	timers []timer
	tocall []cb
}

type cb struct {
	f func(ecs.Entity)
	e ecs.Entity
}

type timer struct {
	t        ecs.ComponentType
	remain   Duration
	period   Duration
	callback func(ecs.Entity)
}

// Init sets up the timer facility, attached to the given ecs.Core, and using
// the supplied ComponentType to indicate "has a timer". The given
// ComponentType MUST NOT be registered by another allocator.
func (fac *Facility) Init(core *ecs.Core, t ecs.ComponentType) {
	if fac.core != nil {
		panic("Timers already initialized")
	}
	fac.core = core
	fac.t = t
	fac.timers = []timer{{}}
	fac.core.RegisterAllocator(fac.t, fac.alloc)
	fac.core.RegisterDestroyer(fac.t, fac.destroyTimer)
}

// Now returns the current time, which increments at
// the start of Process()ing.
func (fac Facility) Now() Time { return fac.now }

// After attaches a one-shot timer to the given entity that expires after d
// processing time has elapsed calling the given function with the attached
// entity.
//
// Any prior timer (one-shot or periodic) attached to the entity is overwritten.
//
// Panics if the entity does not belong to the Facility's core, or the duration
// is not positive.
func (fac *Facility) After(ent ecs.Entity, d Duration, callback func(ecs.Entity)) {
	if d <= 0 {
		panic("invalid timer duration")
	}
	id := fac.core.Deref(ent)
	ent.Add(fac.t)
	fac.timers[id] = timer{remain: d, callback: callback}
}

// Every attaches a periodic timer to the given entity that fires every d
// processing time has elapsed, calling the given function with the attached
// entity every time.
//
// Any prior timer (one-shot or periodic) attached to the entity is overwritten.
//
// Panics if the entity does not belong to the Facility's core, or the duration
// is not positive.
func (fac *Facility) Every(ent ecs.Entity, d Duration, callback func(ecs.Entity)) {
	if d <= 0 {
		panic("invalid timer duration")
	}
	id := fac.core.Deref(ent)
	ent.Add(fac.t)
	fac.timers[id] = timer{remain: d, period: d, callback: callback}
}

// Cancel deletes any timer (one-shot or periodic )attached to the given
// entity, returning true only if there was such a timer to delete.
func (fac *Facility) Cancel(ent ecs.Entity) bool {
	_ = fac.core.Deref(ent)
	if ent.Type().HasAll(fac.t) {
		ent.Delete(fac.t)
		return true
	}
	return false
}

// Process calls any timers whose time has come.
//
// Panics if The End Time has come (2^64 Process()ing ticks integer overflow).
//
// Callback functions are called (in an ARBITRARY order) in one batch AFTER all
// expired timers have been processed. Therefore callbacks may re-set a
// one-shot, or cancel a periodic (their own timer, or another).
func (fac *Facility) Process() {
	fac.now = fac.now.Add(1)
	if fac.now == math.MaxUint64 {
		panic("The End is Now!")
	}
	fac.tocall = fac.tocall[:0]
	for it := fac.core.Iter(fac.t.All()); it.Next(); {
		t := &fac.timers[it.ID()]
		if t.remain <= 0 {
			it.Entity().Delete(fac.t)
			continue
		}
		t.remain--
		if t.remain > 0 {
			continue
		}
		ent := it.Entity()
		fac.tocall = append(fac.tocall, cb{t.process(ent), ent})
	}
	for _, cb := range fac.tocall {
		cb.f(cb.e)
	}
}

func (t *timer) process(ent ecs.Entity) func(ecs.Entity) {
	callback := t.callback
	if t.period != 0 {
		t.remain = t.period // interval refresh
	} else {
		ent.Delete(t.t) // one shot
	}
	return callback
}

func (fac *Facility) alloc(id ecs.EntityID, t ecs.ComponentType) {
	fac.timers = append(fac.timers, timer{})
}
func (fac *Facility) destroyTimer(id ecs.EntityID, t ecs.ComponentType) { fac.timers[id] = timer{} }
