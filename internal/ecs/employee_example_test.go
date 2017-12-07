package ecs_test

import (
	"fmt"
	"sort"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/moremath"
)

type workers struct {
	ecs.Core

	name   []string
	age    []int
	skills []skill

	// TODO employee relations like reporting chain
}

type jobs struct {
	ecs.Core

	name   []string
	skills []skill
	work   []int
}

type skill struct {
	brawn int
	brain int
}

type assignment struct {
	ecs.Relation
	wrk *workers
	jb  *jobs

	// NOTE can attach data to each relation since
	// ecs.Relation is just an ecs.Core
}

// declare your type constants
//
// NOTE these need not correspond 1:1 with data storage in your structure:
// - some types may just be a flag, corresponding to no data aspect
// - some types may encompass more than one aspect of data
// - in short, they're mostly up to the ecs-embedding struct itself to hang
//   semantics off of.
const (
	workerName ecs.ComponentType = 1 << iota
	workerStats
	workerAssigned
)

const (
	jobInfo ecs.ComponentType = 1 << iota
	jobWork
	jobAssigned
)

const (
	amtWorking ecs.ComponentType = 1 << iota
	// NOTE components may define kinds of relation, and/or attached data
)

// declare common combinations of our types; these are all about domain
// semantics, and not at all about ecs.Core logic.
const (
	workerInfo = workerName | workerStats
)

// init hooks up the ecs.Core plumbing; note the choice to go with init-style
// methods, which allow both a `NewFoo() *Foo` style constructor, and also
// usage as a value (embedded or standalone).
func (wrk *workers) init() {
	// start out with the zero sentinel and room for 1k entities
	wrk.name = make([]string, 1, 1+1024)
	wrk.age = make([]int, 1, 1+1024)
	wrk.skills = make([]skill, 1, 1+1024)

	// the [0] sentinel will be used by our creators to (re-)initialze data
	wrk.name[0] = "Unnamed"
	wrk.age[0] = 18
	wrk.skills[0] = skill{1, 1}

	// must register one or more allocators that cover all statically-allocated
	// aspects of our data; by "static" we mean: not necessarily tied to Entity
	// lifecycle or type. NOTE:
	// - allocators must be disjoint by registered type
	// - they may initialize memory (similar to creators and destroyers)
	wrk.Core.RegisterAllocator(workerInfo, wrk.alloc)

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
	wrk.Core.RegisterCreator(workerName, wrk.createName)
	wrk.Core.RegisterCreator(workerStats, wrk.createStats)
	wrk.Core.RegisterDestroyer(workerName, wrk.destroyName)
	wrk.Core.RegisterDestroyer(workerStats, wrk.destroyStats)
}

func (jb *jobs) init() {
	jb.name = make([]string, 1, 1+1024)
	jb.skills = make([]skill, 1, 1+1024)
	jb.work = make([]int, 1, 1+1024)

	jb.Core.RegisterAllocator(jobInfo|jobWork, jb.alloc)
	jb.Core.RegisterCreator(jobInfo, jb.createInfo)
	jb.Core.RegisterCreator(jobWork, jb.createWork)
	jb.Core.RegisterDestroyer(jobInfo, jb.destroyInfo)
	jb.Core.RegisterDestroyer(jobWork, jb.destroyWork)
}

func (amt *assignment) init(wrk *workers, jb *jobs) {
	amt.wrk = wrk
	amt.jb = jb
	amt.Relation.Init(&wrk.Core, 0, &jb.Core, 0)
}

func (wrk *workers) alloc(id ecs.EntityID, t ecs.ComponentType) {
	// N.B. we could choose to copy from [0], but the creators do that, so no
	// need; if they didn't, the destroyers shoul reset to [0]'s state for consistency.
	wrk.name = append(wrk.name, "")
	wrk.age = append(wrk.age, 0)
	wrk.skills = append(wrk.skills, skill{})
}
func (wrk *workers) createName(id ecs.EntityID, t ecs.ComponentType)  { wrk.name[id] = wrk.name[0] }
func (wrk *workers) destroyName(id ecs.EntityID, t ecs.ComponentType) { wrk.name[id] = "" }
func (wrk *workers) createStats(id ecs.EntityID, t ecs.ComponentType) {
	wrk.age[id] = wrk.age[0]
	wrk.skills[id] = wrk.skills[0]
}
func (wrk *workers) destroyStats(id ecs.EntityID, t ecs.ComponentType) {
	wrk.age[id] = 0
	wrk.skills[id] = skill{}
}

func (jb *jobs) alloc(id ecs.EntityID, t ecs.ComponentType) {
	jb.name = append(jb.name, "")
	jb.skills = append(jb.skills, skill{})
	jb.work = append(jb.work, 0)
}
func (jb *jobs) createInfo(id ecs.EntityID, t ecs.ComponentType) {
	jb.name[id] = jb.name[0]
	jb.skills[id] = jb.skills[0]
}
func (jb *jobs) destroyInfo(id ecs.EntityID, t ecs.ComponentType) {
	jb.name[id] = ""
	jb.skills[id] = skill{}
}
func (jb *jobs) createWork(id ecs.EntityID, t ecs.ComponentType)  { jb.work[id] = jb.work[0] }
func (jb *jobs) destroyWork(id ecs.EntityID, t ecs.ComponentType) { jb.work[id] = 0 }

func (wrk *workers) load(args ...interface{}) {
	for i := 0; i < len(args); {
		id := wrk.AddEntity(workerInfo).ID()
		wrk.name[id] = args[i].(string)
		i++
		wrk.age[id] = args[i].(int)
		i++
		wrk.skills[id].brawn = args[i].(int)
		i++
		wrk.skills[id].brain = args[i].(int)
		i++
	}
}

func (jb *jobs) load(args ...interface{}) {
	for i := 0; i < len(args); {
		id := jb.AddEntity(workerInfo).ID()
		jb.name[id] = args[i].(string)
		i++
		jb.skills[id].brawn = args[i].(int)
		i++
		jb.skills[id].brain = args[i].(int)
		i++
		jb.work[id] = args[i].(int)
		i++
	}
}

func (sk skill) key() uint64 {
	return moremath.Shuffle(
		moremath.ClampInt32(sk.brawn),
		moremath.ClampInt32(sk.brain),
	)
}

func (wrk *workers) unassigned() ecs.Iterator {
	cl := ecs.And(
		workerStats.All(),
		workerAssigned.NotAny(),
	)
	return wrk.Iter(cl)
}

func (jb *jobs) unassigned() ecs.Iterator {
	return jb.Iter(ecs.And(
		(jobInfo | jobWork).All(),
		jobAssigned.NotAny(),
	))
}

func (amt *assignment) assign() {
	// order workers by their skill
	var wids []ecs.EntityID
	for it := amt.wrk.unassigned(); it.Next(); {
		wids = append(wids, it.ID())
	}
	sort.Slice(wids, func(i, j int) bool {
		return amt.wrk.skills[wids[i]].key() < amt.wrk.skills[wids[j]].key()
	})

	// match each job with best worker
	amt.Upsert(nil, func(uc *ecs.UpsertCursor) {
		for it := amt.jb.unassigned(); len(wids) > 0 && it.Next(); {
			// pick best worker and remove it from the list
			jk := amt.jb.skills[it.ID()].key()
			wix := sort.Search(len(wids), func(i int) bool {
				return amt.wrk.skills[wids[i]].key() >= jk
			})
			if wix >= len(wids) {
				wix = len(wids) - 1
			}
			wid := wids[wix]
			copy(wids[wix:], wids[wix+1:])
			wids = wids[:len(wids)-1]

			// assign worker to job
			worker, job := amt.wrk.Ref(wid), it.Entity()
			uc.Create(amtWorking, worker, job)
			worker.Add(workerAssigned)
			job.Add(jobAssigned)
		}
	})
}

func Example_employees() {
	var (
		// NOTE if this were for real, you'd probably wrap
		// this in a world struct; world itself could be an
		// ecs.Core to bind worker and job info into some sort
		// of space.
		wrk workers
		jb  jobs
		amt assignment
	)

	wrk.init()
	jb.init()
	amt.init(&wrk, &jb)

	// put some data in
	wrk.load(
		"Doug", 31, 7, 4,
		"Bob", 23, 3, 6,
		"Alice", 33, 4, 7,
		"Cathy", 27, 8, 4,
		"Ent", 25, 6, 6,
	)
	jb.load(
		"t1", 10, 1, 5,
		"t2", 1, 10, 5,
		"t3", 3, 6, 3,
		"t4", 6, 3, 3,
		"t5", 4, 4, 2,
	)

	// NOTE finally we get to The Point: an ECS is designed primarily in
	// service to its processing phase; what you do with all the data matters
	// most, not how you use a single or a few pieces of data.

	amt.assign()

	fmt.Printf("assignments:\n")
	for cur := amt.Select(amtWorking.All()); cur.Scan(); {
		worker, job := cur.A(), cur.B()
		ws := wrk.skills[worker.ID()]
		js := jb.skills[job.ID()]
		fmt.Printf(
			"- %s working on %q, rating: <%.1f, %.1f>\n",
			wrk.name[worker.ID()],
			jb.name[job.ID()],
			float64(ws.brawn)/float64(js.brawn),
			float64(ws.brain)/float64(js.brain),
		)
	}

	it := wrk.unassigned()
	fmt.Printf("\nunassigned workers %v\n", it.Count())
	for it.Next() {
		fmt.Printf("- %v %q\n", it.Entity(), wrk.name[it.ID()])
	}

	it = jb.unassigned()
	fmt.Printf("\nunassigned jobs %v\n", it.Count())
	for it.Next() {
		fmt.Printf("- %v %q\n", it.Entity(), jb.name[it.ID()])
	}

	// Output:
	// assignments:
	// - Cathy working on "t1", rating: <0.8, 4.0>
	// - Ent working on "t2", rating: <6.0, 0.6>
	// - Bob working on "t3", rating: <1.0, 1.0>
	// - Doug working on "t4", rating: <1.2, 1.3>
	// - Alice working on "t5", rating: <1.0, 1.8>
	//
	// unassigned workers 0
	//
	// unassigned jobs 0
}
