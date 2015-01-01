package dig8

import (
	"bytes"
	"fmt"
	"io"
)

type noopWriter struct{}

var noop io.Writer = new(noopWriter)

func (w *noopWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

// Printer prints indented log.
type Printer struct {
	Prefix string
	Indent string
	Shift  int
	Writer io.Writer
	Error  error
}

// NewPrinter creates a new printer that writes to w
// if w is nil, than all prints to the printer will be noops
func NewPrinter(w io.Writer) *Printer {
	if w == nil {
		w = noop
	}

	return &Printer{
		Indent: "    ",
		Writer: w,
	}
}

func (p *Printer) p(n *int, a ...interface{}) {
	if p.Error != nil {
		return
	}

	i, e := fmt.Fprint(p.Writer, a...)
	p.Error = e
	*n += i
}

func (p *Printer) pln(n *int, a ...interface{}) {
	if p.Error != nil {
		return
	}

	i, e := fmt.Fprintln(p.Writer, a...)
	p.Error = e
	*n += i
}

func (p *Printer) pf(n *int, format string, a ...interface{}) {
	if p.Error != nil {
		return
	}

	i, e := fmt.Fprintf(p.Writer, format, a...)
	p.Error = e
	*n += i
}

func (p *Printer) pre(n *int) {
	p.p(n, p.Prefix)
	for i := 0; i < p.Shift; i++ {
		p.p(n, p.Indent)
	}
}

// Print prints stuff similar to log.Print
func (p *Printer) Print(a ...interface{}) (int, error) {
	n := 0
	p.pre(&n)
	p.p(&n, a...)
	p.pln(&n)

	return n, p.Error
}

// Println prints stuff similar to log.Println
func (p *Printer) Println(a ...interface{}) (int, error) {
	n := 0
	p.pre(&n)
	p.pln(&n, a...)

	return n, p.Error
}

// Printf prints stuff similar to log.Printf
func (p *Printer) Printf(format string, a ...interface{}) (int, error) {
	n := 0
	p.pre(&n)
	p.pf(&n, format, a...)
	p.pln(&n)

	return n, p.Error
}

// ShiftIn indents right one level.
func (p *Printer) ShiftIn() {
	p.Shift++
}

// ShiftOut indents left one level.
func (p *Printer) ShiftOut(a ...interface{}) {
	if p.Shift == 0 {
		panic("shift already left most")
	}
	p.Shift--

	if len(a) > 0 {
		p.Print(a...)
	}
}

// Printable is an object that can be printed into
// a printer.
type Printable interface {
	PrintTo(p *Printer)
}

// PrintStr converts a printable object into a string
func PrintStr(p Printable) string {
	buf := new(bytes.Buffer)
	dev := NewPrinter(buf)
	p.PrintTo(dev)
	return buf.String()
}
