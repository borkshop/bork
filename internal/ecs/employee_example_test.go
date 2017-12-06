package ecs_test

import (
	"fmt"
	"sort"

	"github.com/borkshop/bork/internal/ecs"
)

// shop manages employee data.
type shop struct {
	ecs.Core

	// NOTE uses a dense-array storage strategy:
	// - EntityID will simply be an index into these arrays
	// - you have an immediate choice how you want to deal with the fact that
	//   EntityID(0) is means "no entity":
	//   - you could use a [0] sentintel...
	//   - ...or you could do off-by-one math everywhere
	//   - the [0] sentinel could even serve as a "template" for new entities
	//     if that's useful
	// - this is the fundamental design trade-off of building an ECS: how much
	//   data do you put into each aspect is informed primarily by perfmance and flexibility considerations:
	//   - from a performance angle, here we make the bet that when we process
	//     name data, we don't need age data, and vice-versa (or at least that
	//     needing both is far less performance critical).
	//   - from a flexibility angle, we're saying "it'll be useful to choose
	//     whether an entity has a name and age independently"; i.e. not all
	//     entities have both if they have one.
	//   - we could split the differenece on those two angles by using one
	//     combined ComponentType code (e.g. `shopNameAge`) rather than the
	//     separated types below.
	name []string
	age  []int

	// TODO a job board and employee <-> job relations
	// TODO employee relations like reporting chain
}

// declare your type constants
//
// NOTE these need not correspond 1:1 with data storage in your structure:
// - some types may just be a flag, corresponding to no data aspect
// - some types may encompass more than one aspect of data
// - in short, they're mostly up to the ecs-embedding struct itself to hang
//   semantics off of.
const (
	shopName ecs.ComponentType = 1 << iota
	shopAge
)

// declare common combinations of our types; these are all about domain
// semantics, and not at all about ecs.Core logic.
const (
	personInfo = shopName | shopAge
)

// init hooks up the ecs.Core plumbing; note the choice to go with init-style
// methods, which allow both a `NewFoo() *Foo` style constructor, and also
// usage as a value (embedded or standalone).
func (shop *shop) init() {
	// start out with the zero sentinel and room for 1k entities
	shop.name = make([]string, 1, 1+1024)
	shop.age = make([]int, 1, 1+1024)

	// the [0] sentinel will be used by our creators to (re-)initialze data
	shop.name[0] = "Unnamed"
	shop.age[0] = 18

	// must register one or more allocators that cover all statically-allocated
	// aspects of our data; by "static" we mean: not necessarily tied to Entity
	// lifecycle or type. NOTE:
	// - allocators must be disjoint by registered type
	// - they may initialize memory (similar to creators and destroyers)
	shop.Core.RegisterAllocator(shopName|shopAge, shop.alloc)

	// creators and destroyers on the other hand, need not be disjoint (you may
	// registor 0-or-more of them for any overlapping set of types); they serve
	// as entity lifecycle callbacks.
	//
	// NOTE: you have nothing but a field of design choice here:
	// - You could clear/reset state in the destroyer...
	// - ...or in the creator ahead of re-use.
	// - You can also register destroyers and creators against ecs.NoType, and
	//   whey will fire at end-of-life and start-of-life respectively for any
	//   Entity.
	shop.Core.RegisterCreator(shopAge, shop.createAge)
	shop.Core.RegisterCreator(shopName, shop.createName)
	shop.Core.RegisterDestroyer(shopName, shop.destroyName)
	shop.Core.RegisterDestroyer(shopAge, shop.destroyAge)
}

func (shop *shop) alloc(id ecs.EntityID, t ecs.ComponentType) {
	// N.B. we could choose to copy from [0], but the creators do that, so no
	// need; if they didn't, the destroyers shoul reset to [0]'s state for consistency.
	shop.name = append(shop.name, "")
	shop.age = append(shop.age, 0)
}
func (shop *shop) createName(id ecs.EntityID, t ecs.ComponentType)  { shop.name[id] = shop.name[0] }
func (shop *shop) createAge(id ecs.EntityID, t ecs.ComponentType)   { shop.age[id] = shop.age[0] }
func (shop *shop) destroyName(id ecs.EntityID, t ecs.ComponentType) { shop.name[id] = "" }
func (shop *shop) destroyAge(id ecs.EntityID, t ecs.ComponentType)  { shop.age[id] = 0 }

// load a bunch of personal info into the shop.
func (shop *shop) loadPersonInfo(args ...interface{}) {
	for i := 0; i < len(args); {
		id := shop.AddEntity(personInfo).ID()
		shop.name[id] = args[i].(string)
		i++
		shop.age[id] = args[i].(int)
		i++
	}
}

func Example_employees() {
	var shop shop
	shop.init()

	// put some data in
	shop.loadPersonInfo(
		"Doug", 31,
		"Bob", 23,
		"Alice", 33,
		"Cathy", 27,
	)

	// NOTE finally we get to The Point: an ECS is designed primarily in
	// service to its processing phase; what you do with all the data matters
	// most, not how you use a single or a few pieces of data.

	// pull some ids out
	var ids []ecs.EntityID
	for it := shop.Iter(personInfo.All()); it.Next(); {
		ids = append(ids, it.ID())
	}

	// shake them all about
	sort.Slice(ids, func(i, j int) bool { return shop.name[ids[i]] < shop.name[ids[j]] })
	fmt.Println("Alphabetical:")
	for _, id := range ids {
		fmt.Printf("- %s\n", shop.name[id])
	}

	fmt.Println("")
	sort.Slice(ids, func(i, j int) bool { return shop.age[ids[i]] < shop.age[ids[j]] })
	fmt.Println("By Age:")
	for _, id := range ids {
		fmt.Printf("- %v %s\n", shop.age[id], shop.name[id])
	}

	// Output:
	// Alphabetical:
	// - Alice
	// - Bob
	// - Cathy
	// - Doug
	//
	// By Age:
	// - 23 Bob
	// - 27 Cathy
	// - 31 Doug
	// - 33 Alice
}
