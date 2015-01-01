package dig8

import (
	"bytes"
)

// Rdata is a general interface for rdata records.
type Rdata interface {
	PrintTo(out *bytes.Buffer)
	Pack() []byte
}
