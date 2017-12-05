package ecs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/ecs"
)

func setupGraphTest(flag ecs.RelationFlags) (*stuff, *ecs.Graph) {
	s := newStuff()
	s1 := s.addData(3)
	s2 := s.addData(5)
	s3 := s.addData(8)
	s4 := s.addData(13)
	s5 := s.addData(21)
	s6 := s.addData(34)
	s7 := s.addData(55)

	G := ecs.NewGraph(&s.Core, 0)
	G.Upsert(nil, func(uc *ecs.UpsertCursor) {
		uc.Create(srFoo, s1, s2)
		uc.Create(srFoo, s1, s3)
		uc.Create(srFoo, s2, s4)
		uc.Create(srFoo, s2, s5)
		uc.Create(srFoo, s3, s6)
		uc.Create(srFoo, s3, s7)
	})

	return s, G
}

func TestGraph_Roots(t *testing.T) {
	s, G := setupGraphTest(0)
	roots := G.Roots(ecs.TrueClause, nil)
	assert.Equal(t, 1, len(roots))
	assert.Equal(t, s.Ref(1), roots[0])
}

func gtids(gt ecs.GraphTraverser) []ecs.EntityID {
	var ids []ecs.EntityID
	for gt.Traverse() {
		ids = append(ids, gt.Node().ID())
	}
	return ids
}

func TestGraph_Traverse(t *testing.T) {
	testCases{

		{"DFS", func(t *testing.T) {
			_, G := setupGraphTest(0)
			gt := G.Traverse(ecs.TrueClause, ecs.TraverseDFS)
			gt.Init()
			assert.Equal(t,
				[]ecs.EntityID{1, 2, 4, 5, 3, 6, 7},
				gtids(gt))
		}},

		{"CoDFS", func(t *testing.T) {
			_, G := setupGraphTest(0)
			gt := G.Traverse(ecs.TrueClause, ecs.TraverseCoDFS)
			for _, ids := range [][]ecs.EntityID{
				{4, 2, 1},
				{5, 2, 1},
				{6, 3, 1},
				{7, 3, 1},
			} {
				gt.Init(ids[0])
				assert.Equal(t, ids, gtids(gt))
			}
		}},
	}.run(t)
}
