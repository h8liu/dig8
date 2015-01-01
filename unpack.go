package dig8

import (
	"bytes"
	"errors"
	"strings"
)

// UnpackLabels unpacks set of labels from a package buffer.
func UnpackLabels(buf *bytes.Reader, p []byte) ([]string, error) {
	isRedirect := func(b byte) bool { return b&0xc0 == 0xc0 }
	offset := func(n, b byte) int { return (int(n&0x3f) << 8) + int(b) }

	labels := make([]string, 0, 5)

	for {
		n, e := buf.ReadByte() // label length
		if e != nil {
			return nil, e
		}
		if n == 0 {
			break
		}
		if isRedirect(n) {
			b, e := buf.ReadByte()
			if e != nil {
				return nil, e
			}
			off := offset(n, b)
			if off >= len(p) {
				return nil, errors.New("offset out of range")
			}
			buf = bytes.NewReader(p[off:])
			continue
		}
		if n > 63 {
			return nil, errors.New("label too long")
		}

		labelBuf := make([]byte, n)
		if _, e := buf.Read(labelBuf); e != nil {
			return nil, e
		}

		label := strings.ToLower(string(labelBuf))

		labels = append(labels, label)
	}

	return labels, nil
}
