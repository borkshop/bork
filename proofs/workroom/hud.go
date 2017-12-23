package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
)

type hudT struct {
	views.Panel // TODO what good is this; maybe just boxlayout

	modal views.Widget

	title  *views.TextBar
	keybar *keybar
	status *views.SimpleStyledTextBar
	view   *worldView
}

type keybar struct {
	*views.SimpleStyledText
	actions map[rune]keybarAction
	prior   map[rune]keybarAction
}

type keybarAction struct {
	l string
	f func()
}

func newKeybar() *keybar {
	kb := &keybar{}
	kb.SimpleStyledText = views.NewSimpleStyledText()
	kb.actions = make(map[rune]keybarAction)
	return kb
}

func (kb *keybar) addAction(k rune, label string, f func()) {
	if _, def := kb.actions[k]; def {
		panic(fmt.Sprintf("duplicate action %q", k))
	}
	kb.actions[k] = keybarAction{label, f}
	kb.refresh()
}

func (kb *keybar) refresh() {
	parts := make([]string, 0, len(kb.actions))
	for k, a := range kb.actions {
		parts = append(parts, fmt.Sprintf("%%S[%s]%%A%s%%N", string(k), a.l))
	}
	sort.Strings(parts)
	kb.SetMarkup(strings.Join(parts, "  "))
}

func (kb *keybar) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			k := ev.Rune()
			a, def := kb.actions[k]
			if !def {
				if unicode.IsLower(k) {
					k = unicode.ToUpper(k)
				} else {
					k = unicode.ToLower(k)
				}
				a, def = kb.actions[k]
			}
			if def {
				a.f()
				return true
			}
		}
	}
	return false
}

func (hud *hudT) init() {
	hud.title = views.NewTextBar()
	hud.title.SetCenter("Welcome to Wørkrüm", tcell.StyleDefault)

	hud.keybar = newKeybar()
	hud.keybar.RegisterStyle('N', tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite))
	hud.keybar.RegisterStyle('A', tcell.StyleDefault.
		Background(tcell.ColorDarkBlue).
		Foreground(tcell.ColorSlateBlue))
	hud.keybar.RegisterStyle('S', tcell.StyleDefault.
		Background(tcell.ColorSlateBlue).
		Foreground(tcell.ColorDarkBlue))

	hud.status = views.NewSimpleStyledTextBar()
	// hud.status.SetStyle(tcell.StyleDefault.
	// 	Background(tcell.ColorBlue).
	// 	Foreground(tcell.ColorYellow))
	// hud.status.RegisterLeftStyle('N', tcell.StyleDefault.
	// 	Background(tcell.ColorYellow).
	// 	Foreground(tcell.ColorBlack))

	hud.view = newView(world)

	hud.SetMenu(hud.status)
	hud.SetTitle(hud.title)
	hud.SetStatus(hud.keybar)
	hud.SetContent(hud.view)
}

func (hud *hudT) showModal(wid views.Widget) {
	if hud.keybar.prior == nil {
		hud.keybar.prior = hud.keybar.actions
	}
	hud.modal = wid
	hud.SetContent(wid)
	hud.keybar.actions = make(map[rune]keybarAction)
	hud.keybar.addAction('Q', "Resume Game", hud.hideModal)
	// TODO support adding other modal actions
}

func (hud *hudT) hideModal() {
	if hud.keybar.prior != nil {
		hud.keybar.actions = hud.keybar.prior
		hud.keybar.prior = nil
		hud.keybar.refresh()
	}
	hud.modal = nil
	hud.SetContent(hud.view)
}

func (hud *hudT) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlL:
			app.Refresh()
			return true
		case tcell.KeyCtrlC:
			app.Quit()
			return true
		}
	}

	if hud.Panel.HandleEvent(ev) {
		return true
	}
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			hud.status.SetLeft(fmt.Sprintf("?rune %q", ev.Rune()))
		default:
			hud.status.SetLeft(fmt.Sprintf("?key %v", ev.Key()))
		}
	default:
		hud.status.SetLeft(fmt.Sprintf("?ev %T", ev))
	}
	return false
}
