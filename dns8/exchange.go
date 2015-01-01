package dns8

import (
	"fmt"
)

// Exchange is the packet exchange for a query.
// It has the message sent and the message received.
type Exchange struct {
	Query     *Query
	Send      *Message
	Recv      *Message
	Error     error
	PrintFlag int
}

// PrintTo prints the exchange to a printer
func (x *Exchange) PrintTo(p *Printer) {
	x.printSend(p)
	x.printRecv(p)
}

func (x *Exchange) printSend(p *Printer) {
	p.Printf("%s {", x.Query.String())
	p.ShiftIn()

	switch x.PrintFlag {
	case PrintAll:
		p.Print("send {")
		p.ShiftIn()
		x.Send.PrintTo(p)
		p.ShiftOut("}")
	case PrintReply:
		// do nothing
	default:
		panic("unknown print flag")
	}
}

func (x *Exchange) printTimeTaken(p *Printer) {
	d := x.Recv.Timestamp.Sub(x.Send.Timestamp)
	n := d.Nanoseconds()
	var s string
	if n < 1e3 {
		s = fmt.Sprintf("%dns", n)
	} else if n < 1e6 {
		s = fmt.Sprintf("%.1fus", float64(n)/1e3)
	} else if n < 1e9 {
		s = fmt.Sprintf("%.2fms", float64(n)/1e6)
	} else {
		s = fmt.Sprintf("%.3fs", float64(n)/1e9)
	}

	p.Printf("(in %v)", s)
}

func (x *Exchange) printRecv(p *Printer) {
	switch x.PrintFlag {
	case PrintAll:
		if x.Recv != nil {
			p.Print("recv {")
			p.ShiftIn()
			x.Recv.PrintTo(p)
			p.ShiftOut("}")
		}

		if x.Error != nil {
			p.Printf("error %v", x.Error)
		}
	case PrintReply:
		if x.Recv != nil {
			x.Recv.Packet.PrintTo(p)
			x.printTimeTaken(p)
		}
		if x.Error != nil {
			p.Printf("error %v", x.Error)
		}
	default:
		panic("unknown print flag")
	}

	p.ShiftOut("}")
}

func (x *Exchange) String() string {
	return PrintStr(x)
}

// Timeout checks if the exchange has met a timeout error
func (x *Exchange) Timeout() bool {
	return x.Error == errTimeout
}
