package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/view/hud/prompt"
)

// TODO: more body parts: thigh/calf, forearm/upper, neck, fingers, toes, organs, joints, items

const (
	bcHP ecs.ComponentType = 1 << iota
	bcPart
	bcDerived
	// TODO damage and armor components
	bcName

	bcRight
	bcLeft

	// TODO: draw a line between flags above, and
	// monotonic / exclusive types below; probably a
	// 32/32-bit split is a good place to start

	bcHead     // o
	bcTorso    // O
	bcUpperArm // \ /
	bcForeArm  // \ /
	bcHand     // w
	bcThigh    // |
	bcCalf     // |
	bcFoot     // ^
	bcTail
)

const (
	bcHPart    = bcHP | bcPart
	bcArmPart  = bcUpperArm | bcForeArm | bcHand
	bcLegPart  = bcFoot | bcCalf | bcThigh
	bcPartMask = bcHead | bcTorso | bcTail | bcArmPart | bcLegPart
	bcLocMask  = bcRight | bcLeft
)

const (
	brControl ecs.ComponentType = 1 << iota
)

type body struct {
	ecs.Core
	rel ecs.Graph

	fmt   []string
	maxHP []int
	hp    []int
	dmg   []int
	armor []int

	derived []ecs.Entity // TODO: replace this with a Relation;
	// TODO: Relation would need foreign key support (i.e. everything in the B
	//       side may come from a different Core (body)).
	// TODO: Relation would need a new flag "last cascades" so that destroy
	//       only happens after the last relation is destroyed.

	coGT ecs.GraphTraverser
}

// TODO: split package and intermediate through:
// - body api
// - type Part struct { ecs.Entity }

// TODO: func (Part) Stats()
type bodyStats struct {
	HP, MaxHP int
	Damage    int
	Armor     int
}

// TODO: func (Part) Serialize()
type bodyPart struct {
	Type ecs.ComponentType
	Name string
	Desc string
	bodyStats
}

func newBody() *body {
	bo := &body{
		// TODO: consider eliminating the padding for EntityID(0)
		fmt:     []string{""},
		maxHP:   []int{0},
		hp:      []int{0},
		dmg:     []int{0},
		armor:   []int{0},
		derived: []ecs.Entity{ecs.NilEntity},
	}
	bo.rel.Init(&bo.Core, 0)
	bo.RegisterAllocator(bcPart, bo.allocPart)
	bo.RegisterDestroyer(bcDerived, bo.destroyDerived)
	bo.coGT = bo.rel.Traverse(brControl.All(), ecs.TraverseCoDFS)
	return bo
}

func (bo *body) Clear() {
	bo.rel.Clear()
	bo.Core.Clear()
}

func (bo *body) allocPart(id ecs.EntityID, t ecs.ComponentType) {
	bo.fmt = append(bo.fmt, "")
	bo.maxHP = append(bo.maxHP, 0)
	bo.hp = append(bo.hp, 0)
	bo.dmg = append(bo.dmg, 0)
	bo.armor = append(bo.armor, 0)
	bo.derived = append(bo.derived, ecs.NilEntity)
}

func (bo *body) destroyDerived(id ecs.EntityID, t ecs.ComponentType) {
	if ent := bo.derived[id]; ent != ecs.NilEntity {
		bo.derived[id] = ecs.NilEntity
		any := false
		for it := bo.Iter(bcDerived.All()); it.Next(); {
			if bo.derived[it.ID()] == ent {
				any = true
				break
			}
		}
		if !any {
			ent.Destroy()
		}
	}
}

func (bo *body) build(rng *rand.Rand) {
	head := bo.AddPart(bcHead, 5, 3, 4)
	torso := bo.AddPart(bcTorso, 8, 0, 2)

	rightUpperArm := bo.AddPart(bcRight|bcUpperArm, 5, 3, 2)
	leftUpperArm := bo.AddPart(bcLeft|bcUpperArm, 5, 3, 2)

	rightForeArm := bo.AddPart(bcRight|bcForeArm, 4, 4, 2)
	leftForeArm := bo.AddPart(bcLeft|bcForeArm, 4, 4, 2)

	rightHand := bo.AddPart(bcRight|bcHand, 2, 5, 1)
	leftHand := bo.AddPart(bcLeft|bcHand, 2, 5, 1)

	rightThigh := bo.AddPart(bcRight|bcThigh, 6, 1, 3)
	leftThigh := bo.AddPart(bcLeft|bcThigh, 6, 1, 3)

	rightCalf := bo.AddPart(bcRight|bcCalf, 4, 5, 3)
	leftCalf := bo.AddPart(bcLeft|bcCalf, 4, 5, 3)

	rightFoot := bo.AddPart(bcRight|bcFoot, 3, 6, 2)
	leftFoot := bo.AddPart(bcLeft|bcFoot, 3, 6, 2)

	bo.rel.Upsert(nil, func(uc *ecs.UpsertCursor) {
		uc.Create(brControl, head, torso)
		uc.Create(brControl, torso, rightUpperArm)
		uc.Create(brControl, torso, leftUpperArm)
		uc.Create(brControl, torso, rightThigh)
		uc.Create(brControl, torso, leftThigh)
		uc.Create(brControl, rightUpperArm, rightForeArm)
		uc.Create(brControl, leftUpperArm, leftForeArm)
		uc.Create(brControl, rightForeArm, rightHand)
		uc.Create(brControl, leftForeArm, leftHand)
		uc.Create(brControl, rightThigh, rightCalf)
		uc.Create(brControl, leftThigh, leftCalf)
		uc.Create(brControl, rightCalf, rightFoot)
		uc.Create(brControl, leftCalf, leftFoot)
	})
}

func (bo *body) AddPart(t ecs.ComponentType, hp, dmg, armor int) ecs.Entity {
	ent := bo.AddEntity(bcHP | bcPart | t)
	id := ent.ID()
	bo.maxHP[id] = hp
	bo.hp[id] = hp
	bo.dmg[id] = dmg
	bo.armor[id] = armor
	return ent
}

func (bo *body) Stats() bodyStats {
	var s bodyStats
	for it := bo.Iter(bcHPart.All()); it.Next(); {
		s.HP += bo.hp[it.ID()]
		s.MaxHP += bo.maxHP[it.ID()]
		s.Damage += bo.dmg[it.ID()]
		s.Armor += bo.armor[it.ID()]
	}
	return s
}

func (bo *body) HPRange() (hp, maxHP int) {
	for it := bo.Iter(bcHP.All()); it.Next(); {
		hp += bo.hp[it.ID()]
		maxHP += bo.maxHP[it.ID()]
	}
	return
}

func (bo *body) HP() int {
	hp := 0
	for it := bo.Iter(bcHP.All()); it.Next(); {
		hp += bo.hp[it.ID()]
	}
	return hp
}

func (bo *body) choosePart(want func(prior, ent ecs.Entity) bool) ecs.Entity {
	var choice ecs.Entity
	for it := bo.Iter(bcHPart.All()); it.Next(); {
		if want(choice, it.Entity()) {
			choice = it.Entity()
		}
	}
	return choice
}

func (bo *body) chooseRandomPart(rng *rand.Rand, score func(ent ecs.Entity) int) ecs.Entity {
	sum := 0
	return bo.choosePart(func(prior, ent ecs.Entity) bool {
		if w := score(ent); w > 0 {
			sum += w
			return prior == ecs.NilEntity || rng.Intn(sum) < w
		}
		return false
	})
}

func (bo *body) PartName(ent ecs.Entity) string {
	switch ent.Type() & bcPartMask {
	case bcHead:
		return "head"
	case bcTorso:
		return "torso"
	case bcUpperArm:
		return "upper arm"
	case bcForeArm:
		return "forearm"
	case bcHand:
		return "hand"
	case bcThigh:
		return "thigh"
	case bcCalf:
		return "calf"
	case bcFoot:
		return "foot"
	case bcTail:
		return "tail"
	}
	return ""
}

func (bo *body) PartAbbr(ent ecs.Entity) string {
	var s string
	switch ent.Type() & bcPartMask {
	case bcHead:
		return "Hd"
	case bcTorso:
		return "By"
	case bcUpperArm:
		s = "uA"
	case bcForeArm:
		s = "fA"
	case bcHand:
		s = "Hn"
	case bcThigh:
		s = "Th"
	case bcCalf:
		s = "Lg"
	case bcFoot:
		s = "Ft"
	case bcTail:
		s = "Tl"
	default:
		return "?"
	}
	switch ent.Type() & bcLocMask {
	case bcRight:
		return "R" + s
	case bcLeft:
		return "L" + s
	default:
		return s
	}
}

func (bo *body) DescribePart(ent ecs.Entity) string {
	s := bo.PartName(ent)
	if s == "" {
		s = "???"
	}
	switch ent.Type() & bcLocMask {
	case bcRight:
		s = "right " + s
	case bcLeft:
		s = "left " + s
	}
	if ent.Type().HasAll(bcName) {
		s = fmt.Sprintf(bo.fmt[ent.ID()], s)
	}
	return s
}

func (bo *body) damagePart(ent ecs.Entity, dmg int) (int, bodyPart, bool) {
	hp := bo.hp[ent.ID()]
	if dmg < hp {
		bo.hp[ent.ID()] = hp - dmg
		return dmg, bodyPart{}, false
	}
	bo.hp[ent.ID()] = 0
	return hp, bodyPart{
		Type: ent.Type(),
		Name: bo.PartName(ent),
		Desc: bo.DescribePart(ent),
		bodyStats: bodyStats{
			HP:     bo.hp[ent.ID()],
			MaxHP:  bo.maxHP[ent.ID()],
			Damage: bo.dmg[ent.ID()],
			Armor:  bo.armor[ent.ID()],
		},
	}, true
}

func (bo *body) allHeads() []ecs.Entity {
	it := bo.Iter(bcHead.All())
	r := make([]ecs.Entity, 0, it.Count())
	for it.Next() {
		r = append(r, it.Entity())
	}
	return r
}

func (bo *body) partHPRating(ent ecs.Entity) float64 {
	rating := 1.0
	for bo.coGT.Init(ent.ID()); bo.coGT.Traverse(); {
		id := bo.coGT.Node().ID()
		rating *= float64(bo.hp[id]) / float64(bo.maxHP[id])
	}
	return rating
}

func (bo *body) sever(
	log func(string, ...interface{}),
	ents ...ecs.Entity,
) *body {
	// TODO: consider using a DFS traversal

	type rel struct {
		r, a, b ecs.Entity
	}

	var (
		cont  = newBody()
		xlate = make(map[ecs.EntityID]ecs.EntityID)
		q     = make([]ecs.EntityID, len(ents))
		n     = bo.rel.Len()
		rels  = make([]rel, 0, n)
		relis = make(map[ecs.EntityID]struct{}, n)
		entis = make(map[ecs.EntityID]struct{}, n)
	)
	for i := range ents {
		q[i] = bo.Deref(ents[i])
	}

	for len(q) > 0 {
		id := q[0]
		copy(q, q[1:])
		q = q[:len(q)-1]
		if _, seen := entis[id]; seen {
			continue
		}
		entis[id] = struct{}{}

		ent := bo.Ref(id)
		if !ent.Type().HasAll(bcPart) {
			continue
		}

		if bo.hp[id] > 0 {
			eid := cont.AddEntity(ent.Type()).ID()
			xlate[id] = eid
			cont.fmt[eid] = bo.fmt[id]
			cont.maxHP[eid] = bo.maxHP[id]
			cont.hp[eid] = bo.hp[id]
			cont.dmg[eid] = bo.dmg[id]
			cont.armor[eid] = bo.armor[id]
		}

		// collect affected relations for final processing
		for cur := bo.rel.Select(brControl.All(), ecs.InA(ent.ID())); cur.Scan(); {
			id := cur.R().ID()
			if _, seen := relis[id]; !seen {
				relis[id] = struct{}{}
				rels = append(rels, rel{cur.R(), cur.A(), cur.B()})
				q = append(q, cur.B().ID())
			}
		}

		defer ent.Destroy()

		if len(q) == 0 {
			nh, nt, it := 0, 0, bo.Iter(bcPart.All())
			for it.Next() {
				if _, gone := entis[it.ID()]; !gone {
					if it.Type().HasAll(bcHead) {
						nh++
					} else if it.Type().HasAll(bcTorso) {
						nt++
					}
				}
			}
			if nh == 0 || nt == 0 {
				it.Reset()
				for it.Next() {
					q = append(q, it.ID())
				}
			}
		}
	}

	if cont.Len() == 0 {
		return nil
	}

	// finish relation processing
	cont.rel.Upsert(nil, func(uc *ecs.UpsertCursor) {
		for _, rel := range rels {
			if xa, def := xlate[rel.a.ID()]; def {
				a := cont.Ref(xa)
				if xb, def := xlate[rel.b.ID()]; def {
					b := cont.Ref(xb)
					uc.Create(brControl, a, b)
				}
			}
			defer rel.r.Destroy()
		}
	})

	return cont
}

type bodyRemains struct {
	w    *world     // the world it's in
	bo   *body      // the body it's in
	part ecs.Entity // the part
	item ecs.Entity // its container
	ent  ecs.Entity // what's interacting with it
}

func (rem bodyRemains) describeScavenge() string {
	return fmt.Sprintf("scavenge %s (armor:%+d damage:%+d)",
		rem.bo.DescribePart(rem.part),
		rem.bo.armor[rem.part.ID()],
		rem.bo.dmg[rem.part.ID()])
}

func (rem bodyRemains) scavenge(pr prompt.Prompt) (prompt.Prompt, bool) {
	defer rem.part.Destroy()

	entBo := rem.w.bodies[rem.ent.ID()]

	imp := make([]string, 0, 2)
	if armor := rem.bo.armor[rem.part.ID()]; armor > 0 {
		recv := rem.w.chooseAttackedPart(rem.ent)
		entBo.armor[recv.ID()] += armor
		imp = append(imp, fmt.Sprintf("%s armor +%v", entBo.DescribePart(recv), armor))
	}
	if damage := rem.bo.dmg[rem.part.ID()]; damage > 0 {
		recv := rem.w.chooseAttackerPart(rem.ent)
		entBo.dmg[recv.ID()] += damage
		imp = append(imp, fmt.Sprintf("%s damage +%v", entBo.DescribePart(recv), damage))
	}

	if soulInvolved(rem.ent, rem.item) {
		rem.w.log("%s gained %v from %s's %s",
			rem.w.getName(rem.ent, "unknown"),
			strings.Join(imp, " and "),
			rem.w.getName(rem.item, "unknown"),
			rem.bo.DescribePart(rem.part),
		)
	}

	if rem.bo.Len() == 0 {
		defer rem.item.Destroy()
	}

	pr, _ = rem.w.itemPrompt(pr.Unwind(), rem.ent)
	return pr, false
}
