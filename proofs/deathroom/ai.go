package main

import (
	"image"

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
		var move image.Point
		if target, found := w.aiTarget(ai); found {
			pos, _ := w.pos.Get(ai)
			move = point.Sign(target.Sub(pos))
		}
		w.moves.AddPendingMove(ai, move, 1, maxRestingCharge)
	}
}

func (w *world) aiTarget(ai ecs.Entity) (image.Point, bool) {
	// chase the thing we hate the most
	opp, hate := ecs.NilEntity, 0
	for cur := w.moves.Select(mrAgro.All(), ecs.InA(ai.ID())); cur.Scan(); {
		if cur.B() == ai {
			continue
		}
		// TODO: take other factors like distance into account
		if n := w.moves.Mag(cur.R()); n > hate {
			if b := cur.B(); b.Type().HasAll(wcBody | wcSolid) {
				opp, hate = b, n
			}
		}
	}
	if opp != ecs.NilEntity {
		return w.pos.Get(opp)
	}

	// revert to our goal...
	goalPos, found := image.ZP, false
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

		move := uc.R()
		myPos, _ := w.pos.Get(ai)
		goalPos, _ = w.pos.Get(goal)

		lastPos, lastPosDef := w.moves.Dir(move)
		if !lastPosDef || !lastPos.Eq(myPos) {
			w.moves.SetMag(move, 0)
			w.moves.SetDir(move, myPos)
		} else {
			n := w.moves.Mag(move) + 1
			if n >= 3 {
				w.moves.timers.Every(w.addAgro(ai, ai, 32), 1, w.moves.decayN)
				// stuck trying to get that one, give up
				return
			}
			w.moves.SetMag(move, n)
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
	score := point.SumSQ(goalPos.Sub(myPos))
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
	for it := w.Iter(ecs.And(wcSolid.All(), wcBody.NotAll())); it.Next(); {
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

	for cur := w.moves.Collisions(); cur.Scan(); {
		ai, goal := cur.A(), cur.B()
		if _, isGoal := goals[ab{ai.ID(), goal.ID()}]; !isGoal {
			continue
		}

		if goal.Type().HasAll(wcItem) {
			// can haz?
			if pr, ok := w.itemPrompt(prompt.Prompt{}, ai); ok {
				w.runAIInteraction(pr, ai)
			}
			// can haz moar?
			if pr, ok := w.itemPrompt(prompt.Prompt{}, ai); !ok || pr.Len() == 0 {
				goals[ab{ai.ID(), goal.ID()}].Destroy()
			}
		} else {
			// have booped?
			goals[ab{ai.ID(), goal.ID()}].Destroy()
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
