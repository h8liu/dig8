package dig8

import (
	"bytes"
)

// PackLabels packs the labels into a buffer.
func PackLabels(buf *bytes.Buffer, labels []string) {
	for _, lab := range labels {
		_lab := []byte(lab)
		buf.WriteByte(byte(len(_lab)))
		buf.Write(_lab)
	}
	buf.WriteByte(0)
}
