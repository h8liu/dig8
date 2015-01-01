package dns8

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
)

// Domain saves a domain name
type Domain struct {
	name   string
	labels []string
}

// Equal returns true if the domain is the same.
// It also works with nill domains.
func (d *Domain) Equal(o *Domain) bool {
	switch {
	case d == nil:
		return o == nil
	case o == nil:
		return false
	case len(d.labels) != len(o.labels):
		return false
	}

	for i, lab := range d.labels {
		if o.labels[i] != lab {
			return false
		}
	}
	return true
}

// String returns the domain name string representation
func (d *Domain) String() string {
	if d.IsRoot() {
		return "."
	}
	return d.name
}

// Root is the root domain.
var Root = &Domain{"", []string{}}

// ParseDomain parses a domain from a string.
func ParseDomain(s string) (*Domain, error) {
	// a helper for generating error messages
	err := func(s, r string) error {
		return fmt.Errorf("'%s': %s", s, r)
	}

	orig := s

	ip := net.ParseIP(s)
	if ip != nil {
		return nil, err(orig, "IP addr")
	}

	n := len(s)
	if n > 255 {
		return nil, err(orig, "name too long")
	}
	if n > 0 && s[n-1] == '.' {
		s = s[:n-1]
	}

	if s == "" {
		return Root, nil
	}

	s = strings.ToLower(s)
	labels := strings.Split(s, ".")

	for _, label := range labels {
		e := checkLabel(label)
		if e != nil {
			return nil, err(orig, e.Error())
		}
	}

	return &Domain{s, labels}, nil
}

// D is a shortcut to generate domains.
// It panics when the domain is invalid
func D(s string) *Domain {
	ret, e := ParseDomain(s)
	if e != nil {
		panic(e)
	}
	return ret
}

// IsRoot checks if the domain is root.
func (d *Domain) IsRoot() bool { return len(d.labels) == 0 }

// IsZoneOf checks if the domain o is in the zone of d.
func (d *Domain) IsZoneOf(o *Domain) bool {
	return d.Equal(o) || d.IsParentOf(o)
}

// IsParentOf checks if domain d is parent of domain o.
func (d *Domain) IsParentOf(o *Domain) bool { return o.IsChildOf(d) }

// IsChildOf checks if domain d is child domain of domain o.
func (d *Domain) IsChildOf(c *Domain) bool {
	n := len(d.labels)
	nc := len(c.labels)
	if nc >= n {
		return false
	}

	delta := n - nc
	for i, lab := range c.labels {
		if d.labels[i+delta] != lab {
			return false
		}
	}

	return true
}

// Parent returns the parent domain.
// Root.Parent() == nil.
func (d *Domain) Parent() *Domain {
	if d.IsRoot() {
		return nil
	}
	if len(d.labels) == 1 {
		return Root
	}

	labels := d.labels[1:]
	name := d.name[len(d.labels[0])+1:]
	return &Domain{name, labels}
}

// RegParts returns the registered domain and registrar domain.
// If the domain is already a registrar domain, it returns
// nil for the registered part.
func (d *Domain) RegParts() (registered *Domain, registrar *Domain) {
	var last *Domain
	cur := d
	parent := d.Parent()
	for {
		if parent == nil {
			// cur is root
			return last, cur
		}
		if parent.IsRoot() {
			// top level domain
			return last, cur
		}
		if superRegs[cur.name] {
			return last, cur
		}
		if superRegs[parent.name] && nonRegs[cur.name] {
			return last, cur
		}
		if regs[cur.name] {
			return last, cur
		}

		last = cur
		cur = parent
		parent = parent.Parent()
	}
}

// Registered returns the registered domain.
func (d *Domain) Registered() *Domain {
	ret, _ := d.RegParts()
	return ret
}

// Registrar returns the registrar domain.
func (d *Domain) Registrar() *Domain {
	_, ret := d.RegParts()
	return ret
}

// IsRegistrar checks if the domain is a registrar.
func (d *Domain) IsRegistrar() bool {
	reged, _ := d.RegParts()
	return reged == nil
}

// Pack packs the labels into a buffer.
func (d *Domain) Pack(buf *bytes.Buffer) {
	PackLabels(buf, d.labels)
}

// UnpackDomain creates a domain from a dns packet buffer.
func UnpackDomain(buf *bytes.Reader, p []byte) (*Domain, error) {
	labels, e := UnpackLabels(buf, p)
	if e != nil {
		return nil, e
	}

	for _, lab := range labels {
		if e := checkLabel(lab); e != nil {
			return nil, e
		}
	}

	name := strings.Join(labels, ".")

	if len(name) > 255 {
		return nil, errors.New("domain too long")
	}

	return &Domain{name, labels}, nil
}
