//go:build !(unix || (js && wasm))

package cli

import (
	"os"
	"runtime"
)

// quitSignal returns the appropriate signal to use to request that a process
// quit execution.
func quitSignal() os.Signal {
	if runtime.GOOS == "windows" {
		// Per https://golang.org/pkg/os/#Signal, “Interrupt is not implemented on
		// Windows; using it with os.Process.Signal will return an error.”
		// Fall back to Kill instead.
		return os.Kill
	}
	return os.Interrupt
}
