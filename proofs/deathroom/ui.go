package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/perf"
	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud"
	"github.com/borkshop/bork/internal/view/hud/prompt"
	termbox "github.com/nsf/termbox-go"
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

func (ab *actionBar) RenderSize() (wanted, needed point.Point) {
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

func (ab *actionBar) Render(g view.Grid) {
	if !ab.Prompt.IsRoot() {
		ab.Prompt.Render(g)
		return
	}

	// TODO: maybe use EITHER one row OR one column, not a mix (grid of action
	// items)

	x, y, i := 0, 0, 0
	x += g.WriteString(x, y, ab.label(i))
	i++
	// TODO: missing seps
	for ; i < len(ab.items); i++ {
		lb := ab.label(i)
		if rem := g.Size.X - x; rem >= utf8.RuneCountInString(ab.sep)+utf8.RuneCountInString(lb) {
			x += g.WriteString(x, y, ab.sep)
			x += g.WriteString(x, y, lb)
		} else {
			y++
			x = g.WriteString(0, y, lb)
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

	for it := bs.bo.Iter(ecs.All(bcPart | bcHP)); it.Next(); {
		bs.hp += bs.bo.hp[it.ID()]
		bs.maxHP += bs.bo.maxHP[it.ID()]
	}

	headArmor := 0
	for it := bs.bo.Iter(ecs.All(bcPart | bcHead)); it.Next(); {
		headArmor += bs.bo.armor[it.ID()]
	}

	torsoArmor := 0
	for it := bs.bo.Iter(ecs.All(bcPart | bcTorso)); it.Next(); {
		torsoArmor += bs.bo.armor[it.ID()]
	}

	for _, part := range bs.bo.rel.Leaves(ecs.AllRel(brControl), nil) {
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

func (bs bodySummary) RenderSize() (wanted, needed point.Point) {
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

func (bs bodySummary) partHPColor(part ecs.Entity) termbox.Attribute {
	if part == ecs.NilEntity {
		return itemColors[0]
	}
	id := bs.bo.Deref(part)
	if !part.Type().All(bcPart) {
		return itemColors[0]
	}
	hp := bs.bo.hp[id]
	maxHP := bs.bo.maxHP[id]
	return safeColorsIX(itemColors, 1+(len(itemColors)-2)*hp/maxHP)
}

func (bs bodySummary) Render(g view.Grid) {
	// TODO: bodyHPColors ?
	// TODO: support scaling body with grafting

	w := g.Size.X
	y := 0
	mess := fmt.Sprintf("%.0f%%", float64(bs.hp)/float64(bs.maxHP)*100)
	g.WriteString((w-len(mess))/2, y, mess)
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
		ch   rune
		t    ecs.ComponentType
	}{
		{xo + 2, y + 0, '_', bcUpperArm | bcLeft},
		{xo + 3, y + 0, 'O', bcHead},
		{xo + 4, y + 0, '_', bcUpperArm | bcRight},

		{xo + 1, y + 1, '/', bcForeArm | bcLeft},
		{xo + 3, y + 1, '|', bcTorso},
		{xo + 5, y + 1, '\\', bcForeArm | bcRight},

		{xo + 1, y + 2, '=', bcHand | bcLeft},
		{xo + 3, y + 2, '|', bcTorso},
		{xo + 5, y + 2, '=', bcHand | bcRight},

		{xo + 2, y + 3, '/', bcThigh | bcLeft},
		{xo + 4, y + 3, '\\', bcThigh | bcRight},

		{xo + 0, y + 4, '_', bcFoot | bcLeft},
		{xo + 1, y + 4, '/', bcCalf | bcLeft},
		{xo + 5, y + 4, '\\', bcCalf | bcRight},
		{xo + 6, y + 4, '_', bcFoot | bcRight},
	} {
		it := bs.bo.Iter(ecs.All(bcPart | pt.t))
		if it.Next() {
			g.Set(pt.x, pt.y, pt.ch, bs.partHPColor(it.Entity()), 0)
		}
	}

	y += 5

	g.WriteString(0, y, strings.Join(bs.chargeParts, " "))
	y++

	for i := 0; i < len(bs.damageParts); {
		j := i + 2
		if j > len(bs.damageParts) {
			j = len(bs.damageParts)
		}
		g.WriteString(0, y, strings.Join(bs.damageParts[i:j], " "))
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

func (ui *ui) handle(k view.KeyEvent) (proc, handled bool, err error) {
	if ui.perfDash.HandleKey(k) {
		return false, true, nil
	}

	defer func() {
		if !handled {
			ui.prompt.Clear()
			ui.bar.reset()
		}
	}()

	if k.Key == termbox.KeyEsc {
		return false, true, view.ErrStop
	}

	if handled, canceled, prompting := ui.prompt.Handle(k); handled {
		proc = !prompting && !canceled
		if proc {
			ui.prompt.Clear()
		}
		return proc, true, nil
	}

	if handled, canceled, prompting := ui.bar.Handle(k); handled {
		return !prompting && !canceled, true, nil
	}

	return false, false, nil
}

func (w *world) HandleKey(k view.KeyEvent) (rerr error) {
	proc, handled, err := w.ui.handle(k)
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

	// special keys
	if !handled {
		switch k.Ch {
		case ',':
			if player != ecs.NilEntity {
				if itemPrompt, haveItemsHere := w.itemPrompt(w.prompt, player); haveItemsHere {
					w.prompt, _ = itemPrompt.RunPrompt(w.prompt.Unwind())
				}
			}
			proc, handled = false, true
		case '_':
			if player != ecs.NilEntity {
				if player.Type().All(wcCollide) {
					player.Delete(wcCollide)
					w.Glyphs[player.ID()] = '~'
				} else {
					player.Add(wcCollide)
					w.Glyphs[player.ID()] = 'X'
				}
			}
			proc, handled = true, true
		}
	}

	// parse player move
	if !handled {
		if move, ok := parseMove(k); ok {
			for it := w.Iter(ecs.All(playMoveMask)); it.Next(); {
				w.addPendingMove(it.Entity(), move)
			}
			proc, handled = true, true
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

func parseMove(k view.KeyEvent) (point.Point, bool) {
	switch k.Key {
	case termbox.KeyArrowDown:
		return point.Point{X: 0, Y: 1}, true
	case termbox.KeyArrowUp:
		return point.Point{X: 0, Y: -1}, true
	case termbox.KeyArrowLeft:
		return point.Point{X: -1, Y: 0}, true
	case termbox.KeyArrowRight:
		return point.Point{X: 1, Y: 0}, true
	}
	switch k.Ch {
	case 'y':
		return point.Point{X: -1, Y: -1}, true
	case 'u':
		return point.Point{X: 1, Y: -1}, true
	case 'n':
		return point.Point{X: 1, Y: 1}, true
	case 'b':
		return point.Point{X: -1, Y: 1}, true
	case 'h':
		return point.Point{X: -1, Y: 0}, true
	case 'j':
		return point.Point{X: 0, Y: 1}, true
	case 'k':
		return point.Point{X: 0, Y: -1}, true
	case 'l':
		return point.Point{X: 1, Y: 0}, true
	case '.':
		return point.Zero, true
	}
	return point.Zero, false
}

func (w *world) Render(termGrid view.Grid) error {
	hud := hud.HUD{
		Logs:  w.ui.Logs,
		World: w.renderViewport(termGrid.Size),
	}

	hud.HeaderF(">%v souls v %v demons", w.Iter(ecs.All(wcSoul)).Count(), w.Iter(ecs.All(wcAI)).Count())

	hud.AddRenderable(&w.ui.bar, view.AlignLeft|view.AlignBottom)
	hud.AddRenderable(&w.ui.prompt, view.AlignLeft|view.AlignBottom)

	hud.AddRenderable(&w.ui.perfDash, view.AlignRight|view.AlignBottom)

	for it := w.Iter(ecs.All(wcSoul | wcBody)); it.Next(); {
		hud.AddRenderable(makeBodySummary(w, it.Entity()),
			view.AlignBottom|view.AlignRight|view.AlignHFlush)
	}

	hud.Render(termGrid)
	return nil
}

func (w *world) renderViewport(max point.Point) view.Grid {
	// collect world extent, and compute a viewport focus position
	var (
		bbox  point.Box
		focus point.Point
	)
	for it := w.Iter(ecs.All(renderMask)); it.Next(); {
		pos, _ := w.pos.Get(it.Entity())
		if it.Type().All(wcSoul) {
			// TODO: centroid between all souls would be a way to move beyond
			// "last wins"
			focus = pos
		}
		bbox = bbox.ExpandTo(pos)
	}

	// center clamped grid around focus
	offset := bbox.TopLeft.Add(bbox.Size().Div(2)).Sub(focus)
	ofbox := bbox.Add(offset)
	if ofbox.TopLeft.X < 0 {
		offset.X -= ofbox.TopLeft.X
	}
	if ofbox.TopLeft.Y < 0 {
		offset.Y -= ofbox.TopLeft.Y
	}

	// TODO: re-use
	grid := view.MakeGrid(ofbox.Size().Min(max))
	zVals := make([]uint8, len(grid.Data))

	// TODO: use an pos range query
	for it := w.Iter(ecs.Clause(wcPosition, wcGlyph|wcBG)); it.Next(); {
		pos, _ := w.pos.Get(it.Entity())
		pos = pos.Add(offset)
		gi := pos.Y*grid.Size.X + pos.X
		if gi < 0 || gi >= len(grid.Data) {
			// TODO: debug
			continue
		}

		if it.Type().All(wcGlyph) {
			var fg termbox.Attribute
			var zVal uint8

			zVal = 1

			// TODO: move to hp update
			if it.Type().All(wcBody) && it.Type().Any(wcSoul|wcAI) {
				zVal = 255
				hp, maxHP := w.bodies[it.ID()].HPRange()
				if !it.Type().All(wcSoul) {
					zVal--
					fg = safeColorsIX(aiColors, 1+(len(aiColors)-2)*hp/maxHP)
				} else {
					fg = safeColorsIX(soulColors, 1+(len(soulColors)-2)*hp/maxHP)
				}
			} else if it.Type().All(wcSoul) {
				zVal = 127
				fg = soulColors[0]
			} else if it.Type().All(wcAI) {
				zVal = 126
				fg = aiColors[0]
			} else if it.Type().All(wcItem) {
				zVal = 10
				fg = itemColors[len(itemColors)-1]
				if dur, ok := w.items[it.ID()].(durableItem); ok {
					fg = itemColors[0]
					if hp, maxHP := dur.HPRange(); maxHP > 0 {
						fg = safeColorsIX(itemColors, (len(itemColors)-1)*hp/maxHP)
					}
				}
			} else {
				zVal = 2
				if it.Type().All(wcFG) {
					fg = w.FG[it.ID()]
				}
			}

			if ch := w.Glyphs[it.ID()]; zVal >= zVals[gi] && ch != 0 {
				grid.Data[gi].Ch = ch
				zVals[gi] = zVal
				if fg != 0 {
					grid.Data[gi].Fg = fg + 1
				} else {
					grid.Data[gi].Fg = 0
				}
			} else {
				continue
			}
		}

		if it.Type().All(wcBG) {
			if bg := w.BG[it.ID()]; bg != 0 {
				grid.Data[gi].Bg = bg + 1
			}
		}
	}

	return grid
}

func (w *world) itemPrompt(pr prompt.Prompt, ent ecs.Entity) (prompt.Prompt, bool) {
	// TODO: once we have a proper spatial index, stop relying on
	// collision relations for this
	prompting := false
	for i, cur := 0, w.moves.Cursor(
		ecs.RelClause(mrCollide, mrItem),
		func(r ecs.RelationType, rel, a, b ecs.Entity) bool { return a == ent },
	); i < 9 && cur.Scan(); i++ {
		if !prompting {
			pr = pr.Sub("Items Here")
			prompting = true
		}
		worldItemAction{w, cur.B(), ent}.addAction(&pr, '1'+rune(i))
	}
	return pr, prompting
}

func (bo *body) interact(pr prompt.Prompt, w *world, item, ent ecs.Entity) (prompt.Prompt, bool) {
	if !ent.Type().All(wcBody) {
		if ent.Type().All(wcSoul) {
			w.log("you have no body!")
		}
		return pr, false
	}

	pr = pr.Sub(w.getName(item, "unknown item"))

	for i, it := 0, bo.Iter(ecs.All(bcPart)); i < 9 && it.Next(); i++ {
		part := it.Entity()
		rem := bodyRemains{w, bo, part, item, ent}
		// TODO: inspect menu when more than just scavengable

		// any part can be scavenged
		pr.AddAction('1'+rune(i), prompt.Func(rem.scavenge), rem.describeScavenge())
	}

	return pr, true
}

func safeColorsIX(colors []termbox.Attribute, i int) termbox.Attribute {
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
