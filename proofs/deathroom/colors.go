package main

import (
	"math/rand"

	termbox "github.com/nsf/termbox-go"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/markov"
	"github.com/borkshop/bork/internal/point"
)

var (
	aiColors    = []termbox.Attribute{124, 160, 196, 202, 208, 214}
	soulColors  = []termbox.Attribute{19, 20, 21, 27, 33, 39}
	itemColors  = []termbox.Attribute{22, 23, 29, 35, 41, 47}
	wallColors  = []termbox.Attribute{233, 234, 235, 236, 237, 238, 239}
	floorColors = []termbox.Attribute{232, 233, 234}

	wallTable  = newColorTable()
	floorTable = newColorTable()
)

func init() {
	wallTable.addLevelTransitions(wallColors, 12, 2, 2, 12, 2)
	floorTable.addLevelTransitions(floorColors, 24, 1, 30, 2, 1)
}

const (
	componentTableColor ecs.ComponentType = 1 << iota
)

type colorTable struct {
	ecs.Core
	*markov.Table
	color  []termbox.Attribute
	lookup map[termbox.Attribute]ecs.EntityID
}

func newColorTable() *colorTable {
	ct := &colorTable{
		// TODO: consider eliminating the padding for EntityID(0)
		color:  []termbox.Attribute{0},
		lookup: make(map[termbox.Attribute]ecs.EntityID, 1),
	}
	ct.Table = markov.NewTable(&ct.Core)
	ct.RegisterAllocator(componentTableColor, ct.allocTableColor)
	ct.RegisterDestroyer(componentTableColor, ct.destroyTableColor)
	return ct
}

func (ct *colorTable) allocTableColor(id ecs.EntityID, t ecs.ComponentType) {
	ct.color = append(ct.color, 0)
}

func (ct *colorTable) destroyTableColor(id ecs.EntityID, t ecs.ComponentType) {
	delete(ct.lookup, ct.color[id])
	ct.color[id] = 0
}

func (ct *colorTable) addLevelTransitions(
	colors []termbox.Attribute,
	zeroOn, zeroUp int,
	oneDown, oneOn, oneUp int,
) {
	n := len(colors)
	c0 := colors[0]

	for i, c1 := range colors {
		if c1 == c0 {
			continue
		}

		ct.addTransition(c0, c0, (n-i)*zeroOn)
		ct.addTransition(c0, c1, (n-i)*zeroUp)

		ct.addTransition(c1, c0, (n-1)*oneDown)
		ct.addTransition(c1, c1, (n-1)*oneOn)

		for _, c2 := range colors {
			if c2 != c1 && c2 != c0 {
				ct.addTransition(c1, c2, (n-1)*oneUp)
			}
		}
	}
}

func (ct *colorTable) toEntity(a termbox.Attribute) ecs.Entity {
	if id, def := ct.lookup[a]; def {
		return ct.Ref(id)
	}
	ent := ct.AddEntity(componentTableColor)
	id := ent.ID()
	ct.color[id] = a
	ct.lookup[a] = id
	return ent
}

func (ct *colorTable) toColor(ent ecs.Entity) (termbox.Attribute, bool) {
	if !ent.Type().All(componentTableColor) {
		return 0, false
	}
	return ct.color[ent.ID()], true
}

func (ct *colorTable) addTransition(a, b termbox.Attribute, w int) (ae, be ecs.Entity) {
	ae, be = ct.toEntity(a), ct.toEntity(b)
	ct.AddTransition(ae, be, w)
	return
}

func (ct *colorTable) genTile(
	rng *rand.Rand,
	box point.Box,
	f func(point.Point, termbox.Attribute),
) {
	// TODO: better 2d generation
	last := floorTable.Ref(1)
	var pos point.Point
	for pos.Y = box.TopLeft.Y + 1; pos.Y < box.BottomRight.Y; pos.Y++ {
		first := last
		for pos.X = box.TopLeft.X + 1; pos.X < box.BottomRight.X; pos.X++ {
			c, _ := floorTable.toColor(last)
			f(pos, c)
			last = floorTable.ChooseNext(rng, last)
		}
		last = floorTable.ChooseNext(rng, first)
	}
}
