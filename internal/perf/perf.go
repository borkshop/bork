package perf

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/borkshop/bork/internal/ecs"
)

const (
	numSamples = 64
)

// Perf is an ecs.Proc that collects Perf data.
type Perf struct {
	ecs.Proc

	outputBase    string
	shouldProfile bool
	profiling     bool
	err           error
	cpuProfF      *os.File
	profDebug     int

	round    int
	i        int
	time     [numSamples]struct{ start, end time.Time }
	memStats [numSamples]runtime.MemStats
}

// Init sets up the perf system, writing results into a timestamped directory
// with an optional name prefix.
func (perf *Perf) Init(name string, proc ecs.Proc) {
	const timeFormat = "20060102T150405Z0700"
	perf.profDebug = 2
	perf.Proc = proc
	if nowf := time.Now().Format(timeFormat); name == "" {
		perf.outputBase = fmt.Sprintf("prof-%s", nowf)
	} else {
		perf.outputBase = fmt.Sprintf("%s-prof-%s", name, nowf)
	}
}

// Process runs a round of the perf system.
func (perf *Perf) Process() {
	perf.round++

	if err := perf.maybeProfile(); err != nil {
		perf.err = err
		_ = perf.stopProfiling()
	}

	if perf.Proc != nil {
		perf.time[perf.i].start = time.Now()
		perf.Proc.Process()
		perf.time[perf.i].end = time.Now()
	}

	runtime.ReadMemStats(&perf.memStats[perf.i])

	if perf.profiling {
		if err := perf.takeProfile(); err != nil {
			perf.err = err
			_ = perf.stopProfiling()
		}
	}

	perf.i = (perf.i + 1) % numSamples
}

// Start requests profiling to start, this happens during the next Process
// round.
func (perf *Perf) Start() { perf.shouldProfile = true }

// Stop requests profiling to start, this happens during the next Process
// round.
func (perf *Perf) Stop() { perf.shouldProfile = false }

// Close cleans up the profiler, returning any error.
func (perf *Perf) Close() error {
	if serr := perf.stopProfiling(); perf.err == nil {
		perf.err = serr
	}
	return perf.err
}

// Err return any profiling error encountered; if this is non-nil, then
// profiling will not start.
func (perf *Perf) Err() error { return perf.err }

// Running returns whether profiling has been requested, and whether it
// actually active.
func (perf *Perf) Running() (should, are bool) {
	return perf.shouldProfile,
		perf.profiling
}

func (perf *Perf) maybeProfile() error {
	if perf.err != nil {
		return perf.err
	} else if perf.profiling && !perf.shouldProfile {
		return perf.stopProfiling()
	} else if !perf.profiling && perf.shouldProfile {
		return perf.startProfiling()
	}
	return nil
}

func (perf *Perf) startProfiling() error {
	if perf.profiling {
		return nil
	}
	if err := perf.copyExecutable(); err != nil {
		return err
	}
	f, err := perf.createOutput("cpu")
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		return err
	}
	perf.cpuProfF = f
	perf.profiling = true
	return perf.takeProfile()
}

func (perf *Perf) stopProfiling() (err error) {
	if perf.cpuProfF != nil {
		pprof.StopCPUProfile()
		err = perf.cpuProfF.Close()
		perf.cpuProfF = nil
		if err != nil {
			err = fmt.Errorf("failed to close \"cpu\" output file: %v", err)
		}
	}
	perf.shouldProfile = false
	perf.profiling = false
	return err
}

func (perf *Perf) takeProfile() error {
	for _, prof := range pprof.Profiles() {
		f, err := perf.createOutput(prof.Name())
		if err != nil {
			return err
		}
		err = prof.WriteTo(f, perf.profDebug)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (perf *Perf) copyExecutable() (rerr error) {
	dstName := path.Join(perf.outputBase, "exe")

	var sysStat syscall.Stat_t
	err := syscall.Stat(dstName, &sysStat)
	if err != nil && err != syscall.ENOENT {
		return err
	}

	dst, err := createMkdirAll(dstName)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); rerr == nil {
			rerr = cerr
		}
	}()

	srcName, err := os.Executable()
	if err != nil {
		return err
	}
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := src.Close(); rerr == nil {
			rerr = cerr
		}
	}()

	_, err = io.Copy(dst, src)
	return err
}

func (perf *Perf) createOutput(name string) (*os.File, error) {
	pth := path.Join(perf.outputBase, fmt.Sprintf("t%d", perf.round), name)
	f, err := createMkdirAll(pth)
	if err != nil {
		err = fmt.Errorf("failed to create %q output file: %v", name, err)
	}
	return f, err
}

func createMkdirAll(name string) (*os.File, error) {
	f, err := os.Create(name)
	if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.ENOENT {
		err = os.MkdirAll(path.Dir(name), 0777)
		if err == nil {
			return os.Create(name)
		}
	}
	return f, err
}
