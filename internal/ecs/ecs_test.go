package ecs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/ecs"
)

const (
	scData ecs.ComponentType = 1 << iota
	scD2
)

type stuff struct {
	ecs.Core

	d1 []int
	d2 [][]int
}

func newStuff() *stuff {
	s := &stuff{
		d1: []int{0},
		d2: [][]int{nil},
	}
	s.RegisterAllocator(scData, s.allocData)
	s.RegisterCreator(scD2, s.createD2)
	s.RegisterDestroyer(scD2, s.destroyD2)
	return s
}

func (s *stuff) addData(d1 int, d2 ...int) ecs.Entity {
	ent := s.AddEntity(scData)
	id := ent.ID()
	s.d1[id] = d1
	if len(d2) > 0 {
		ent.Add(scD2)
		s.d2[id] = append(s.d2[id], d2...)
	}
	return ent
}

func (s *stuff) allocData(id ecs.EntityID, t ecs.ComponentType) {
	s.d1 = append(s.d1, 0)
	s.d2 = append(s.d2, nil)
}

func (s *stuff) createD2(id ecs.EntityID, t ecs.ComponentType) {
	if s.d2[id] == nil {
		s.d2[id] = make([]int, 0, 5)
	}
}

func (s *stuff) destroyD2(id ecs.EntityID, t ecs.ComponentType) {
	s.d2[id] = s.d2[id][:0]
}

func TestBasics(t *testing.T) {
	s := newStuff()
	assert.True(t, s.Empty())

	e1 := s.AddEntity(scData)
	assert.False(t, s.Empty())

	assert.Nil(t, s.d2[e1.ID()])
	e1.Add(scD2)
	assert.NotNil(t, s.d2[e1.ID()])
	assert.Equal(t, 0, len(s.d2[e1.ID()]))

	s.d2[e1.ID()] = append(s.d2[e1.ID()], 3, 1, 4)
	assert.Equal(t, 3, len(s.d2[e1.ID()]))

	e2 := s.AddEntity(scData | scD2)
	assert.NotNil(t, s.d2[e2.ID()])
	assert.Equal(t, 0, len(s.d2[e2.ID()]))

	e1.Delete(scD2)
	assert.Equal(t, 0, len(s.d2[e1.ID()]))
	assert.NotNil(t, s.d2[e1.ID()])

	e1.Destroy()

	e3 := s.AddEntity(scData | scD2)
	assert.Equal(t, e1.ID(), e3.ID())

	assert.False(t, s.Empty())
	s.Clear()
	assert.True(t, s.Empty())
}

func TestIter_empty(t *testing.T) {
	s := newStuff()
	it := s.Iter()
	assert.Equal(t, 0, it.Count())

	assert.False(t, it.Next())
	assert.Equal(t, ecs.NilEntity, it.Entity())
	assert.Equal(t, ecs.EntityID(0), it.ID())
	assert.Equal(t, ecs.NoType, it.Type())
}

func TestIter_one(t *testing.T) {
	s := newStuff()

	s1 := s.AddEntity(scData)
	s.d1[s1.ID()] = 3

	it := s.Iter()
	assert.Equal(t, 1, it.Count())

	assert.True(t, it.Next())
	assert.Equal(t, s1, it.Entity())
	assert.Equal(t, ecs.EntityID(1), it.ID())
	assert.Equal(t, scData, it.Type())

	assert.False(t, it.Next())
	assert.Equal(t, ecs.NilEntity, it.Entity())
	assert.Equal(t, ecs.EntityID(0), it.ID())
	assert.Equal(t, ecs.NoType, it.Type())
}

func TestIter_two(t *testing.T) {
	s := newStuff()

	e1 := s.AddEntity(scData)
	s.d1[e1.ID()] = 3
	e2 := s.AddEntity(scData | scD2)
	s.d1[e2.ID()] = 4
	s.d2[e2.ID()] = append(s.d2[e2.ID()], 2, 2, 3, 5, 8)

	it := s.Iter()
	assert.Equal(t, 2, it.Count())

	// iterate all 3
	assert.True(t, it.Next())
	assert.Equal(t, e1, it.Entity())
	assert.Equal(t, ecs.EntityID(1), it.ID())
	assert.Equal(t, scData, it.Type())

	assert.True(t, it.Next())
	assert.Equal(t, e2, it.Entity())
	assert.Equal(t, ecs.EntityID(2), it.ID())
	assert.Equal(t, scData|scD2, it.Type())

	assert.False(t, it.Next())
	assert.Equal(t, ecs.NilEntity, it.Entity())
	assert.Equal(t, ecs.EntityID(0), it.ID())
	assert.Equal(t, ecs.NoType, it.Type())

	// filtering
	it = s.Iter(scD2.All())
	assert.Equal(t, 1, it.Count())

	assert.True(t, it.Next())
	assert.Equal(t, e2, it.Entity())
	assert.Equal(t, ecs.EntityID(2), it.ID())
	assert.Equal(t, scData|scD2, it.Type())

	assert.False(t, it.Next())
	assert.Equal(t, ecs.NilEntity, it.Entity())
	assert.Equal(t, ecs.EntityID(0), it.ID())
	assert.Equal(t, ecs.NoType, it.Type())
}
