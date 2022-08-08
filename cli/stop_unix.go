//go:build unix || (js && wasm)

package cli

import (
	"os"
	"syscall"
)

func quitSignal() os.Signal {
	return syscall.SIGQUIT
}
