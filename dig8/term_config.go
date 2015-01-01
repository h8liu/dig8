package dig8

import (
	"io"
)

// TermConfig are the options for configuring a term.
type TermConfig struct {
	Log       io.Writer
	Out       io.Writer
	PrintFlag int
	Retry     int
}
