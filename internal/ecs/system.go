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

// AddProc adds processor(s) to the system.
func (sys *System) AddProc(procs ...Proc) {
	sys.Procs = append(sys.Procs, procs...)
}

// AddProcFunc adds processing fucntion(s) to the system.
func (sys *System) AddProcFunc(fns ...func()) {
	procs := make([]Proc, len(fns))
	for i := range fns {
		procs[i] = ProcFunc(fns[i])
	}
	sys.Procs = append(sys.Procs, procs...)
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
