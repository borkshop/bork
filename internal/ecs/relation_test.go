package ecs_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/ecs"
)

const (
	srFoo ecs.ComponentType = 1 << iota
	srBar
)

func setupRelTest(aFlags, bFlags ecs.RelationFlags) (a, b *stuff, rel *ecs.Relation) {
	a = newStuff()
	a1 := a.addData(3)
	a2 := a.addData(6)
	a3 := a.addData(9)
	a4 := a.addData(12)
	a5 := a.addData(15, 30, 45, 60)
	a6 := a.addData(18)
	a7 := a.addData(21)
	_ = a.addData(24) // a8

	b = newStuff()
	b1 := b.addData(5)
	b2 := b.addData(10, 20, 30, 40)
	b3 := b.addData(15)
	b4 := b.addData(20, 40, 60, 80)
	b5 := b.addData(25)
	b6 := b.addData(30, 60, 90, 120)
	b7 := b.addData(35)
	_ = b.addData(40, 80, 120, 160) // b8

	rel = ecs.NewRelation(&a.Core, aFlags, &b.Core, bFlags)

	rel.Upsert(nil, func(uc *ecs.UpsertCursor) {

		uc.Create(srFoo, a1, b2)
		uc.Create(srFoo, a1, b3)
		uc.Create(srFoo, a2, b4)
		uc.Create(srFoo, a2, b5)
		uc.Create(srFoo, a3, b6)
		uc.Create(srFoo, a3, b7)

		uc.Create(srFoo|srBar, a2, b1)
		uc.Create(srFoo|srBar, a3, b1)
		uc.Create(srFoo|srBar, a4, b2)
		uc.Create(srFoo|srBar, a5, b2)
		uc.Create(srFoo|srBar, a6, b3)
		uc.Create(srFoo|srBar, a7, b3)

	})

	return a, b, rel
}

type testCases []testCase
type testCase struct {
	name string
	run  func(t *testing.T)
}

func (tcs testCases) run(t *testing.T) {
	for _, tc := range tcs {
		t.Run(tc.name, tc.run)
	}
}

func collectAIDs(cur ecs.Cursor) (ids []ecs.EntityID) {
	for cur.Scan() {
		ids = append(ids, cur.A().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func collectBIDs(cur ecs.Cursor) (ids []ecs.EntityID) {
	for cur.Scan() {
		ids = append(ids, cur.B().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func TestRelation_Select(t *testing.T) {
	testCases{

		{"TypeClause", func(t *testing.T) {
			_, _, r := setupRelTest(0, 0)
			assert.Equal(t, 12, r.Select(srFoo.All()).Count())
			assert.Equal(t, 6, r.Select(srBar.All()).Count())
		}},

		{"Lookup A", func(t *testing.T) {
			_, _, r := setupRelTest(0, 0)
			cur := r.Select(ecs.InA(1, 5, 7, 9))
			assert.Equal(t, 4, cur.Count())
			assert.Equal(t, []ecs.EntityID{2, 2, 3, 3}, collectBIDs(cur))
		}},

		{"Lookup B", func(t *testing.T) {
			_, _, r := setupRelTest(0, 0)
			cur := r.Select(ecs.InB(6, 7, 8, 9))
			assert.Equal(t, 2, cur.Count())
			assert.Equal(t, []ecs.EntityID{3, 3}, collectAIDs(cur))
		}},

		{"filter", func(t *testing.T) {
			a, _, r := setupRelTest(0, 0)
			fil := func(cur ecs.Cursor) bool {
				return a.d1[cur.A().ID()] < 10
			}
			cur := r.Select(ecs.Filter(fil))
			assert.Equal(t, 8, cur.Count())
			assert.Equal(t, []ecs.EntityID{1, 1, 2, 2, 2, 3, 3, 3}, collectAIDs(cur))
			cur = r.Select(ecs.Filter(fil))
			assert.Equal(t, []ecs.EntityID{1, 1, 2, 3, 4, 5, 6, 7}, collectBIDs(cur))
		}},
	}.run(t)
}

func TestRelation_destruction(t *testing.T) {
	testCases{
		{"clear A", func(t *testing.T) {
			a, b, r := setupRelTest(0, 0)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			a.Clear()
			assert.True(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
		}},

		{"clear B", func(t *testing.T) {
			a, b, r := setupRelTest(0, 0)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			b.Clear()
			assert.False(t, a.Empty())
			assert.True(t, b.Empty())
			assert.True(t, r.Empty())
		}},

		{"clear rels", func(t *testing.T) {
			a, b, r := setupRelTest(0, 0)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			r.Clear()
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
		}},

		{"A cascades", func(t *testing.T) {
			a, b, r := setupRelTest(ecs.RelationCascadeDestroy, 0)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			assert.Equal(t, 8, a.Len())
			assert.Equal(t, 8, b.Len())

			b.Ref(1).Destroy()
			assert.Equal(t, 6, a.Len())
			assert.Equal(t, ecs.NoType, a.Ref(2).Type())
			assert.Equal(t, ecs.NoType, a.Ref(3).Type())

			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())

			b.Clear()
			assert.False(t, a.Empty())
			assert.True(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 1, a.Len())
			assert.Equal(t, 0, b.Len())

			a, b, r = setupRelTest(ecs.RelationCascadeDestroy, 0)
			r.Clear()
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 1, a.Len())
			assert.Equal(t, 8, b.Len())
		}},

		{"B cascades", func(t *testing.T) {
			a, b, r := setupRelTest(0, ecs.RelationCascadeDestroy)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			assert.Equal(t, 8, a.Len())
			assert.Equal(t, 8, b.Len())

			a.Ref(1).Destroy()
			assert.Equal(t, 6, b.Len())
			assert.Equal(t, ecs.NoType, b.Ref(2).Type())
			assert.Equal(t, ecs.NoType, b.Ref(3).Type())

			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())

			a.Clear()
			assert.True(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 0, a.Len())
			assert.Equal(t, 1, b.Len())

			a, b, r = setupRelTest(0, ecs.RelationCascadeDestroy)
			r.Clear()
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 8, a.Len())
			assert.Equal(t, 1, b.Len())
		}},

		{"A & B cascade", func(t *testing.T) {
			a, b, r := setupRelTest(ecs.RelationCascadeDestroy, ecs.RelationCascadeDestroy)
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.False(t, r.Empty())
			assert.Equal(t, 8, a.Len())
			assert.Equal(t, 8, b.Len())

			a.Ref(1).Destroy()
			assert.Equal(t, 3, a.Len())
			assert.Equal(t, 6, b.Len())
			assert.Equal(t, ecs.NoType, b.Ref(2).Type())
			assert.Equal(t, ecs.NoType, b.Ref(3).Type())

			b.Ref(1).Destroy()
			assert.Equal(t, 1, a.Len())
			assert.Equal(t, 1, b.Len())
			assert.Equal(t, ecs.NoType, a.Ref(2).Type())
			assert.Equal(t, ecs.NoType, a.Ref(3).Type())

			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())

			a.Clear()
			assert.True(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 0, a.Len())
			assert.Equal(t, 1, b.Len())

			a, b, r = setupRelTest(ecs.RelationCascadeDestroy, ecs.RelationCascadeDestroy)
			r.Clear()
			assert.False(t, a.Empty())
			assert.False(t, b.Empty())
			assert.True(t, r.Empty())
			assert.Equal(t, 1, a.Len())
			assert.Equal(t, 1, b.Len())
		}},
	}.run(t)
}
