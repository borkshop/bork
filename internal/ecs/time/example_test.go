package time_test

import (
	"fmt"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/ecs/time"
)

const (
	wcTime ecs.ComponentType = 1 << iota
	wcFoo
	wcBar
	wcBaz
)

type world struct {
	ecs.System
	time time.Facility
}

func (w *world) init() {
	w.time.Init(&w.Core, wcTime)
	w.Procs = append(w.Procs,
		&w.time,
		ecs.ProcFunc(w.dump),
	)

	w.RegisterCreator(wcFoo, func(id ecs.EntityID, _ ecs.ComponentType) {
		fmt.Printf("[%v] +foo after 2 bar\n", id)
		w.time.After(w.Ref(id), 2, wcBar.ApplyTo)
	})

	w.RegisterCreator(wcBar, func(id ecs.EntityID, _ ecs.ComponentType) {
		fmt.Printf("[%v] +bar after 3 none\n", id)
		w.time.After(w.Ref(id), 3, ecs.NoType.ApplyTo)
	})

	w.RegisterCreator(wcBaz, func(id ecs.EntityID, _ ecs.ComponentType) {
		fmt.Printf("[%v] +baz after 4 bar\n", id)
		w.time.After(w.Ref(id), 4, wcBar.ApplyTo)
	})

	w.RegisterDestroyer(ecs.NoType, func(id ecs.EntityID, _ ecs.ComponentType) {
		fmt.Printf("[%v] gone\n", id)
	})

}

func (w *world) dump() {
	fmt.Printf("now=%v\n", w.time.Now())
	for it := w.Iter(); it.Next(); {
		if it.Type() != ecs.NoType {
			fmt.Printf("- [%v]<%v>\n", it.ID(), it.Type())
		}
	}
	fmt.Printf("\n")
}

func Example() {
	var w world
	w.init()

	w.AddEntity(wcFoo)
	w.AddEntity(wcBar)
	w.AddEntity(wcBaz)
	w.dump()

	for i := 0; i < 1000 && !w.Empty(); i++ {
		w.Process()
	}

	// Output:
	// [1] +foo after 2 bar
	// [2] +bar after 3 none
	// [3] +baz after 4 bar
	// now=t0
	// - [1]<0000000000000003>
	// - [2]<0000000000000005>
	// - [3]<0000000000000009>
	//
	// now=t1
	// - [1]<0000000000000003>
	// - [2]<0000000000000005>
	// - [3]<0000000000000009>
	//
	// [1] +bar after 3 none
	// now=t2
	// - [1]<0000000000000005>
	// - [2]<0000000000000005>
	// - [3]<0000000000000009>
	//
	// [2] gone
	// now=t3
	// - [1]<0000000000000005>
	// - [3]<0000000000000009>
	//
	// [3] +bar after 3 none
	// now=t4
	// - [1]<0000000000000005>
	// - [3]<0000000000000005>
	//
	// [1] gone
	// now=t5
	// - [3]<0000000000000005>
	//
	// now=t6
	// - [3]<0000000000000005>
	//
	// [3] gone
	// now=t7
	//
}
