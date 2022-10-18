// Source: https://github.com/ukautz/clif/blob/v1/output.go
package std

import (
	"fmt"
	"io"
	"os"
)

// Output is interface for
type Output interface {

	// Printf applies format (renders styles) and writes to output
	Printf(msg string, args ...any)

	// Sprintf applies format (renders styles) and returns as string
	Sprintf(msg string, args ...any) string

	// Writer returns the `io.Writer` used by this output
	Writer() io.Writer
}

// DefaultOutput is the default used output type
type DefaultOutput struct {
	io io.Writer
}

// NewOutput generates a new (default) output with provided io writer (if nil
// then `os.Stdout` is used) and a formatter
func NewDefaultOutput(io io.Writer) *DefaultOutput {
	if io == nil {
		io = os.Stdout
	}
	return &DefaultOutput{
		io: io,
	}
}

func (o *DefaultOutput) Printf(msg string, args ...any) {
	o.io.Write([]byte(o.Sprintf(msg, args...)))
}

func (o *DefaultOutput) Sprintf(msg string, args ...any) string {
	return fmt.Sprintf(msg, args...)
}

func (o *DefaultOutput) Writer() io.Writer {
	return o.io
}
