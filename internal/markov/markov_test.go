package markov_test

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/markov"
)

const (
	componentWord ecs.ComponentType = 1 << iota
)

type Corpus struct {
	ecs.Core
	markov.Table

	word   []string
	lookup map[string]ecs.EntityID
}

func NewCorpus() *Corpus {
	c := &Corpus{
		word:   []string{""},
		lookup: make(map[string]ecs.EntityID),
	}
	c.Table.Init(&c.Core)
	c.RegisterAllocator(componentWord, c.allocWord)
	c.RegisterDestroyer(componentWord, c.destroyWord)
	return c
}

func (c *Corpus) allocWord(id ecs.EntityID, t ecs.ComponentType) {
	c.word = append(c.word, "")
}

func (c *Corpus) destroyWord(id ecs.EntityID, t ecs.ComponentType) {
	delete(c.lookup, c.word[id])
	c.word[id] = ""
}

func (c *Corpus) Enity(s string) ecs.Entity {
	if id, def := c.lookup[s]; def {
		return c.Ref(id)
	}
	ent := c.AddEntity(componentWord)
	c.word[ent.ID()] = s
	c.lookup[s] = ent.ID()
	return ent
}

func (c *Corpus) EntityString(ent ecs.Entity) string {
	id := c.Deref(ent)
	if ent.Type().HasAll(componentWord) {
		return c.word[id]
	}
	return ""
}

func (c *Corpus) Ingest(chain []string) {
	term := c.Enity("")
	last := term
	for _, s := range chain {
		ent := c.Enity(s)
		c.AddTransition(last, ent, 1)
		last = ent
	}
	c.AddTransition(last, term, 1)
}

func Example_markovChain() {
	c := NewCorpus()
	for _, s := range []string{
		"it was the best of times",
		"it was the worst of times",
		"power brings out the worst in men",
		"I always try to be my best",
		"the best of the best hoorah",
	} {
		c.Ingest(strings.Fields(s))
	}

	rng := rand.New(rand.NewSource(0))

	var parts []string
	for i := 0; i < 10; i++ {
		ent := c.Enity("")
		for {
			ent = c.ChooseNext(rng, ent)
			s := c.EntityString(ent)
			if s == "" {
				break
			}
			parts = append(parts, s)
		}
		fmt.Printf("%s\n", strings.Join(parts, " "))
		parts = parts[:0]
	}
	// Output:
	// I always try to be my best of times
	// it was the worst in men
	// power brings out the worst in men
	// the best hoorah
	// power brings out the best hoorah
	// I always try to be my best hoorah
	// the worst in men
	// power brings out the best hoorah
	// I always try to be my best of times
	// power brings out the best hoorah

}
