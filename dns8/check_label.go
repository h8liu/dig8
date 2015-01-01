package dns8

import (
	"errors"
	"fmt"
)

func checkLabel(label string) error {
	nl := len(label)
	if nl == 0 {
		return errors.New("empty label")
	}
	if nl >= 64 {
		return errors.New("label too long")
	}
	if label[0] == '-' {
		return errors.New("label starts with dash")
	}
	if label[nl-1] == '-' {
		return errors.New("label ends with dash")
	}

	for _, c := range label {
		if 'a' <= c && c <= 'z' {
			continue
		}
		if '0' <= c && c <= '9' {
			continue
		}
		if c == '_' || c == '-' {
			continue
		}

		return fmt.Errorf("invalid char: %c", c)
	}

	return nil
}
