package main

import (
	"image"

	"github.com/gdamore/tcell"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/ecs/eps"
	"github.com/borkshop/bork/internal/ecs/time"
	"github.com/borkshop/bork/internal/point"
)

const (
	wcPosition ecs.ComponentType = 1 << iota
	wcSolid
	wcTimer
	wcName
	wcGlyph
	wcBG
	wcFG
	wcHP
	wcStats
	wcFloor
	wcWall
	wcPlayerControl
)

type worldT struct {
	ecs.System
	pos    eps.EPS
	moves  eps.Moves
	timers time.Facility
	names  []string

	// TODO experiment with a []cell, where cell may have a glyph, fg, or bg
	// (at least one), and has a z-value; entities could then have 0-or-more cells
	glyphs []rune
	zval   []uint8
	bg     []tcell.Color
	fg     []tcell.Color

	hp    []hitpoints
	stats []charStats

	post   func(func())
	posted bool
}

type hitpoints struct {
	hp, maxHP int
}

type charStats struct {
	ap, maxAP int
}

func (world *worldT) postProc() {
	if !world.posted {
		world.posted = true
		world.post(world.Process)
	}
}

func (world *worldT) Process() {
	world.posted = false
	world.System.Process()
}

func (world *worldT) checkMove(uc *ecs.UpsertCursor, dir image.Point, mag int) (_ image.Point, _, limit int) {
	limit = 1
	return dir, mag, limit
}

func (world *worldT) init(post func(func())) {
	world.post = post

	world.timers.Init(&world.Core, wcTimer)
	world.pos.Init(&world.Core, wcPosition)
	world.moves.Init(&world.pos, wcSolid)
	world.moves.PreCheck = world.checkMove

	world.names = []string{""}
	world.glyphs = []rune{0}
	world.zval = []uint8{0}
	world.bg = []tcell.Color{tcell.ColorDefault}
	world.fg = []tcell.Color{tcell.ColorDefault}
	world.hp = []hitpoints{{maxHP: 20}}
	world.stats = []charStats{{
		ap:    0,
		maxAP: 128,
	}}

	world.RegisterAllocator(wcGlyph|wcBG|wcFG, world.allocCell)
	world.RegisterAllocator(wcName|wcHP|wcStats, world.allocStats)
	world.RegisterDestroyer(wcName, world.destroyName)
	world.RegisterDestroyer(wcGlyph, world.destroyGlyph)
	world.RegisterDestroyer(wcBG, world.destroyBG)
	world.RegisterDestroyer(wcFG, world.destroyFG)
	world.RegisterDestroyer(wcHP, world.destroyHP)
	world.RegisterDestroyer(wcStats, world.destroyStats)

	world.AddProc(
		&world.timers,
		&world.moves,
		ecs.ProcFunc(world.doDig),
	)
}

func (world *worldT) doDig() {
	dug := false
	// TODO BType / AType CursorOpts
	for cur := world.moves.Collisions(ecs.Filter(func(cur ecs.Cursor) bool {
		return cur.B().Type().HasAll(wcWall)
	})); cur.Scan(); {
		wall := cur.B()
		hp := &world.hp[wall.ID()]
		if hp.hp--; hp.hp > 0 {
			c := 0x30 + int32(hp.hp)*4
			world.bg[wall.ID()] = tcell.NewRGBColor(c, c, c)
			continue
		}

		pos, _ := world.pos.Get(wall)
		wall.Destroy()
		dug = true

		// TODO some sort of eps range query (or even a stencil!)
		for _, pt := range []image.Point{
			image.Pt(pos.X-1, pos.Y),
			image.Pt(pos.X+1, pos.Y),
			image.Pt(pos.X, pos.Y-1),
			image.Pt(pos.X-1, pos.Y-1),
			image.Pt(pos.X+1, pos.Y-1),
			image.Pt(pos.X, pos.Y+1),
			image.Pt(pos.X-1, pos.Y+1),
			image.Pt(pos.X+1, pos.Y+1),
		} {
			if len(world.pos.At(pt)) == 0 {
				world.addWall(pt, 0)
				world.addFloor(pt, 0)
			}
		}
	}
	if dug {
		// TODO partial update from dug, rather than a full re-analyze
		world.analyze()
	}
}

func (world *worldT) analyze() {
	// build a dense array of empty floor spaces, called tiles, along with an
	// EntityID-to-tileIndex mapping
	var (
		box    = world.pos.Bounds()
		frame  = point.ZFrame{Bounds: box}
		zmax   = frame.Key(box.Max)
		tiles  = make([]ecs.EntityID, zmax)
		spaces = make(map[ecs.EntityID]uint64, zmax)
	)
	for it := world.Iter((wcFloor | wcWall).Any(), wcPosition.All()); it.Next(); {
		pos, _ := world.pos.Get(it.Entity())
		ix := frame.Key(pos)
		t := it.Type()
		tid := tiles[ix]
		if t.HasAll(wcWall) {
			if tid > 0 {
				delete(spaces, tid)
			}
			tiles[ix] = -1
		} else if t.HasAll(wcFloor) && tid == 0 {
			id := it.ID()
			tiles[ix] = id
			spaces[id] = ix
		}
	}

	// erase wall placeholders, and prune hallways
	for ix := 0; ix < len(tiles); ix++ {
		if tiles[ix] < 0 {
			tiles[ix] = 0
			continue
		} else if tiles[ix] == 0 {
			continue
		}

		pos := frame.Point(uint64(ix))
		pat := nePat(0)
		for i, d := range nePos {
			if qpt := pos.Add(d); qpt.In(box) {
				if qix := frame.Key(qpt); tiles[qix] > 0 {
					pat |= 1 << uint(i)
					if neRoom[pat] {
						break
					}
				}
			}
		}
		if !neRoom[pat] {
			delete(spaces, tiles[ix])
			tiles[ix] = 0
		}
	}

	// assign a region to all remaining tiles spaces by flood filling
	// TODO this is basically a slightly-less-naive recursive flood fill;
	// should be able to do better using something from
	// http://www.adammil.net/blog/v126_A_More_Efficient_Flood_Fill.html
	numRegions := 0
	stack := make([]uint64, 64)
	for len(spaces) > 0 {
		for id, ix := range spaces {
			stack = append(stack[:0], ix)
			delete(spaces, id)
			break
		}
		regionID := 0
		for len(stack) > 0 {
			i := len(stack) - 1
			ix := stack[i]
			stack = stack[:i]

			if regionID == 0 {
				numRegions++
				regionID = numRegions
			}
			tiles[ix] = ecs.EntityID(-regionID)

			pos := frame.Point(ix)
			for _, d := range []image.Point{
				image.Pt(-1, 0), image.Pt(+1, 0),
				image.Pt(0, -1), image.Pt(0, +1),
			} {
				if pt := pos.Add(d); pt.In(box) {
					nix := frame.Key(pt)
					nid := tiles[nix]
					if _, def := spaces[nid]; def {
						stack = append(stack, nix)
						delete(spaces, nid)
					}
				}
			}
		}
	}

	// build region glyphs for the first 62 regions
	regionGlyphs := make([]rune, numRegions)
	i := 0
	for ; i < numRegions && i < 10; i++ {
		regionGlyphs[i] = rune('1' + i)
	}
	for ; i < numRegions && i < 10+26; i++ {
		regionGlyphs[i] = rune('a' + i - 10)
	}
	for ; i < numRegions && i < 10+26+26; i++ {
		regionGlyphs[i] = rune('A' + i - 10 - 26)
	}

	// assign tile region glyphs to floors (or clear any prior floor glyph)
	glyphColor := tcell.NewHexColor(0x505050)
	for it := world.Iter((wcFloor | wcPosition).All()); it.Next(); {
		tile := it.Entity()
		pos, _ := world.pos.Get(tile)
		ix := frame.Key(pos)
		var glyph rune
		if regionID := -tiles[ix]; regionID > 0 {
			glyph = regionGlyphs[regionID-1]
		}
		if glyph != 0 {
			tile.Add(wcGlyph | wcFG)
			world.glyphs[tile.ID()] = glyph
			world.fg[tile.ID()] = glyphColor
		} else {
			tile.Delete(wcGlyph | wcFG)
		}
	}

	// TODO proper artifacts, like storing numRegions, a
	// floor:region relation, etc
}

func (world *worldT) allocCell(id ecs.EntityID, _ ecs.ComponentType) {
	for n := int(id) + 1; len(world.zval) < n; {
		world.glyphs = append(world.glyphs, 0)
		world.zval = append(world.zval, 0)
		world.bg = append(world.bg, world.bg[0])
		world.fg = append(world.fg, world.fg[0])
	}
}

func (world *worldT) allocStats(id ecs.EntityID, _ ecs.ComponentType) {
	for n := int(id) + 1; len(world.names) < n; {
		world.names = append(world.names, "")
		world.hp = append(world.hp, world.hp[0])
		world.stats = append(world.stats, world.stats[0])
	}
}

func (world *worldT) destroyName(id ecs.EntityID, _ ecs.ComponentType) {
	world.names[id] = ""
}

func (world *worldT) destroyGlyph(id ecs.EntityID, t ecs.ComponentType) {
	world.glyphs[id] = 0
	if !t.HasAny(wcGlyph | wcBG | wcFG) {
		world.zval[id] = 0
	}
}
func (world *worldT) destroyBG(id ecs.EntityID, t ecs.ComponentType) {
	world.bg[id] = world.bg[0]
	if !t.HasAny(wcGlyph | wcBG | wcFG) {
		world.zval[id] = 0
	}
}
func (world *worldT) destroyFG(id ecs.EntityID, t ecs.ComponentType) {
	world.fg[id] = world.fg[0]
	if !t.HasAny(wcGlyph | wcBG | wcFG) {
		world.zval[id] = 0
	}
}
func (world *worldT) destroyHP(id ecs.EntityID, _ ecs.ComponentType) { world.hp[id] = world.hp[0] }
func (world *worldT) destroyStats(id ecs.EntityID, _ ecs.ComponentType) {
	world.stats[id] = world.stats[0]
}

func (world *worldT) addTile(bg tcell.Color, glyph rune, pos image.Point) ecs.Entity {
	t := wcBG | wcPosition
	if glyph != 0 {
		t |= wcGlyph
	}
	ent := world.AddEntity(t)
	id := ent.ID()
	if glyph != 0 {
		world.glyphs[id] = glyph
	}
	world.bg[id] = bg
	world.zval[id] = 100
	world.pos.Set(ent, pos)
	return ent
}

func (world *worldT) addBlock(bg tcell.Color, glyph rune, pos image.Point) ecs.Entity {
	ent := world.addTile(bg, glyph, pos)
	id := ent.ID()
	ent.Add(wcSolid)
	world.zval[id]++
	return ent
}

func (world *worldT) addFloor(pos image.Point, glyph rune) ecs.Entity {
	floor := world.addTile(tcell.NewHexColor(0x303030), glyph, pos)
	floor.Add(wcFloor)
	return floor
}

func (world *worldT) addWall(pos image.Point, glyph rune) ecs.Entity {
	wall := world.addBlock(tcell.NewHexColor(0x404040), glyph, pos)
	wall.Add(wcWall | wcHP)
	hp := &world.hp[wall.ID()]
	hp.hp = 4 // TODO = hp.maxHP
	return wall
}

func (world *worldT) addRoom(box image.Rectangle) {
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			world.addFloor(image.Pt(x, y), 0)
		}
	}

	w, h := box.Dx(), box.Dy()
	pos := box.Min
	for _, step := range []struct {
		d image.Point
		n int
	}{
		{image.Pt(1, 0), w - 1},
		{image.Pt(0, 1), h - 1},
		{image.Pt(-1, 0), w - 1},
		{image.Pt(0, -1), h - 1},
	} {
		for i := 0; i < step.n; i++ {
			world.addWall(pos, 0)
			pos = pos.Add(step.d)
		}
	}
}

func (world *worldT) addChar(name string, glyph rune, color tcell.Color, pos image.Point) ecs.Entity {
	ent := world.AddEntity(wcName | wcGlyph | wcFG | wcPosition | wcSolid | wcHP | wcStats)
	id := ent.ID()
	world.names[id] = name
	world.glyphs[id] = glyph
	world.fg[id] = color
	world.zval[id] = 200
	world.pos.Set(ent, pos)
	hp := &world.hp[id]
	hp.hp = hp.maxHP
	stats := &world.stats[id]
	stats.ap = stats.maxAP
	return ent
}

// neRoom is a lookup table for whether a cell is part of a room based
// on its immediate neighborhood (8 surrounding cells); it is keyed by
// a bit string corresponding to whether each cell contains an empty
// tile.
var neRoom [512]bool

// nePat represents a bit-packed key into neRoom.
type nePat uint16

// nePos definies the bit positions in nePat in row-major order.
var nePos = []image.Point{
	image.Pt(-1, -1), // 0x01
	image.Pt(0, -1),  // 0x02
	image.Pt(1, -1),  // 0x04
	image.Pt(-1, 0),  // 0x08
	image.Pt(1, 0),   // 0x10
	image.Pt(-1, 1),  // 0x20
	image.Pt(0, 1),   // 0x40
	image.Pt(1, 1),   // 0x80
}

func init() {
	stack := append(make([]nePat, 0, 4+5*4*3*2),
		// __#
		// __#
		// ###
		0x01|0x02|0x08,

		// #__
		// #__
		// ###
		0x02|0x04|0x10,

		// ###
		// #__
		// #__
		0x08|0x20|0x40,

		// ###
		// __#
		// __#
		0x10|0x40|0x80,
	)
	for len(stack) > 0 {
		i := len(stack) - 1
		pat := stack[i]
		stack = stack[:i]
		neRoom[pat] = true
		for b := nePat(1); b < 0x200; b <<= 1 {
			if npat := pat | b; npat != pat && !neRoom[npat] {
				stack = append(stack, npat)
			}
		}
	}
}
