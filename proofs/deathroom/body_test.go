package main

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/borkshop/bork/internal/ecs"
)

func TestBody_sever_vital(t *testing.T) {
	for _, tc := range []struct {
		name     string
		destroy  bool
		part     ecs.ComponentType
		expected []string
	}{
		{
			"sever left thigh",
			false,
			bcLeft | bcThigh,
			[]string{"left calf", "left foot", "left thigh"},
		},
		{
			"destroy left thigh",
			true,
			bcLeft | bcThigh,
			[]string{"left calf", "left foot"},
		},

		{
			"destroy head",
			true,
			bcHead,
			[]string{
				"left calf", "left foot",
				"left forearm", "left hand",
				"left thigh", "left upper arm",
				"right calf", "right foot",
				"right forearm", "right hand",
				"right thigh", "right upper arm",
				"torso",
			},
		},

		{
			"destroy torso",
			true,
			bcTorso,
			[]string{
				"head",
				"left calf", "left foot",
				"left forearm", "left hand",
				"left thigh", "left upper arm",
				"right calf", "right foot",
				"right forearm", "right hand",
				"right thigh", "right upper arm",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(rand.Int63()))

			bo := newBody()
			bo.build(rng)

			it := bo.Iter((tc.part | bcPart).All())
			require.True(t, it.Next())
			part := it.Entity()
			require.NotEqual(t, ecs.NilEntity, part)

			prior := bo.Len()

			if tc.destroy {
				bo.hp[part.ID()] = 0
			}

			severed := bo.sever(t.Logf, part)
			require.NotNil(t, severed)

			it = severed.Iter()
			got := make([]string, 0, it.Count())
			for it.Next() {
				got = append(got, severed.DescribePart(it.Entity()))
			}
			sort.Strings(got)

			if assert.Equal(t, tc.expected, got) {
				n := prior - len(tc.expected)
				if tc.destroy {
					n--
				}
				assert.Equal(t, n, bo.Len())
			}
		})
	}
}
