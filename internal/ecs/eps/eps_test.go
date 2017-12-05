package eps_test

import (
	"sort"
	"testing"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/ecs/eps"
	"github.com/borkshop/bork/internal/point"
	"github.com/stretchr/testify/assert"
)

const (
	tpsPos ecs.ComponentType = 1 << iota
	tpsNom
)

type tps struct {
	ecs.Core
	pos eps.EPS
	nom []string
}

func (tps *tps) init() {
	tps.pos.Init(&tps.Core, tpsPos)
	tps.nom = []string{""}
	tps.Core.RegisterAllocator(tpsNom, tps.alloc)
	tps.Core.RegisterDestroyer(tpsNom, tps.destroyNom)
}

func (tps *tps) alloc(id ecs.EntityID, t ecs.ComponentType) {
	tps.nom = append(tps.nom, "")
}

func (tps *tps) destroyNom(id ecs.EntityID, t ecs.ComponentType) {
	tps.nom[id] = ""
}

func (tps *tps) nomed(nom string) ecs.Entity {
	for it := tps.Iter(tpsNom.All()); it.Next(); {
		if tps.nom[it.ID()] == nom {
			return it.Entity()
		}
	}
	return ecs.NilEntity
}

func (tps *tps) noms(ents []ecs.Entity) []string {
	if len(ents) == 0 {
		return nil
	}
	ss := make([]string, len(ents))
	for i, ent := range ents {
		ss[i] = tps.nom[ent.ID()]
	}
	return ss
}

func (tps *tps) load(xx ...interface{}) {
	for i := 0; i < len(xx); {
		ent := tps.AddEntity(tpsPos | tpsNom)
		tps.nom[ent.ID()] = xx[i].(string)
		i++
		x := xx[i].(int)
		i++
		y := xx[i].(int)
		i++
		tps.pos.Set(ent, point.Pt(x, y))
	}
}

func TestEPS(t *testing.T) {
	var tps tps
	tps.init()
	tps.load(
		"0", 0, 0,
		"a", -1, -1,
		"b", 1, -1,
		"c", 1, 1,
		"d", -1, 1,
	)

	t.Run("Get loaded", func(t *testing.T) {
		for i, tc := range []struct {
			nom  string
			x, y int
			ok   bool
		}{
			{"0", 0, 0, true},
			{"a", -1, -1, true},
			{"b", 1, -1, true},
			{"c", 1, 1, true},
			{"d", -1, 1, true},
			{"X", 0, 0, false},
		} {
			if pos, ok := tps.pos.Get(tps.nomed(tc.nom)); assert.Equal(t, tc.ok, ok, "[%v] ok", i) {
				assert.Equal(t, point.Pt(tc.x, tc.y), pos, "[%v] pos", i)
			}
		}
	})

	t.Run("At singles", func(t *testing.T) {
		for i, tc := range []struct {
			x, y int
			noms []string
		}{
			{-1, -1, []string{"a"}},
			{0, -1, nil},
			{1, -1, []string{"b"}},
			{-1, 0, nil},
			{0, 0, []string{"0"}},
			{1, 0, nil},
			{-1, 1, []string{"d"}},
			{0, 1, nil},
			{1, 1, []string{"c"}},
		} {
			ents := tps.pos.At(point.Pt(tc.x, tc.y))
			noms := tps.noms(ents)
			sort.Strings(noms)
			assert.Equal(t, tc.noms, noms, "[%v] noms", i)
		}
	})

	tps.pos.Set(tps.nomed("a"), point.Pt(1, 1))
	tps.pos.Set(tps.nomed("b"), point.Pt(-1, 1))

	t.Run("At moved", func(t *testing.T) {
		for i, tc := range []struct {
			x, y int
			noms []string
		}{
			{-1, -1, nil},
			{1, -1, nil},
			{-1, 1, []string{"b", "d"}},
			{1, 1, []string{"a", "c"}},
		} {
			ents := tps.pos.At(point.Pt(tc.x, tc.y))
			noms := tps.noms(ents)
			sort.Strings(noms)
			assert.Equal(t, tc.noms, noms, "[%v] noms", i)
		}
	})

	tps.nomed("c").Delete(tpsPos)
	tps.nomed("d").Destroy()

	t.Run("At deleted", func(t *testing.T) {
		for i, tc := range []struct {
			x, y int
			noms []string
		}{
			{-1, 1, []string{"b"}},
			{1, 1, []string{"a"}},
		} {
			ents := tps.pos.At(point.Pt(tc.x, tc.y))
			noms := tps.noms(ents)
			sort.Strings(noms)
			assert.Equal(t, tc.noms, noms, "[%v] noms", i)
		}
	})

	tps.load("e", 9, 9)
	tps.pos.Set(tps.nomed("a"), point.Pt(9, 9))

	t.Run("At re-use", func(t *testing.T) {
		for i, tc := range []struct {
			x, y int
			noms []string
		}{
			{9, 9, []string{"a", "e"}},
		} {
			ents := tps.pos.At(point.Pt(tc.x, tc.y))
			noms := tps.noms(ents)
			sort.Strings(noms)
			assert.Equal(t, tc.noms, noms, "[%v] noms", i)
		}
	})

}
