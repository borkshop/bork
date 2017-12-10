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
	for it := w.Iter(ecs.All(aiMoveMask)); it.Next(); {
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
	for cur := w.moves.LookupA(ecs.AllRel(mrAgro), ai.ID()); cur.Scan(); {
		if cur.B() == ai {
			continue
		}
		if ent := cur.Entity(); ent.Type().All(movN) {
			// TODO: take other factors like distance into account
			if n := w.moves.n[ent.ID()]; n > hate {
				if b := cur.B(); b.Type().All(combatMask) {
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
	w.moves.UpsertMany(
		ecs.AllRel(mrGoal),
		func(r ecs.RelationType, ent, a, b ecs.Entity) bool { return a == ai },
		func(
			r ecs.RelationType, rel, _, goal ecs.Entity,
			emit func(r ecs.RelationType, a, b ecs.Entity,
			) ecs.Entity) {
			if goal == ecs.NilEntity {
				// no goal, pick one!
				if goal := w.chooseAIGoal(ai); goal != ecs.NilEntity {
					emit(mrGoal, ai, goal)
					goalPos, found = w.pos.Get(goal)
				}
				return
			}

			if !goal.Type().All(wcPosition) {
				// no position, drop it
				return
			}

			myPos, _ := w.pos.Get(ai)
			goalPos, _ = w.pos.Get(goal)
			if id := rel.ID(); !rel.Type().All(movN | movP) {
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
			emit(mrGoal, ai, goal)
		})

	return goalPos, found
}

func (w *world) scoreAIGoal(ai, goal ecs.Entity) int {
	const itemLimit = 25

	myPos, _ := w.pos.Get(ai)
	goalPos, _ := w.pos.Get(goal)
	score := goalPos.Sub(myPos).SumSQ()
	if goal.Type().All(wcItem) {
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
	for it := w.Iter(ecs.All(collMask)); it.Next(); {
		if it.Type().All(combatMask) {
			continue
		}
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
	for cur := w.moves.Cursor(
		ecs.AllRel(mrGoal),
		func(r ecs.RelationType, ent, a, b ecs.Entity) bool {
			return a.Type().All(aiMoveMask)
		},
	); cur.Scan(); {
		goals[ab{cur.A().ID(), cur.B().ID()}] = cur.Entity()
	}

	for cur := w.moves.Cursor(
		ecs.RelClause(mrCollide, mrItem|mrHit),
		func(r ecs.RelationType, ent, a, b ecs.Entity) bool {
			_, isGoal := goals[ab{a.ID(), b.ID()}]
			return isGoal
		},
	); cur.Scan(); {
		ai, b := cur.A(), cur.B()

		if b.Type().All(wcItem) {
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
