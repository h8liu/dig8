package dns8

import (
	"encoding/hex"
	"log"
	"net"
	"time"
)

// Client is a DNS query client
type Client struct {
	conn   *net.UDPConn
	idPool *idPool

	jobs       map[uint16]*job
	newJobs    chan *job
	sendErrors chan *job
	recvs      chan *Message
	timer      <-chan time.Time

	closed  bool
	closing chan struct{}

	Logger *log.Logger
}

// NewClientPort creates a client at a particular port
func NewClientPort(port uint16) (*Client, error) {
	ret := new(Client)

	addr := &net.UDPAddr{Port: int(port)}
	if port == 0 {
		addr = nil
	}

	var e error
	ret.conn, e = net.ListenUDP("udp4", addr)
	if e != nil {
		return nil, e
	}

	ret.newJobs = make(chan *job, 0)
	ret.sendErrors = make(chan *job, 10)
	ret.recvs = make(chan *Message, 10)
	ret.timer = time.Tick(time.Millisecond * 100)
	ret.idPool = newIDPool()
	ret.jobs = make(map[uint16]*job)
	ret.closing = make(chan struct{})

	go ret.recv()
	go ret.serve()

	return ret, nil
}

// NewClient creates a client at port 0 (any port available).
func NewClient() (*Client, error) {
	return NewClientPort(0)
}

const packetMaxSize = 1600

func newRecvBuf() []byte {
	return make([]byte, packetMaxSize)
}

// Close the client (asyncly)
func (c *Client) Close() error {
	c.closed = true
	c.closing <- struct{}{}
	return c.conn.Close()
}

func (c *Client) recv() {
	buf := newRecvBuf()

	for {
		n, addr, e := c.conn.ReadFromUDP(buf)
		if e != nil {
			if c.closed {
				break
			}

			if c.Logger != nil {
				c.Logger.Print("recv:", e)
			}
			continue
		}

		p, e := Unpack(buf[:n])
		if e != nil {
			if c.Logger != nil {
				c.Logger.Print("unpack: ", e)
				c.Logger.Print(hex.Dump(buf[:n]))
			}

			continue
		}

		m := &Message{
			RemoteAddr: addr,
			Packet:     p,
			Timestamp:  time.Now(),
		}
		c.recvs <- m

		buf = newRecvBuf()
	}
}

func (c *Client) delJob(id uint16) {
	if c.jobs[id] != nil {
		delete(c.jobs, id)
		c.idPool.Return(id)
	}
}

func (c *Client) serve() {
	for {
		select {
		case <-c.closing:
			return
		case job := <-c.newJobs:
			id := job.id
			bugOn(c.jobs[id] != nil)
			c.jobs[id] = job
		case job := <-c.sendErrors:
			/*
				Need to check if it is still the same job. In some rare racing
				cases, sendErrors will be delayed (like by a send that takes
				too long), and timeout might trigger first, hence reallocate
				the id to another job.
			*/
			if c.jobs[job.id] == job {
				// still the same job
				c.delJob(job.id)
			}
		case m := <-c.recvs:
			id := m.Packet.ID
			job := c.jobs[id]
			if job == nil {
				// might happen when the timeout window is too small
				if c.Logger != nil {
					c.Logger.Printf("recved zombie msg with id %d", id)
				}
			} else {
				bugOn(job.id != id)
				job.CloseRecv(m)
				c.delJob(id)
			}
		case now := <-c.timer:
			timeouts := make([]uint16, 0, 1024)

			for id, job := range c.jobs {
				bugOn(job.id != id)
				if job.deadline.Before(now) {
					job.CloseErr(errTimeout)

					// iterating the map, so delete afterwards for safty
					timeouts = append(timeouts, id)
				}
			}

			for _, id := range timeouts {
				c.delJob(id)
			}
		}
	}
}

const timeout = time.Second * 3

// Send schedules a new query. It sends the exchange data back
// to the channel
func (c *Client) Send(q *QueryPrinter, ch chan<- *Exchange) {
	id := c.idPool.Fetch()
	message := newMessage(q.Query, id)
	if message.RemoteAddr.Port == 0 {
		message.RemoteAddr.Port = DNSPort
	}

	exchange := &Exchange{
		Query:     q.Query,
		Send:      message,
		PrintFlag: q.PrintFlag,
	}
	job := &job{
		id:       id,
		exchange: exchange,
		deadline: time.Now().Add(timeout),
		printer:  q.Printer,
		c:        ch,
	}

	c.newJobs <- job // set a place in mapping

	if q.Printer != nil {
		exchange.printSend(q.Printer)
	}

	e := c.send(message)
	if e != nil {
		job.CloseErr(e)

		// release the spot reserved if not timed out
		c.sendErrors <- job
	}
}

func (c *Client) send(m *Message) error {
	_, e := c.conn.WriteToUDP(m.Packet.Bytes, m.RemoteAddr)
	return e
}

// AsyncQuery sends a query and call-back f with the exchange.
func (c *Client) AsyncQuery(q *QueryPrinter, f func(*Exchange)) {
	ch := make(chan *Exchange)
	go func() {
		f(<-ch)
	}()

	c.Send(q, ch)
}

// Query queries the query and returns the exchange.
func (c *Client) Query(q *QueryPrinter) *Exchange {
	ch := make(chan *Exchange, 1) // we need a slot in case of send error
	c.Send(q, ch)

	return <-ch
}
