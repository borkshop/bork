package eps

import (
	"image"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/point"
)

// Moves is a movement system around an EPS.
//
// Movement works by registering a pending movement intent, expressed as a
// direction and magnitude. Pending moves are then processed, updating entity
// positions (modulo collisions).
//
// After processing, collision relations are available to build further
// mechanics upon. Collision relations carry the moved direction, and any
// remaining magnitude not spent on the move.
//
// Collisions only occur if more than one entity share any bits under the type
// mask given to Init().
//
// Pending moves are expressed as a self-relation (a == b == ent), while
// collisions have a == movingEntity and b == hitEntity.
type Moves struct {
	// PreCheck is a hook that may modify a pending move just before
	// application; e.g. to discount move magnitude due to inability of the
	// entity. The function may also return a limit that will cap how much mag
	// may be spent on the move; any remaining mag passes on in the subsequent
	// collision or pending relation.
	PreCheck func(uc *ecs.UpsertCursor, dir image.Point, mag int) (_ image.Point, _, limit int)

	eps      *EPS
	collMask ecs.ComponentType

	ecs.Relation
	dir []image.Point
	mag []int
}

const (
	movDir ecs.ComponentType = 1 << iota
	movMag

	movRelPending
	movRelCollide

	// MaxMoveTypeBit is where extension types may pick up within Moves's
	// ComponentType space. Use it like:
	// 	const (
	// 		myMoveType ecs.ComponentType = 1 << iota + eps.MaxMoveTypeBit
	// 		myOtherMoveType // etc...
	// 	)
	MaxMoveTypeBit = iota
)

// Init ialize the movement system, attached to the given positioning system,
// and using the given bits collision mask.
func (mov *Moves) Init(eps *EPS, collMask ecs.ComponentType) {
	mov.eps = eps
	mov.collMask = collMask

	mov.Relation.Init(mov.eps.core, 0, mov.eps.core, 0)
	mov.dir = []image.Point{image.ZP}
	mov.mag = []int{0}
	mov.Core.RegisterAllocator(movDir|movMag, mov.alloc)
	mov.Core.RegisterDestroyer(movDir, mov.destroyDir)
	mov.Core.RegisterDestroyer(movMag, mov.destroyMag)
}

func (mov *Moves) alloc(id ecs.EntityID, _ ecs.ComponentType) {
	mov.dir = append(mov.dir, image.ZP)
	mov.mag = append(mov.mag, 0)
}

func (mov *Moves) destroyDir(id ecs.EntityID, _ ecs.ComponentType) { mov.dir[id] = image.ZP }
func (mov *Moves) destroyMag(id ecs.EntityID, _ ecs.ComponentType) { mov.mag[id] = 0 }

func (mov *Moves) pendingCur(ent ecs.Entity) ecs.Cursor {
	return mov.Select(movRelPending.All(),
		ecs.InA(ent.ID()), ecs.InB(ent.ID()))
}

// Mag returns any magnitude associated with the given move relation; 0 means
// no magnitude defined.
func (mov *Moves) Mag(move ecs.Entity) int {
	if move.Type().HasAll(movMag) {
		return mov.mag[move.ID()]
	}
	return 0
}

// Dir returns any direction associated with the given move relation; the bool
// return is true only if the move has a defined direction.
func (mov *Moves) Dir(move ecs.Entity) (image.Point, bool) {
	if move.Type().HasAll(movDir) {
		return mov.dir[move.ID()], true
	}
	return image.ZP, false
}

// SetMag sets the magnitude associated with the given move relation; 0 removes
// the magnitude component (and destroys the move if it is pending).
func (mov *Moves) SetMag(move ecs.Entity, mag int) {
	if mag == 0 {
		move.Delete(movMag)
		if move.Type().HasAll(movRelPending) {
			move.Destroy()
		}
	} else {
		move.Add(movMag)
		mov.mag[move.ID()] = mag
	}
}

// SetDir sets the direction associated with the given move relation.
func (mov *Moves) SetDir(move ecs.Entity, dir image.Point) {
	move.Add(movDir)
	mov.dir[move.ID()] = dir
}

// DeleteDir deletes any direction associated with the given move relation,
// destroying it if it's now reduced to a magnitude-less pending move.
func (mov *Moves) DeleteDir(move ecs.Entity) {
	move.Delete(movDir)
	if t := move.Type(); t.HasAll(movRelPending) && !t.HasAny(movMag) {
		move.Destroy()
	}
}

// GetPendingMove returns the associated pending move relation for the given
// entity.
func (mov *Moves) GetPendingMove(ent ecs.Entity) ecs.Entity {
	if cur := mov.pendingCur(ent); cur.Scan() {
		return cur.R()
	}
	return ecs.NilEntity
}

// AddPendingMove adds the given magnitude to any pending move for the given
// entity, overwriting the direction. Ensures that at most one pending move is
// definded for the given entity.
func (mov *Moves) AddPendingMove(ent ecs.Entity, dir image.Point, mag, maxMag int) ecs.Entity {
	// TODO better support upsert reduction
	accum := ecs.NilEntity
	mov.Upsert(mov.pendingCur(ent), func(uc *ecs.UpsertCursor) {
		if move := uc.R(); move.Type().HasAll(movMag) {
			mag += mov.mag[move.ID()]
			if maxMag > 0 && mag > maxMag {
				mag = maxMag
			}
		}
		if accum == ecs.NilEntity {
			accum = uc.Emit(movRelPending|movDir, ent, ent)
			mov.dir[accum.ID()] = dir
		}
		mov.SetMag(accum, mag)
	})
	return accum
}

// SetPendingMove sets both magnitude and direction on an existing or new
// pending move for the given entity. Ensures that at most one pending move is
// defined for the given entity.
func (mov *Moves) SetPendingMove(ent ecs.Entity, dir image.Point, mag int) ecs.Entity {
	// TODO better support upsert reduction
	accum := ecs.NilEntity
	mov.Upsert(mov.pendingCur(ent), func(uc *ecs.UpsertCursor) {
		if accum != ecs.NilEntity {
			return
		}
		accum = uc.Emit(movRelPending|movDir, ent, ent)
		mov.dir[accum.ID()] = dir
		if mag > 0 {
			accum.Add(movMag)
			mov.mag[accum.ID()] = mag
		}
	})
	return accum
}

// Process applies pending moves, generating any consequesnt collisons; any
// prior collisions are first deleted.
func (mov *Moves) Process() {
	mov.Upsert(mov.Select(movRelCollide.All()), nil)
	// TODO 2-phase application so that mid-line and glancing collisions are possible
	mov.Upsert(mov.Select((movRelPending | movDir | movMag).All()), mov.processPendingMove)
}

func (mov *Moves) processPendingMove(uc *ecs.UpsertCursor) {
	move := uc.R()
	if move == ecs.NilEntity {
		return
	}

	dir, mag := mov.dir[move.ID()], mov.mag[move.ID()]
	limit := mag
	if mov.PreCheck != nil {
		dir, mag, limit = mov.PreCheck(uc, dir, mag)
		if mag == 0 {
			return
		}
	}

	t, ent := move.Type(), uc.A()
	if !dir.Eq(image.ZP) {
		var hit ecs.Entity
		hit, mag = mov.runMove(ent, point.Sign(dir), mag, limit)
		if hit != ecs.NilEntity {
			t = t & ^movRelPending | movRelCollide
		} else if mag <= 0 {
			return
		} else { // hit == ecs.NilEntity
			hit = ent
		}
		move = uc.Emit(t, ent, hit)
	} else {
		move = uc.Emit(t, ent, ent)
	}
	mov.SetMag(move, mag)
}

func (mov *Moves) runMove(
	ent ecs.Entity,
	unit image.Point, mag, limit int,
) (ecs.Entity, int) {
	pos, _ := mov.eps.Get(ent)
	for limit > 0 && mag > 0 {
		limit--
		mag--
		new := pos.Add(unit)
		if atc := ent.Type() & mov.collMask; atc != 0 {
			for _, b := range mov.eps.At(new) {
				if b.Type().HasAny(atc) {
					mov.eps.Set(ent, pos)
					return b, mag
				}
			}
		}
		pos = new
	}
	mov.eps.Set(ent, pos)
	return ecs.NilEntity, mag
}

// Pending returns a relation cursor over all pending moves; thse are either
// unprocessed moves, or leftover/unused magnitudes.
func (mov *Moves) Pending(opts ...ecs.CursorOpt) ecs.Cursor {
	sopts := make([]ecs.CursorOpt, 1, 1+len(opts))
	sopts[0] = movRelPending.All()
	sopts = append(sopts, opts...)
	return mov.Select(sopts...)
}

// Collisions returns a relation cursor over all collisions from the last
// processing round.
func (mov *Moves) Collisions(opts ...ecs.CursorOpt) ecs.Cursor {
	sopts := make([]ecs.CursorOpt, 1, 1+len(opts))
	sopts[0] = movRelCollide.All()
	sopts = append(sopts, opts...)
	return mov.Select(sopts...)
}
