package dns8

import (
	"time"
)

// job is a query job for a client
type job struct {
	id       uint16
	exchange *Exchange
	deadline time.Time
	printer  *Printer
	c        chan<- *Exchange
}

func (j *job) Close() {
	if j.printer != nil {
		j.exchange.printRecv(j.printer)
	}
	j.c <- j.exchange
}

func (j *job) CloseErr(e error) {
	j.exchange.Error = e
	j.Close()
}

func (j *job) CloseRecv(m *Message) {
	j.exchange.Recv = m
	bugOn(j.exchange.Error != nil)
	j.Close()
}
