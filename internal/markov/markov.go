package markov

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sort"

	"github.com/borkshop/bork/internal/ecs"
)

const componentTransition ecs.ComponentType = 1<<63 - iota

// Table implements the core of a markov transition table for use in an ECS.
//
// FIXME say more.
type Table struct {
	*ecs.Core

	weights [][]int
	next    [][]ecs.EntityID
}

// NewTable creates a new markov transition table for the given Core.
func NewTable(core *ecs.Core) *Table {
	tab := &Table{}
	tab.Init(core)
	return tab
}

// Init sets up a new markov transition table for the given Core's entity
// space; this method is useful to setup an embedded table from another
// constructor.
func (tab *Table) Init(core *ecs.Core) {
	// TODO: consider eliminating the padding for EntityID(0)
	tab.Core = core
	tab.weights = [][]int{nil}
	tab.next = [][]ecs.EntityID{nil}
	core.RegisterAllocator(componentTransition, tab.allocTransition)
	core.RegisterDestroyer(componentTransition, tab.destroyTransition)
}

func (tab *Table) allocTransition(id ecs.EntityID, t ecs.ComponentType) {
	tab.weights = append(tab.weights, nil)
	tab.next = append(tab.next, nil)
}

func (tab *Table) destroyTransition(id ecs.EntityID, t ecs.ComponentType) {
	tab.weights[id] = tab.weights[id][:0]
	tab.next[id] = tab.next[id][:0]
}

// AddTransition adds an entity transition to the table.
func (tab *Table) AddTransition(a, b ecs.Entity, weight int) {
	aid := tab.Deref(a)
	bid := tab.Deref(b)

	next, weights := tab.next[aid], tab.weights[aid]

	i := sort.Search(len(next), func(i int) bool { return next[i] >= bid })
	if i < len(next) && next[i] == bid {
		weights[i] += weight
		return
	}

	n := len(next) + 1

	if n <= cap(next) {
		next = next[:n]
	} else {
		next = append(next, 0)
	}
	copy(next[i+1:], next[i:])
	next[i] = bid
	tab.next[aid] = next

	if n <= cap(weights) {
		weights = weights[:n]
	} else {
		weights = append(weights, 0)
	}
	copy(weights[i+1:], weights[i:])
	weights[i] = weight
	tab.weights[aid] = weights
}

// ChooseNext returns a randomly chosen entity that was previously added as a
// transition.
func (tab *Table) ChooseNext(rng *rand.Rand, ent ecs.Entity) ecs.Entity {
	id := tab.Deref(ent)
	i, sum := -1, 0
	for j, w := range tab.weights[id] {
		sum += w
		if rng.Intn(sum) <= w {
			i = j
		}
	}
	if i < 0 {
		return ecs.NilEntity
	}
	id = tab.next[id][i]
	return tab.Ref(id)
}

type serd struct {
	ID      ecs.EntityID   `json:"id"`
	Next    []ecs.EntityID `json:"next"`
	Weights []int          `json:"weights"`
}

// MarhsalJSON marshal's the markov transition data into a json array.
func (tab *Table) MarhsalJSON() ([]byte, error) {
	it := tab.Iter(componentTransition.All())
	data := make([]serd, 0, it.Count())
	for it.Next() {
		id := it.ID()
		data = append(data, serd{
			ID:      id,
			Next:    tab.next[id],
			Weights: tab.weights[id],
		})
	}
	return json.Marshal(data)
}

// UnmarshalJSON unmarshal's markov transition data into this table; table must
// be empty.
func (tab *Table) UnmarshalJSON(d []byte) error {
	if tab.Len() > 0 {
		return errors.New("markov table already has data")
	}
	var data []serd
	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	n := ecs.EntityID(0)
	for _, dat := range data {
		if dat.ID >= n {
			n = dat.ID
		}
	}
	tab.next = make([][]ecs.EntityID, n)
	tab.weights = make([][]int, n)
	for _, dat := range data {
		tab.next[dat.ID] = dat.Next
		tab.weights[dat.ID] = dat.Weights
	}
	return nil
}
