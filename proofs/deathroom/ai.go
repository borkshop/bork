package main

import (
	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view/hud/prompt"
)

const (
	aiMoveMask = wcPosition | wcInput | wcAI
)

func (w *world) generateAIMoves() {
	for it := w.Iter(aiMoveMask.All()); it.Next(); {
		ai := it.Entity()
		// TODO: if too damaged, rest
		var move point.Point
		if target, found := w.aiTarget(ai); found {
			pos, _ := w.pos.Get(ai)
			move = target.Sub(pos).Sign()
		}
		w.addPendingMove(ai, move)
	}
}

func (w *world) aiTarget(ai ecs.Entity) (point.Point, bool) {
	// chase the thing we hate the most
	opp, hate := ecs.NilEntity, 0
	for cur := w.moves.Select(mrAgro.All(), ecs.InA(ai.ID())); cur.Scan(); {
		if cur.B() == ai {
			continue
		}
		if ent := cur.R(); ent.Type().HasAll(movN) {
			// TODO: take other factors like distance into account
			if n := w.moves.n[ent.ID()]; n > hate {
				if b := cur.B(); b.Type().HasAll(combatMask) {
					opp, hate = b, n
				}
			}
		}
	}
	if opp != ecs.NilEntity {
		return w.pos.Get(opp)
	}

	// revert to our goal...
	goalPos, found := point.Zero, false
	w.moves.Upsert(w.moves.Select(mrGoal.All(), ecs.InA(ai.ID())), func(uc *ecs.UpsertCursor) {
		goal := uc.B()
		if goal == ecs.NilEntity {
			// no goal, pick one!
			if goal := w.chooseAIGoal(ai); goal != ecs.NilEntity {
				uc.Emit(mrGoal, ai, goal)
				goalPos, found = w.pos.Get(goal)
			}
			return
		}

		if !goal.Type().HasAll(wcPosition) {
			// no position, drop it
			return
		}

		rel := uc.R()
		myPos, _ := w.pos.Get(ai)
		goalPos, _ = w.pos.Get(goal)
		if id := rel.ID(); !rel.Type().HasAll(movN | movP) {
			rel.Add(movN | movP)
			w.moves.p[id] = myPos
		} else if lastPos := w.moves.p[id]; lastPos != myPos {
			w.moves.n[id] = 0
			w.moves.p[id] = myPos
		} else {
			w.moves.n[id]++
			if w.moves.n[id] >= 3 {
				w.addFrustration(ai, 32)
				// stuck trying to get that one, give up
				return
			}
		}
		found = true

		// see if we can do better
		alt := w.chooseAIGoal(ai)
		score := w.scoreAIGoal(ai, goal)
		score *= 16 // inertia bonus
		altScore := w.scoreAIGoal(ai, alt)
		if w.rng.Intn(score+altScore) < altScore {
			pos, _ := w.pos.Get(alt)
			goal, goalPos = alt, pos
		}

		// keep or update
		uc.Emit(mrGoal, ai, goal)
	})

	return goalPos, found
}

func (w *world) scoreAIGoal(ai, goal ecs.Entity) int {
	const itemLimit = 25

	myPos, _ := w.pos.Get(ai)
	goalPos, _ := w.pos.Get(goal)
	score := goalPos.Sub(myPos).SumSQ()
	if goal.Type().HasAll(wcItem) {
		if score > itemLimit {
			return 0
		}
		return (itemLimit - score) * itemLimit
	}
	return score
}

func (w *world) chooseAIGoal(ai ecs.Entity) ecs.Entity {
	// TODO: doesn't always cause progress, get stuck on the edge sometimes
	goal, sum := ecs.NilEntity, 0
	for it := w.Iter(ecs.And(collMask.All(), combatMask.NotAll())); it.Next(); {
		if score := w.scoreAIGoal(ai, it.Entity()); score > 0 {
			sum += score
			if sum <= 0 || w.rng.Intn(sum) < score {
				goal = it.Entity()
			}
		}
	}
	return goal
}

func (w *world) processAIItems() {
	type ab struct{ a, b ecs.EntityID }
	goals := make(map[ab]ecs.Entity)
	for cur := w.moves.Select(
		mrGoal.All(),
		// TODO ecs.WhereAType(aiMoveMask)
		ecs.Filter(func(cur ecs.Cursor) bool { return cur.A().Type().HasAll(aiMoveMask) }),
	); cur.Scan(); {
		goals[ab{cur.A().ID(), cur.B().ID()}] = cur.R()
	}

	for cur := w.moves.Select(
		ecs.And(mrCollide.All(), (mrItem|mrHit).Any()),
		ecs.Filter(func(cur ecs.Cursor) bool {
			_, isGoal := goals[ab{cur.A().ID(), cur.B().ID()}]
			return isGoal
		}),
	); cur.Scan(); {
		ai, b := cur.A(), cur.B()

		if b.Type().HasAll(wcItem) {
			// can haz?
			if pr, ok := w.itemPrompt(prompt.Prompt{}, ai); ok {
				w.runAIInteraction(pr, ai)
			}
			// can haz moar?
			if pr, ok := w.itemPrompt(prompt.Prompt{}, ai); !ok || pr.Len() == 0 {
				goals[ab{ai.ID(), b.ID()}].Destroy()
			}
		} else {
			// have booped?
			goals[ab{ai.ID(), b.ID()}].Destroy()
		}
	}
}

func (w *world) runAIInteraction(pr prompt.Prompt, ai ecs.Entity) {
	for n := pr.Len(); n > 0; n = pr.Len() {
		next, prompting, valid := pr.Run(w.rng.Intn(n))
		if !valid || !prompting {
			return
		}
		pr = next
	}
}
