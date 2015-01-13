package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validArchName(n string) bool {
	if n == "" {
		return false
	}
	for _, r := range n {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r == '-' || r == '_' {
			continue
		}
		return false
	}

	return true
}

func parseArch(a string) (string, error) {
	if a == "" {
		return a, nil
	}

	names := strings.Split(a, ".")

	for _, n := range names {
		if !validArchName(n) {
			return a, fmt.Errorf("invalid archive path: %s", a)
		}
	}

	return filepath.Join(names...), nil
}

func dequeue() {

}
