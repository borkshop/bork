package ecs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/ecs"
)

func setupRelTest(aFlags, bFlags ecs.RelationFlags) (a, b *stuff, rel *ecs.Relation) {
	a = newStuff()
	a1 := a.AddEntity(scData)
	a2 := a.AddEntity(scData)
	a3 := a.AddEntity(scData)
	a4 := a.AddEntity(scData)
	a5 := a.AddEntity(scData)
	a6 := a.AddEntity(scData)
	a7 := a.AddEntity(scData)
	_ = a.AddEntity(scData) // a8

	b = newStuff()
	b1 := b.AddEntity(scData)
	b2 := b.AddEntity(scData)
	b3 := b.AddEntity(scData)
	b4 := b.AddEntity(scData)
	b5 := b.AddEntity(scData)
	b6 := b.AddEntity(scData)
	b7 := b.AddEntity(scData)
	_ = b.AddEntity(scData) // b8

	rel = ecs.NewRelation(&a.Core, aFlags, &b.Core, bFlags)

	rel.InsertMany(func(insert func(r ecs.ComponentType, a ecs.Entity, b ecs.Entity) ecs.Entity) {

		insert(1, a1, b2)
		insert(1, a1, b3)
		insert(1, a2, b4)
		insert(1, a2, b5)
		insert(1, a3, b6)
		insert(1, a3, b7)

		insert(1, a2, b1)
		insert(1, a3, b1)
		insert(1, a4, b2)
		insert(1, a5, b2)
		insert(1, a6, b3)
		insert(1, a7, b3)

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
