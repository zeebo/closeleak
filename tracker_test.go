package closeleak

import (
	"runtime"
	"testing"
)

func init() { Enable() }

func TestTracker(t *testing.T) {
	// Doesn't actually test anything, but is useful
	// to check what a leak looks like.

	cl := New()
	cl = nil
	runtime.GC()
	_ = cl
}

func BenchmarkTracker(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		t := New()
		t.Close()
	}
}
