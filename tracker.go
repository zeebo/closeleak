package closeleak

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
)

// enabled keeps track of if the trackers are enabled.
var enabled uint64

// Enable causes the returned trackers to actually check.
// It only affects calls to New after it is called.
// It is safe to be called concurrently.
func Enable() { atomic.StoreUint64(&enabled, 1) }

// Disable causes the returned trackers to not check.
// It only affects calls to New after it is called.
// It is safe to be called concurrently.
func Disable() { atomic.StoreUint64(&enabled, 0) }

// Tracker keeps track of if it was closed. If it was not closed
// then it logs stack information about where it was created to
// standard error when finalized.
type Tracker struct{ stack []uintptr }

// New returns a tracker if enabled. If disabled, it returns nil.
// It is safe to be called concurrently.
func New() *Tracker {
	if atomic.LoadUint64(&enabled) == 0 {
		return nil
	}

	var buf [256]uintptr
	n := runtime.Callers(2, buf[:])

	t := &Tracker{stack: make([]uintptr, n)}
	copy(t.stack, buf[:n])

	runtime.SetFinalizer(t, finalize)
	return t
}

// Close should be called on the tracker, or it will log to
// standard error if finalized. It is safe to call on nil.
func (t *Tracker) Close() {
	if t != nil {
		runtime.SetFinalizer(t, nil)
	}
}

// finalize is called when a tracker is leaked.
func finalize(t *Tracker) {
	fmt.Fprintln(os.Stderr, "==================")
	fmt.Fprintf(os.Stderr, "WARNING: CLOSE LEAKED (%p)\n", t)
	frames := runtime.CallersFrames(t.stack)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		diff := uintptr(0)
		if frame.Entry != 0 {
			diff = frame.PC - frame.Entry + 1
		}

		fmt.Fprintf(os.Stderr, "%s(...)\n\t%s:%d +0x%x\n",
			frame.Function, frame.File, frame.Line, diff)
	}
	fmt.Fprintln(os.Stderr, "==================")
}
