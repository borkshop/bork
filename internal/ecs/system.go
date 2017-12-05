package ecs

// Proc is a piece of domain logic attached to a Core.
type Proc interface {
	Process()
}

// System is a Core with an attached set of Proc-s; it is itself a
// Proc.
type System struct {
	Core
	Procs []Proc
}

// Process calls each Proc.
func (sys *System) Process() {
	for i := range sys.Procs {
		sys.Procs[i].Process()
	}
}

// ProcFunc is a convenience for implementing Proc around an arbitrary void
// function.
type ProcFunc func()

// Process calls the wrapped function.
func (f ProcFunc) Process() { f() }
