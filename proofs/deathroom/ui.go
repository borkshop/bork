package main

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/perf"
	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud"
	"github.com/borkshop/bork/internal/view/hud/prompt"
)

type actionItem interface {
	prompt.Runner
	label() string
}

type keyedActionItem interface {
	actionItem
	key() rune
}

type actionBar struct {
	prompt.Prompt
	items []actionItem
	sep   string
}

type labeldRunner struct {
	prompt.Runner
	lb string
}

func (lr labeldRunner) label() string { return lr.lb }

func labeled(run prompt.Runner, mess string, args ...interface{}) actionItem {
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	return labeldRunner{run, mess}
}

func (ab *actionBar) reset() {
	ab.Prompt = ab.Prompt.Unwind()
}

func (ab *actionBar) addAction(ai actionItem) {
	ab.setAction(len(ab.items), ai)
}

func (ab *actionBar) removeLabel(mess string) {
	for i := range ab.items {
		if ab.items[i].label() == mess {
			if j := i + 1; j < len(ab.items) {
				copy(ab.items[i:], ab.items[j:])
			}
			ab.items = ab.items[:len(ab.items)-1]
			return
		}
	}
}

func (ab *actionBar) replaceLabel(mess string, ai actionItem) {
	for i := range ab.items {
		if ab.items[i].label() == mess {
			ab.setAction(i, ai)
			return
		}
	}
	ab.addAction(ai)
}

func (ab *actionBar) setAction(i int, ai actionItem) {
	if i >= len(ab.items) {
		items := make([]actionItem, i+1)
		copy(items, ab.items)
		ab.items = items
	}

	ab.items[i] = ai

	ab.Prompt.Clear()
	for i, ai := range ab.items {
		if r := ab.rune(i); r != 0 {
			ab.AddAction(r, ai, ai.label())
		}
	}
}

func (ab actionBar) rune(i int) rune {
	ai := ab.items[i]
	if ai == nil {
		return 0
	}
	if kai, ok := ai.(keyedActionItem); ok {
		return kai.key()
	} else if i < 9 {
		return '1' + rune(i)
	}
	return 0
}

func (ab actionBar) label(i int) string {
	ai := ab.items[i]
	if r := ab.rune(i); r != 0 {
		return fmt.Sprintf("%s|%s", ai.label(), string(r))
	}
	return fmt.Sprintf("%s|Ø", ai.label())
}

func (ab *actionBar) RenderSize() (wanted, needed image.Point) {
	if ab.Prompt.Len() == 0 {
		return
	}
	if !ab.Prompt.IsRoot() {
		return ab.Prompt.RenderSize()
	}
	if len(ab.items) == 0 {
		return
	}

	i := 0
	wanted.X = utf8.RuneCountInString(ab.label(i))
	needed.X = wanted.X
	wanted.Y = len(ab.items)
	needed.Y = len(ab.items)
	i++

	nsep := utf8.RuneCountInString(ab.sep)
	for ; i < len(ab.items); i++ {
		n := utf8.RuneCountInString(ab.label(i))
		wanted.X += nsep
		wanted.X += n
		if n > needed.X {
			needed.X = n
		}
	}

	return wanted, needed
}

func (ab *actionBar) Render(d *display.Display) {
	if !ab.Prompt.IsRoot() {
		ab.Prompt.Render(d)
		return
	}

	// TODO: maybe use EITHER one row OR one column, not a mix (grid of action
	// items)

	x, y, i := d.Rect.Min.X, d.Rect.Min.Y, 0
	x += d.WriteString(x, y, nil, nil, ab.label(i))
	i++
	// TODO: missing seps
	for ; i < len(ab.items); i++ {
		lb := ab.label(i)
		if rem := d.Rect.Max.X - x; rem >= utf8.RuneCountInString(ab.sep)+utf8.RuneCountInString(lb) {
			x += d.WriteString(x, y, nil, nil, ab.sep)
			x += d.WriteString(x, y, nil, nil, lb)
		} else {
			y++
			x = d.WriteString(0, y, nil, nil, lb)
		}
	}
}

type worldItemAction struct {
	w         *world
	item, ent ecs.Entity
}

func (wia worldItemAction) addAction(pr *prompt.Prompt, ch rune) bool {
	name := wia.w.getName(wia.item, "unknown item")
	return pr.AddAction(ch, wia, name)
}

func (wia worldItemAction) RunPrompt(pr prompt.Prompt) (prompt.Prompt, bool) {
	if item := wia.w.items[wia.item.ID()]; item != nil {
		return item.interact(pr, wia.w, wia.item, wia.ent)
	}
	return pr.Unwind(), false
}

type ui struct {
	View *view.View

	hud.Logs
	perfDash perf.Dash
	prompt   prompt.Prompt
	bar      actionBar
}

type bodySummary struct {
	w   *world
	bo  *body
	ent ecs.Entity
	a   view.Align

	charge      int
	hp, maxHP   int
	armorParts  []string
	damageParts []string
	chargeParts []string
}

func makeBodySummary(w *world, ent ecs.Entity) view.Renderable {
	bs := bodySummary{
		w:   w,
		bo:  w.bodies[ent.ID()],
		ent: ent,
	}
	bs.build()
	return bs
}

func (bs *bodySummary) reset() {
	n := bs.bo.Len() + 1
	bs.charge = 0
	bs.armorParts = nstrings(1, n, bs.armorParts)
	bs.damageParts = nstrings(0, n, bs.damageParts)
	bs.chargeParts = nstrings(1, 1, bs.chargeParts)
}

func (bs *bodySummary) build() {
	bs.reset()

	bs.charge = bs.w.getCharge(bs.ent)

	for it := bs.bo.Iter(bcHPart.All()); it.Next(); {
		bs.hp += bs.bo.hp[it.ID()]
		bs.maxHP += bs.bo.maxHP[it.ID()]
	}

	headArmor := 0
	for it := bs.bo.Iter((bcPart | bcHead).All()); it.Next(); {
		headArmor += bs.bo.armor[it.ID()]
	}

	torsoArmor := 0
	for it := bs.bo.Iter((bcPart | bcTorso).All()); it.Next(); {
		torsoArmor += bs.bo.armor[it.ID()]
	}

	for _, part := range bs.bo.rel.Leaves(brControl.All(), nil) {
		bs.damageParts = append(bs.damageParts, fmt.Sprintf(
			"%v+%v",
			bs.bo.PartAbbr(part),
			bs.bo.dmg[part.ID()],
		))
	}
	sort.Strings(bs.damageParts)

	bs.armorParts[0] = fmt.Sprintf("Armor: %v %v", headArmor, torsoArmor)
	bs.chargeParts[0] = fmt.Sprintf("Charge: %v", bs.charge)
}

func (bs bodySummary) RenderSize() (wanted, needed image.Point) {
	needed.Y = 5 + 1
	needed.X = moremath.MaxInt(
		7,
		stringsWidth(" ", bs.chargeParts),
	)

	for i := 0; i < len(bs.damageParts); {
		j := i + 2
		if j > len(bs.damageParts) {
			j = len(bs.damageParts)
		}
		needed.X = moremath.MaxInt(needed.X, stringsWidth(" ", bs.damageParts[i:j]))
		needed.Y++
		i = j
	}

	needed.Y++ // XXX why

	return needed, needed
}

func (bs bodySummary) partHPColor(part ecs.Entity) color.RGBA {
	if part == ecs.NilEntity {
		return itemColors[0]
	}
	id := bs.bo.Deref(part)
	if !part.Type().HasAll(bcPart) {
		return itemColors[0]
	}
	hp := bs.bo.hp[id]
	maxHP := bs.bo.maxHP[id]
	return safeColorsIX(itemColors, 1+(len(itemColors)-2)*hp/maxHP)
}

func (bs bodySummary) Render(d *display.Display) {
	// TODO: bodyHPColors ?
	// TODO: support scaling body with grafting

	w := d.Rect.Dx()
	y := 0
	mess := fmt.Sprintf("%.0f%%", float64(bs.hp)/float64(bs.maxHP)*100)
	d.WriteString((w-len(mess))/2, y, nil, nil, mess)
	y++

	//  0123456
	// 0  _O_
	// 1 / | \
	// 2 = | =
	// 3  / \
	// 4_/   \_

	xo := (w - 7) / 2
	for _, pt := range []struct {
		x, y int
		s    string
		t    ecs.ComponentType
	}{
		{xo + 2, y + 0, "_", bcPart | bcUpperArm | bcLeft},
		{xo + 3, y + 0, "O", bcPart | bcHead},
		{xo + 4, y + 0, "_", bcPart | bcUpperArm | bcRight},

		{xo + 1, y + 1, "/", bcPart | bcForeArm | bcLeft},
		{xo + 3, y + 1, "|", bcPart | bcTorso},
		{xo + 5, y + 1, "\\", bcPart | bcForeArm | bcRight},

		{xo + 1, y + 2, "=", bcPart | bcHand | bcLeft},
		{xo + 3, y + 2, "|", bcPart | bcTorso},
		{xo + 5, y + 2, "=", bcPart | bcHand | bcRight},

		{xo + 2, y + 3, "/", bcPart | bcThigh | bcLeft},
		{xo + 4, y + 3, "\\", bcPart | bcThigh | bcRight},

		{xo + 0, y + 4, "_", bcPart | bcFoot | bcLeft},
		{xo + 1, y + 4, "/", bcPart | bcCalf | bcLeft},
		{xo + 5, y + 4, "\\", bcPart | bcCalf | bcRight},
		{xo + 6, y + 4, "_", bcPart | bcFoot | bcRight},
	} {
		if it := bs.bo.Iter(pt.t.All()); it.Next() {
			c := bs.partHPColor(it.Entity())
			d.Set(pt.x, pt.y, pt.s, c, nil)
		}
	}

	y += 5

	d.WriteString(0, y, nil, nil, strings.Join(bs.chargeParts, " "))
	y++

	for i := 0; i < len(bs.damageParts); {
		j := i + 2
		if j > len(bs.damageParts) {
			j = len(bs.damageParts)
		}
		d.WriteString(0, y, nil, nil, strings.Join(bs.damageParts[i:j], " "))
		y++
		i = j
	}
}

func (ui *ui) init(v *view.View, perf *perf.Perf) {
	ui.View = v
	ui.Logs.Init(1000)
	ui.Logs.Align = view.AlignLeft | view.AlignTop | view.AlignHFlush
	ui.bar.sep = " "
	ui.perfDash.Perf = perf
}

func (ui *ui) handle(cmd interface{}) (proc, handled bool, err error) {
	if ui.perfDash.HandleInput(cmd) {
		return false, true, nil
	}

	defer func() {
		if !handled {
			ui.prompt.Clear()
			ui.bar.reset()
		}
	}()

	switch c := cmd.(type) {
	case rune:
		switch c {
		case '':
			return false, true, view.ErrStop
		}
	}

	if handled, canceled, prompting := ui.prompt.Handle(cmd); handled {
		proc = !prompting && !canceled
		if proc {
			ui.prompt.Clear()
		}
		return proc, true, nil
	}

	if handled, canceled, prompting := ui.bar.Handle(cmd); handled {
		return !prompting && !canceled, true, nil
	}

	return false, false, nil
}

func (w *world) HandleInput(cmd interface{}) (rerr error) {
	proc, handled, err := w.ui.handle(cmd)
	defer func() {
		if rerr != nil {
			_ = w.perf.Close()
		} else if err := w.perf.Err(); err != nil {
			rerr = err
		}
	}()

	if err != nil {
		return err
	}

	if w.over {
		return nil
	}

	player := w.findPlayer()
	if player != ecs.NilEntity && w.ui.bar.IsRoot() {
		defer func() {
			if rerr != nil {
				return
			}
			if itemPrompt, haveItemsHere := w.itemPrompt(w.prompt, player); haveItemsHere {
				w.ui.bar.replaceLabel("Inspect", labeled(itemPrompt, "Inspect"))
			} else {
				w.ui.bar.removeLabel("Inspect")
			}
		}()
	}

	if !handled {
		switch c := cmd.(type) {
		case input.RelativeMove:
			for it := w.Iter(playMoveMask.All()); it.Next(); {
				w.moves.AddPendingMove(it.Entity(), c.Point, 1, maxRestingCharge)
			}
			proc, handled = true, true

		case rune:
			switch c {
			case '.':
				for it := w.Iter(playMoveMask.All()); it.Next(); {
					w.moves.AddPendingMove(it.Entity(), image.ZP, 1, maxRestingCharge)
				}
				proc, handled = true, true

			case ',':
				if player != ecs.NilEntity {
					if itemPrompt, haveItemsHere := w.itemPrompt(w.prompt, player); haveItemsHere {
						w.prompt, _ = itemPrompt.RunPrompt(w.prompt.Unwind())
					}
				}
				proc, handled = false, true

			case '_':
				if player != ecs.NilEntity {
					if player.Type().HasAll(wcSolid) {
						player.Delete(wcSolid)
						w.Glyphs[player.ID()] = "~"
					} else {
						player.Add(wcSolid)
						w.Glyphs[player.ID()] = "X"
					}
				}
				proc, handled = true, true

			}
		}
	}

	// default to resting
	if !handled {
		proc = true
	}

	if proc {
		w.Process()
	}

	return nil
}

func (w *world) Render(d *display.Display) error {
	hud := hud.HUD{
		Logs:  w.ui.Logs,
		World: w.renderViewport(d.Rect),
	}

	hud.HeaderF(">%v souls v %v demons",
		w.Iter(wcSoul.All()).Count(),
		w.Iter(wcAI.All()).Count())

	hud.AddRenderable(&w.ui.bar, view.AlignLeft|view.AlignBottom)
	hud.AddRenderable(&w.ui.prompt, view.AlignLeft|view.AlignBottom)

	hud.AddRenderable(&w.ui.perfDash, view.AlignRight|view.AlignBottom)

	for it := w.Iter((wcSoul | wcBody).All()); it.Next(); {
		hud.AddRenderable(makeBodySummary(w, it.Entity()),
			view.AlignBottom|view.AlignRight|view.AlignHFlush)
	}

	hud.Render(d)
	return nil
}

func (w *world) makeViewport(within image.Rectangle) (*display.Display, []uint8) {
	// collect world extent, and compute a viewport focus position
	var (
		bbox  image.Rectangle
		focus image.Point
	)
	for it := w.Iter(renderMask.All()); it.Next(); {
		pos, _ := w.pos.Get(it.Entity())
		if it.Type().HasAll(wcSoul) {
			// TODO: centroid between all souls would be a way to move beyond
			// "last wins"
			focus = pos
		}
		bbox = point.ExpandTo(bbox, pos)
	}

	// center clamped box around focus
	if dx := bbox.Dx() - within.Dx(); dx > 0 {
		bbox.Max.X -= dx
	}
	if dy := bbox.Dy() - within.Dy(); dy > 0 {
		bbox.Max.Y -= dy
	}
	ctr := bbox.Min.Add(bbox.Size().Div(2))
	bbox = bbox.Add(ctr.Sub(focus))

	// TODO: re-use
	dis := display.New(bbox)
	zVals := make([]uint8, len(dis.Text.Strings))

	return dis, zVals
}

func (w *world) renderGlyphEntity(t ecs.ComponentType, id ecs.EntityID) (fg color.RGBA, zVal uint8) {
	// TODO pre-compute color when HP updates?

	if t.HasAll(wcSoul) {
		if t.HasAll(wcBody) {
			hp, maxHP := w.bodies[id].HPRange()
			return safeColorsIX(soulColors, 1+(len(soulColors)-2)*hp/maxHP), 255
		}
		return soulColors[0], 127
	}

	if t.HasAll(wcAI) {
		if t.HasAll(wcBody) {
			hp, maxHP := w.bodies[id].HPRange()
			return safeColorsIX(aiColors, 1+(len(aiColors)-2)*hp/maxHP), 254
		}
		return aiColors[0], 126
	}

	if t.HasAll(wcItem) {
		if dur, ok := w.items[id].(durableItem); ok {
			if hp, maxHP := dur.HPRange(); maxHP > 0 {
				return safeColorsIX(itemColors, (len(itemColors)-1)*hp/maxHP), 10
			}
			return itemColors[0], 10
		}
		return itemColors[len(itemColors)-1], 10
	}

	if t.HasAll(wcFG) {
		fg = w.FG[id]
	}
	return fg, 2
}

func (w *world) renderViewport(within image.Rectangle) *display.Display {
	dis, zVals := w.makeViewport(within)
	// TODO: use an eps range query
	for it := w.Iter(wcPosition.All(), (wcGlyph | wcBG).Any()); it.Next(); {
		pos, _ := w.pos.Get(it.Entity())
		if dis.Rect.Intersect(image.Rectangle{pos, pos}) == image.ZR {
			continue
		}

		var (
			t   = ""
			fg  = color.RGBA{0, 0, 0, 0}
			bg  = color.RGBA{0, 0, 0, 0}
			any = false
		)

		if it.Type().HasAll(wcGlyph) {
			c, zVal := w.renderGlyphEntity(it.Type(), it.ID())
			s := w.Glyphs[it.ID()]
			if s != "" {
				if zi := dis.Text.StringsOffset(pos.X, pos.Y); zVal >= zVals[zi] {
					fg = c
					t = s
					zVals[zi] = zVal
					any = true
				}
			}
		}

		if it.Type().HasAll(wcBG) {
			bg = w.BG[it.ID()]
			any = true
		}

		if any {
			dis.MergeRGBA(pos.X, pos.Y, t, fg, bg)
		}
	}

	return dis
}

func (w *world) itemPrompt(pr prompt.Prompt, ent ecs.Entity) (prompt.Prompt, bool) {
	// TODO: once we have a proper spatial index, stop relying on
	// collision relations for this
	pos, ok := w.pos.Get(ent)
	if !ok {
		return pr, false
	}
	prompting := false
	for _, pent := range w.pos.At(pos) {
		if !pent.Type().HasAll(wcItem) {
			continue
		}
		if !prompting {
			pr = pr.Sub("Items Here")
			prompting = true
		}
		i := pr.Len()
		worldItemAction{w, pent, ent}.addAction(&pr, '1'+rune(i))
		if i >= 8 {
			break
		}
	}
	return pr, prompting
}

func (bo *body) interact(pr prompt.Prompt, w *world, item, ent ecs.Entity) (prompt.Prompt, bool) {
	if !ent.Type().HasAll(wcBody) {
		if ent.Type().HasAll(wcSoul) {
			w.log("you have no body!")
		}
		return pr, false
	}

	pr = pr.Sub(w.getName(item, "unknown item"))

	for i, it := 0, bo.Iter(bcPart.All()); i < 9 && it.Next(); i++ {
		part := it.Entity()
		rem := bodyRemains{w, bo, part, item, ent}
		// TODO: inspect menu when more than just scavengable

		// any part can be scavenged
		pr.AddAction('1'+rune(i), prompt.Func(rem.scavenge), rem.describeScavenge())
	}

	return pr, true
}

func safeColorsIX(colors []color.RGBA, i int) color.RGBA {
	if i < 0 {
		return colors[1]
	}
	if i >= len(colors) {
		return colors[len(colors)-1]
	}
	return colors[i]
}

func nstrings(n, m int, ss []string) []string {
	if m > cap(ss) {
		return make([]string, n, m)
	}
	return ss[:n]
}

func stringsWidth(sep string, parts []string) int {
	n := (len(parts) - 1) + utf8.RuneCountInString(sep)
	for _, part := range parts {
		n += utf8.RuneCountInString(part)
	}
	return n
}
