package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"strings"

	termbox "github.com/nsf/termbox-go"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/ecs/eps"
	"github.com/borkshop/bork/internal/ecs/time"
	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/perf"
	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud/prompt"
)

// TODO: spirit possession

const (
	wcName ecs.ComponentType = 1 << iota
	wcTimer
	wcPosition
	wcCollide
	wcSolid
	wcGlyph
	wcBG
	wcFG
	wcInput
	wcWaiting
	wcBody
	wcSoul
	wcItem
	wcAI
	wcFloor
	wcWall
	wcSpawn
	wcAnt
)

const (
	renderMask    = wcPosition | wcGlyph
	playMoveMask  = wcPosition | wcInput | wcSoul
	charMask      = wcName | wcGlyph | wcBody | wcSolid
	collMask      = wcPosition | wcCollide
	combatMask    = wcCollide | wcBody
	floorTileMask = wcPosition | wcBG | wcFloor
)

type worldItem interface {
	interact(pr prompt.Prompt, w *world, item, ent ecs.Entity) (prompt.Prompt, bool)
}

type durableItem interface {
	worldItem
	HPRange() (int, int)
}

type destroyableItem interface {
	worldItem
	destroy(w *world)
}

type world struct {
	perf perf.Perf
	ui

	logFile *os.File
	logger  *log.Logger
	rng     *rand.Rand

	over         bool
	enemyCounter int

	ecs.System
	pos    eps.EPS
	timers time.Facility

	Names  []string
	Glyphs []rune
	BG     []termbox.Attribute
	FG     []termbox.Attribute
	bodies []*body
	items  []worldItem

	moves   moves // TODO: maybe subsume into pos?
	waiting ecs.Iterator
}

type moves struct {
	ecs.Relation
	timers time.Facility
	n      []int
	p      []point.Point
}

const (
	movN ecs.ComponentType = 1 << iota
	movP
	movT

	mrCollide
	mrHit
	mrItem
	mrGoal
	mrAgro
	mrPending
	mrMoveRange
	mrRest

	movCharge  = ecs.ComponentType(mrPending) | movN
	movPending = ecs.ComponentType(mrPending) | movN | movP
	movResting = ecs.ComponentType(mrRest) | movN
)

func (mov *moves) init(core *ecs.Core) {
	mov.Relation.Init(core, 0, core, 0)
	mov.timers.Init(&mov.Core, movT)
	mov.n = []int{0}
	mov.p = []point.Point{point.Zero}
	mov.RegisterAllocator(movN|movP, mov.allocData)
	mov.RegisterDestroyer(movN, mov.deleteN)
	mov.RegisterDestroyer(movP, mov.deleteP)
}

func (mov *moves) allocData(id ecs.EntityID, t ecs.ComponentType) {
	mov.n = append(mov.n, 0)
	mov.p = append(mov.p, point.Zero)
}

func (mov *moves) deleteN(id ecs.EntityID, t ecs.ComponentType) { mov.n[id] = 0 }
func (mov *moves) deleteP(id ecs.EntityID, t ecs.ComponentType) { mov.p[id] = point.Zero }

func newWorld(v *view.View) (*world, error) {
	// f, err := os.Create(fmt.Sprintf("%v.log", time.Now().Format(time.RFC3339)))
	// if err != nil {
	// 	return nil, err
	// }
	w := &world{
		rng: rand.New(rand.NewSource(rand.Int63())),
		// logger: log.New(f, "", 0),
	}
	w.init(v)
	// w.log("logging to %q", f.Name())
	return w, nil
}

func (w *world) Process() {
	w.perf.Process()
}

func (w *world) init(v *view.View) {
	w.perf.Init("", &w.System)

	w.ui.init(v, &w.perf)
	w.timers.Init(&w.Core, wcTimer)

	w.Procs = append(w.Procs,
		ecs.ProcFunc(func() {
			w.moves.Upsert(w.moves.Select(mrCollide.Any()), nil)
		}),
		&w.timers,
		&w.moves.timers,
		ecs.ProcFunc(w.generateAIMoves), // give AI a chance!
		ecs.ProcFunc(w.applyMoves),      // resolve moves
		ecs.ProcFunc(w.processAIItems),  // nom nom
		ecs.ProcFunc(w.processCombat),   // e.g. deal damage
		ecs.ProcFunc(w.processRest),     // healing etc
		ecs.ProcFunc(w.checkOver),       // no souls => done
		ecs.ProcFunc(w.maybeSpawn),      // spawn more demons
	)

	// TODO: consider eliminating the padding for EntityID(0)
	w.Names = []string{""}
	w.Glyphs = []rune{0}
	w.BG = []termbox.Attribute{0}
	w.FG = []termbox.Attribute{0}
	w.bodies = []*body{nil}
	w.items = []worldItem{nil}

	w.RegisterAllocator(wcName|wcGlyph|wcBG|wcFG|wcBody|wcItem, w.allocWorld)
	w.RegisterCreator(wcInput, w.createInput)
	w.RegisterCreator(wcBody, w.createBody)
	w.RegisterDestroyer(wcBody, w.destroyBody)
	w.RegisterDestroyer(wcItem, w.destroyItem)
	w.RegisterDestroyer(wcInput, w.destroyInput)

	w.pos.Init(&w.Core, wcPosition)
	w.moves.init(&w.Core) // TODO: maybe subsume into pos?
	w.waiting = w.Iter((charMask | wcWaiting).All())
}

var movementRangeLabels = []string{"Walk", "Lunge"}

func newRangeChooser(w *world, ent ecs.Entity) *rangeChooser {
	return &rangeChooser{
		w:   w,
		ent: ent,
	}
}

type rangeChooser struct {
	w   *world
	ent ecs.Entity
}

func (rc *rangeChooser) label() string {
	n := rc.w.getMovementRange(rc.ent)
	return movementRangeLabels[n-1]
}

func (rc *rangeChooser) RunPrompt(prior prompt.Prompt) (next prompt.Prompt, required bool) {
	next = prior.Sub("Set Movement Range")
	for i, label := range movementRangeLabels {
		n := i + 1
		r := '0' + rune(n)
		run := prompt.Func(func(prior prompt.Prompt) (next prompt.Prompt, required bool) {
			return rc.chosen(prior, n)
		})
		if n == 1 {
			next.AddAction(r, run, "%s (%d cell, consumes %d charge)", label, n, n)
		} else {
			next.AddAction(r, run, "%s (%d cells, consumes %d charges)", label, n, n)
		}
	}
	return next, true
}

func (rc *rangeChooser) chosen(prior prompt.Prompt, n int) (next prompt.Prompt, required bool) {
	rc.w.setMovementRange(rc.ent, n)
	next = prior.Unwind()
	next.SetActionMess(0, rc, movementRangeLabels[n-1])
	return next, false
}

func (w *world) allocWorld(id ecs.EntityID, t ecs.ComponentType) {
	w.Names = append(w.Names, "")
	w.Glyphs = append(w.Glyphs, 0)
	w.BG = append(w.BG, 0)
	w.FG = append(w.FG, 0)
	w.bodies = append(w.bodies, nil)
	w.items = append(w.items, nil)
}

func (w *world) createInput(id ecs.EntityID, t ecs.ComponentType) {
	w.setMovementRange(w.Ref(id), 1)
}

func (w *world) createBody(id ecs.EntityID, t ecs.ComponentType) {
	w.bodies[id] = newBody()
}

func (w *world) destroyBody(id ecs.EntityID, t ecs.ComponentType) {
	if bo := w.bodies[id]; bo != nil {
		bo.Clear()
	}
}

func (w *world) destroyItem(id ecs.EntityID, t ecs.ComponentType) {
	item := w.items[id]
	w.items[id] = nil
	if des, ok := item.(destroyableItem); ok {
		des.destroy(w)
	}
}

func (w *world) destroyInput(id ecs.EntityID, t ecs.ComponentType) {
	if name := w.Names[id]; name != "" {
		// TODO: restore attribution
		// w.log("%s destroyed by %s", w.getName(targ, "?!?"), w.getName(src, "!?!"))
		w.log("%s has been destroyed", name)
	}
}

func (w *world) extent() point.Box {
	var bbox point.Box
	for it := w.Iter(renderMask.All()); it.Next(); {
		pos, _ := w.pos.Get(it.Entity())
		bbox = bbox.ExpandTo(point.Point(pos))
	}
	return bbox
}

func (w *world) Close() error { return nil }

const maxChargeFromResting = 4

func (w *world) addPendingMove(ent ecs.Entity, move point.Point) {
	if !ent.Type().HasAll(wcInput) {
		return // who asked you
	}
	// TODO better support upsert reduction
	accum := ecs.NilEntity
	w.moves.Upsert(
		w.moves.Select(mrPending.All(), ecs.InA(ent.ID()), ecs.InB(ent.ID())),
		func(uc *ecs.UpsertCursor) {
			if accum != ecs.NilEntity {
				if rel := uc.R(); rel.Type().HasAll(movN) {
					w.moves.n[accum.ID()] += w.moves.n[rel.ID()]
				}
				return
			}
			id := uc.Emit(mrPending|movP|movN, ent, ent).ID()
			w.moves.p[id] = w.moves.p[id].Add(move)
			if n := w.moves.n[id]; n < maxChargeFromResting {
				w.moves.n[id] = n + 1
			}
		})
}

const maxMovementRange = 2

func (w *world) setMovementRange(a ecs.Entity, n int) {
	_ = w.Deref(a)
	if n > maxMovementRange {
		n = maxMovementRange
	}
	// TODO better support upsert reduction
	found := false
	w.moves.Upsert(
		w.moves.Select(mrMoveRange.All(), ecs.InA(a.ID()), ecs.InB(a.ID())),
		func(uc *ecs.UpsertCursor) {
			if !found {
				movRange := uc.Emit(mrMoveRange|movN, a, a)
				w.moves.n[movRange.ID()] = n
				found = true
			}
		})
}

func (w *world) getMovementRange(a ecs.Entity) int {
	id := w.Deref(a)
	for cur := w.moves.Select((mrMoveRange | movN).All(), ecs.InA(id)); cur.Scan(); {
		if n := w.moves.n[cur.R().ID()]; n <= maxMovementRange {
			return n
		}
	}
	return maxMovementRange
}

func (w *world) getCharge(ent ecs.Entity) (charge int) {
	for cur := w.moves.Select(movCharge.All(), ecs.InA(w.Deref(ent))); cur.Scan(); {
		charge += w.moves.n[cur.R().ID()]
	}
	return charge
}

func (w *world) processRest() {
	w.moves.Upsert(w.moves.Select(movResting.All()), func(uc *ecs.UpsertCursor) {
		if uc.R() == ecs.NilEntity {
			return
		}
		n := w.moves.n[uc.R().ID()]
		if uc.A().Type().HasAll(wcBody) && n == maxChargeFromResting {
			n = w.heal(uc.A(), n)
		}
		if n > 0 {
			rel := uc.Emit(mrPending|movN, uc.A(), uc.B())
			w.moves.n[rel.ID()] = n
		}
	})
}

func (w *world) heal(ent ecs.Entity, n int) int {
	sum, sel := 0, ecs.NilEntity
	bo := w.bodies[ent.ID()]
	for it := bo.Iter(bcHPart.All()); it.Next(); {
		id := it.ID()
		if dmg := bo.maxHP[id] - bo.hp[id]; dmg > 0 {
			sum += dmg
			if w.rng.Intn(sum) < dmg {
				sel = it.Entity()
			}
		}
	}
	if sel != ecs.NilEntity {
		id := sel.ID()
		heal := bo.maxHP[id] - bo.hp[id]
		if heal > n {
			heal = n
		}
		bo.hp[id] += heal
		n -= heal
		w.log("%s's %s healed for %v HP", w.getName(ent, "???"), bo.DescribePart(sel), heal)
	}
	return n
}

func (w *world) applyMoves() {
	// TODO: better resolution strategy based on connected components
	w.moves.Upsert(w.moves.Select(movPending.All()), func(uc *ecs.UpsertCursor) {
		move := uc.R()
		if move == ecs.NilEntity {
			return
		}

		a, b := uc.A(), uc.B()

		// can we actually affect a move?
		pend, n := w.moves.p[move.ID()], w.moves.n[move.ID()]
		if a.Type().HasAll(wcBody) {
			rating := w.bodies[a.ID()].movementRating()
			pend = pend.Mul(int(moremath.Round(rating * float64(n))))
			if pend.SumSQ() == 0 {
				if n > 1 {
					move = uc.Emit(mrRest, a, b)
					move.Add(movN)
					move.Delete(movP)
					w.moves.n[move.ID()] = n
				} else {
					uc.Emit(move.Type(), a, b)
				}
				return
			}
		}

		// move until we collide or exceeding lunging distance
		limit := w.getMovementRange(a)
		pos, _ := w.pos.Get(a)
		i := 0
		for ; i < n && i < limit; i++ {
			new := pos.Add(image.Point(pend.Sign()))
			var hit []ecs.Entity
			if a.Type().HasAll(wcCollide) {
				hit = w.pos.At(new)
			}
			for _, b := range hit {
				if b.Type().HasAll(wcCollide | wcSolid) {
					hitRel := uc.Emit(mrCollide|mrHit, a, b)
					if m := n - i; m > 1 {
						hitRel.Add(movN)
						w.moves.n[hitRel.ID()] = m
					}
					w.pos.Set(a, pos)
					return
				}
			}
			pos = new
		}
		n -= i
		w.pos.Set(a, pos)

		// moved without hitting anything
		move = uc.Emit(move.Type(), a, b)
		w.moves.p[move.ID()], w.moves.n[move.ID()] = point.Zero, n
	})
}

func (bo *body) movementRating() float64 {
	ratings := make(map[ecs.EntityID]float64, 6)
	for it := bo.Iter(bcLegPart.Any()); it.Next(); {
		rating := 1.0
		for bo.coGT.Init(it.ID()); bo.coGT.Traverse(); {
			id := bo.coGT.Node().ID()
			delete(ratings, id)
			rating *= float64(bo.hp[id]) / float64(bo.maxHP[id])
		}
		if it.Type().HasAll(bcCalf) {
			rating *= 2 / 3
		} else if it.Type().HasAll(bcThigh) {
			rating *= 1 / 3
		}
		ratings[it.ID()] = rating
	}
	rating := 0.0
	for _, r := range ratings {
		rating += r
	}
	return rating / 2.0
}

func (w *world) processCombat() {
	// TODO: make this an upsert that transmutes hits into damage/kill relations
	for cur := w.moves.Select(
		(mrCollide | mrHit).All(),
		ecs.Filter(func(cur ecs.Cursor) bool {
			return cur.A().Type().HasAll(combatMask) && cur.B().Type().HasAll(combatMask)
		}),
	); cur.Scan(); {
		src, targ := cur.A(), cur.B()

		aPart, bPart := w.checkAttackHit(src, targ)
		if aPart == ecs.NilEntity || bPart == ecs.NilEntity {
			continue
		}

		srcBo, targBo := w.bodies[src.ID()], w.bodies[targ.ID()]
		rating := srcBo.partHPRating(aPart) / targBo.partHPRating(bPart)
		if cur.R().Type().HasAll(movN) {
			mult := w.moves.n[cur.R().ID()]
			rating *= float64(mult)
		}
		rand := (1 + w.rng.Float64()) / 2 // like an x/2 + 1D(x/2) XXX reconsider
		dmg := int(moremath.Round(float64(srcBo.dmg[aPart.ID()]) * rating * rand))
		dmg -= targBo.armor[bPart.ID()]
		if dmg == 0 {
			if soulInvolved(src, targ) {
				w.log("%s's %s bounces off %s's %s",
					w.getName(src, "!?!"), srcBo.DescribePart(aPart),
					w.getName(targ, "?!?"), targBo.DescribePart(bPart))
			}
			continue
		}

		if dmg < 0 {
			w.dealAttackDamage(targ, bPart, src, aPart, -dmg)
		} else {
			w.dealAttackDamage(src, aPart, targ, bPart, dmg)
		}
	}
}

func (w *world) findPlayer() ecs.Entity {
	if it := w.Iter(playMoveMask.All()); it.Next() {
		return it.Entity()
	}
	return ecs.NilEntity
}

func (w *world) firstSoulBody() ecs.Entity {
	if it := w.Iter((wcBody | wcSoul).All()); it.Next() {
		return it.Entity()
	}
	return ecs.NilEntity
}

func (w *world) checkOver() {
	// count remaining souls
	if w.Iter(wcSoul.All()).Count() == 0 {
		w.log("game over")
		w.over = true
	}
}

func (w *world) maybeSpawn() {
	spawnPoints := w.Iter(wcSpawn.All())
	if spawnPoints.Count() == 0 {
		return
	}

	totalAgro := 0
	for cur := w.moves.Select(mrAgro.All()); cur.Scan(); {
		totalAgro += w.moves.n[cur.R().ID()]
	}

	totalHP, totalDmg, combatCount := 0, 0, 0
	for it := w.Iter(ecs.And((combatMask | wcInput).All(), wcWaiting.NotAll())); it.Next(); {
		combatCount++
		bo := w.bodies[it.ID()]
		if it.Type().HasAll(wcSoul) {
			hp, maxHP := bo.HPRange()
			dmg := maxHP - hp
			totalHP += maxHP
			totalDmg += dmg
		} else {
			totalHP += bo.HP()
		}
	}

	sum := totalHP + totalDmg
	if combatCount > 0 {
		sum -= totalHP/combatCount - totalAgro
	}
	if sum < 0 {
		sum = 0
	}

	w.waiting.Reset()
eachSpawnPoint:
	for spawnPoints.Next() {
		pos, _ := w.pos.Get(spawnPoints.Entity())
		for _, ent := range w.pos.At(pos) {
			if ent.Type().HasAll(collMask) {
				continue eachSpawnPoint
			}
		}
		ent := w.nextWaiting()
		hp := w.bodies[ent.ID()].HP()
		sum += hp
		if w.rng.Intn(sum) < hp {
			ent.Delete(wcWaiting)
			ent.Add(wcCollide | wcInput)
			w.pos.Set(ent, pos)
			w.addFrustration(ent, hp)
			ent = w.nextWaiting()
		}
	}
}

func (w *world) nextWaiting() ecs.Entity {
	if w.waiting.Next() {
		return w.waiting.Entity()
	}
	w.enemyCounter++
	return w.newChar(fmt.Sprintf("enemy%d", w.enemyCounter), 'X', wcAI)
}

func soulInvolved(a, b ecs.Entity) bool {
	// TODO: ability to say "those are soul remains"
	return a.Type().HasAll(wcSoul) || b.Type().HasAll(wcSoul)
}

func (w *world) dealAttackDamage(
	src, aPart ecs.Entity,
	targ, bPart ecs.Entity,
	dmg int,
) (leftover int, destroyed bool) {
	// TODO: store damage and kill relations

	srcBo, targBo := w.bodies[src.ID()], w.bodies[targ.ID()]
	dealt, _, destroyed := targBo.damagePart(bPart, dmg)
	leftover = dmg - dealt
	if !destroyed {
		if soulInvolved(src, targ) {
			w.log("%s's %s dealt %v damage to %s's %s",
				w.getName(src, "!?!"), srcBo.DescribePart(aPart),
				dealt,
				w.getName(targ, "?!?"), targBo.DescribePart(bPart),
			)
		}

		// TODO decouple damage -> agro into a separate upsert after the
		// damage proc; requires damage/kill relations
		// TODO better support upsert reduction
		accum := ecs.NilEntity
		w.moves.Upsert(
			w.moves.Select(mrAgro.All(), ecs.InA(targ.ID()), ecs.InB(src.ID())),
			func(uc *ecs.UpsertCursor) {
				if accum != ecs.NilEntity {
					if next := uc.R(); next.Type().HasAll(movN) {
						w.moves.n[accum.ID()] += w.moves.n[next.ID()]
					}
					return
				}
				accum = uc.Emit(mrAgro|movN, targ, src)
				w.moves.n[accum.ID()] += dmg
			})
		return leftover, false
	}

	if soulInvolved(src, targ) {
		w.log("%s's %s destroyed by %s's %s",
			w.getName(targ, "?!?"), targBo.DescribePart(bPart),
			w.getName(src, "!?!"), srcBo.DescribePart(aPart),
		)
	}

	severed := targBo.sever(w.log, bPart)
	if severed != nil {
		targName := w.getName(targ, "nameless")
		name := fmt.Sprintf("remains of %s", targName)
		pos, _ := w.pos.Get(targ)
		item := w.newItem(pos, name, '%', severed)
		w.timers.Every(item, 5, w.decayRemains)
		if severed.Len() > 0 {
			w.log("%s's remains have dropped on the floor", targName)
		} else {
			w.log("empty severed? %v",
				w.getName(targ, "?!?"), targBo.DescribePart(bPart),
			)
		}
	}

	bPart.Destroy()

	// var xx []string
	// for _, root := range targBo.rel.Roots(ecs.AllRel(brControl), nil) {
	// 	xx = append(xx, targBo.DescribePart(root))
	// }
	// w.log("roots: %v", xx)
	// for cur := targBo.rel.Cursor(ecs.AllRel(brControl), nil); cur.Scan(); {
	// 	w.log("%v => %v (%v)",
	// 		targBo.DescribePart(cur.A()),
	// 		targBo.DescribePart(cur.B()),
	// 		cur.Entity())
	// }

	if bo := w.bodies[targ.ID()]; bo.Iter(bcPart.All()).Count() > 0 {
		return leftover, false
	}

	// may become spirit
	if severed != nil {
		heads, spi := severed.allHeads(), 0
		for _, head := range heads {
			spi += severed.hp[head.ID()]
		}
		if spi > 0 {
			targ.Delete(wcBody | wcCollide)
			w.Glyphs[targ.ID()] = '⟡'
			for _, head := range heads {
				head.Add(bcDerived)
				severed.derived[head.ID()] = targ
			}
			if soulInvolved(src, targ) {
				w.log("%s was disembodied by %s", w.getName(targ, "?!?"), w.getName(src, "!?!"))
			}
			return leftover, true
		}
	}

	targ.Destroy()

	return leftover, true
}

func (w *world) decayRemains(item ecs.Entity) {
	rem := w.items[item.ID()].(*body)
	for it := rem.Iter(bcHPart.All()); it.Next(); {
		id := it.ID()
		rem.hp[id]--
		if rem.hp[id] <= 0 {
			it.Entity().Destroy()
		}
	}
	if rem.Len() == 0 {
		pos, _ := w.pos.Get(item)
		w.dirtyFloorTile(pos)
		// TODO: do something neat after max dirty (spawn something
		// creepy)
		item.Destroy()
	}
}

func (w *world) dirtyFloorTile(pos image.Point) (ecs.Entity, bool) {
	for _, tile := range w.pos.At(pos) {
		if !tile.Type().HasAll(floorTileMask) {
			continue
		}
		bg := w.BG[tile.ID()]
		for i := range floorColors {
			if floorColors[i] == bg {
				j := i + 1
				canDirty := j < len(floorColors)
				if canDirty {
					w.BG[tile.ID()] = floorColors[j]
				}
				return tile, canDirty
			}
		}
	}
	return ecs.NilEntity, false
}

func (w *world) checkAttackHit(src, targ ecs.Entity) (ecs.Entity, ecs.Entity) {
	aPart := w.chooseAttackerPart(src)
	if aPart == ecs.NilEntity {
		if soulInvolved(src, targ) {
			w.log("%s has nothing to hit %s with.", w.getName(src, "!?!"), w.getName(targ, "?!?"))
		}
		return ecs.NilEntity, ecs.NilEntity
	}
	bPart := w.chooseAttackedPart(targ)
	if bPart == ecs.NilEntity {
		if soulInvolved(src, targ) {
			w.log("%s can find nothing worth hitting on %s.", w.getName(src, "!?!"), w.getName(targ, "?!?"))
		}
		return ecs.NilEntity, ecs.NilEntity
	}
	return aPart, bPart
}

func (w *world) chooseAttackerPart(ent ecs.Entity) ecs.Entity {
	bo := w.bodies[ent.ID()]
	return bo.chooseRandomPart(w.rng, func(part ecs.Entity) int {
		if bo.dmg[part.ID()] <= 0 {
			return 0
		}
		return 4*bo.dmg[part.ID()] + 2*bo.armor[part.ID()] + bo.hp[part.ID()]
	})
}

func (w *world) chooseAttackedPart(ent ecs.Entity) ecs.Entity {
	bo := w.bodies[ent.ID()]
	return bo.chooseRandomPart(w.rng, func(part ecs.Entity) int {
		id := part.ID()
		hp := bo.hp[id]
		if hp <= 0 {
			return 0
		}
		maxHP := bo.maxHP[id]
		armor := bo.armor[id]
		stick := 1 + maxHP - hp
		switch part.Type() & bcPartMask {
		case bcTorso:
			stick *= 100
		case bcHead:
			stick *= 10
		}
		return stick - armor
	})
}

func (w *world) addBox(box point.Box, glyph rune) {
	// TODO: the box should be an entity, rather than each cell
	last, sz, pos := wallTable.Ref(1), box.Size(), image.Point(box.TopLeft)
	for _, r := range []struct {
		n int
		d image.Point
	}{
		{n: sz.X, d: image.Pt(1, 0)},
		{n: sz.Y, d: image.Pt(0, 1)},
		{n: sz.X, d: image.Pt(-1, 0)},
		{n: sz.Y, d: image.Pt(0, -1)},
	} {
		for i := 0; i < r.n; i++ {
			wall := w.AddEntity(wcPosition | wcCollide | wcSolid | wcGlyph | wcBG | wcFG | wcWall)
			w.Glyphs[wall.ID()] = glyph
			w.pos.Set(wall, pos)
			c, _ := wallTable.toColor(last)
			w.BG[wall.ID()] = c
			w.FG[wall.ID()] = c + 1
			pos = pos.Add(r.d)
			last = wallTable.ChooseNext(w.rng, last)
		}
	}

	floorTable.genTile(w.rng, box, func(pos point.Point, bg termbox.Attribute) {
		floor := w.AddEntity(wcPosition | wcBG | wcFloor)
		w.pos.Set(floor, image.Point(pos))
		w.BG[floor.ID()] = bg
	})
}

func (w *world) newItem(pos image.Point, name string, glyph rune, val worldItem) ecs.Entity {
	ent := w.AddEntity(wcPosition | wcCollide | wcName | wcGlyph | wcItem)
	w.pos.Set(ent, pos)
	w.Glyphs[ent.ID()] = glyph
	w.Names[ent.ID()] = name
	w.items[ent.ID()] = val
	return ent
}

func (w *world) newChar(name string, glyph rune, t ecs.ComponentType) ecs.Entity {
	ent := w.AddEntity(charMask | wcWaiting | t)
	w.Glyphs[ent.ID()] = glyph
	w.Names[ent.ID()] = name
	w.bodies[ent.ID()].build(w.rng)
	return ent
}

func (w *world) log(mess string, args ...interface{}) {
	s := fmt.Sprintf(mess, args...)
	for _, rule := range []struct{ old, new string }{
		{"you's", "your"},
		{"them's", "their"},
		{"you has", "you have"},
		{"you was", "you were"},
	} {
		s = strings.Replace(s, rule.old, rule.new, -1)
	}
	w.ui.Log(s)
	if w.logger != nil {
		w.logger.Printf(s)
	}
}

func (w *world) getName(ent ecs.Entity, deflt string) string {
	if !ent.Type().HasAll(wcName) {
		return deflt
	}
	if w.Names[ent.ID()] == "" {
		return deflt
	}
	return w.Names[ent.ID()]
}

func (mov *moves) decayN(rel ecs.Entity) {
	n := 0
	if rel.Type().HasAll(movN) {
		id := rel.ID()
		n = mov.n[id]
		n--
		mov.n[id] = n
	}
	if n <= 0 {
		rel.Delete(movT)
	}
}

func (w *world) addFrustration(ent ecs.Entity, n int) {
	// TODO better support upsert reduction
	accum := ecs.NilEntity
	w.moves.Upsert(
		w.moves.Select(mrAgro.All(), ecs.InA(ent.ID()), ecs.InB(ent.ID())),
		func(uc *ecs.UpsertCursor) {
			if accum != ecs.NilEntity {
				if next := uc.R(); next.Type().HasAll(movN) {
					w.moves.n[accum.ID()] += w.moves.n[next.ID()]
				}
				return
			}
			rel := uc.Emit(mrAgro|movN, ent, ent)
			w.moves.n[rel.ID()] += n
			w.moves.timers.Every(rel, 1, w.moves.decayN)
		})
}

func (w *world) getFrustration(ent ecs.Entity) (n int) {
	for cur := w.moves.Select(mrAgro.All(), ecs.InA(w.Deref(ent)), ecs.InB(w.Deref(ent))); cur.Scan(); {
		n += w.moves.n[cur.R().ID()]
	}
	return n
}

func (w *world) addSpawn(x, y int) ecs.Entity {
	spawn := w.AddEntity(wcPosition | wcGlyph | wcFG | wcSpawn)
	w.pos.Set(spawn, image.Pt(x, y))
	w.Glyphs[spawn.ID()] = '✖' // ×
	w.FG[spawn.ID()] = 54
	return spawn
}

func main() {
	if err := view.JustKeepRunning(func(v *view.View) (view.Client, error) {
		w, err := newWorld(v)
		if err != nil {
			return nil, err
		}

		pt := point.Point{X: 12, Y: 8}
		w.addBox(point.Box{TopLeft: pt.Neg(), BottomRight: pt}, '#')

		w.addSpawn(0, -5)
		w.addSpawn(-8, 5)
		w.addSpawn(8, 5)

		player := w.newChar("you", 'X', wcSoul)
		w.ui.bar.addAction(newRangeChooser(w, player))

		w.Process()

		return w, nil
	}); err != nil {
		log.Fatal(err)
	}
}
