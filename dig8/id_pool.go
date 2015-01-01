package dig8

import (
	"math/rand"
)

const (
	idCount  = 65536
	nprepare = idCount / 4
)

// idPool creates a pool if ids for DNS crawling
type idPool struct {
	using    []bool
	nusing   int
	returns  chan uint16
	prepared chan uint16
	rand     *rand.Rand
}

func newIDPool() *idPool {
	ret := new(idPool)
	ret.using = make([]bool, idCount)
	ret.returns = make(chan uint16, 10)
	ret.prepared = make(chan uint16, nprepare)

	// TODO: find a better random source
	src := rand.NewSource(0)
	ret.rand = rand.New(src)

	go ret.serve()

	return ret
}

func (p *idPool) pick() uint16 {
	for {
		ret := uint16(p.rand.Uint32())
		if !p.using[ret] {
			return ret
		}
	}
}

func (p *idPool) prepare() {
	id := p.pick()
	p.using[id] = true
	p.nusing++
	p.prepared <- id
}

func (p *idPool) free(id uint16) bool {
	if !p.using[id] {
		return false
	}

	p.nusing--
	p.using[id] = false
	return true
}

func (p *idPool) serve() {
	for i := 0; i < nprepare; i++ {
		p.prepare()
	}

	for r := range p.returns {
		if p.free(r) {
			p.prepare()
		}
		bugOn(p.nusing != nprepare)
	}
}

// Fetch fetches a new id from the pool.
func (p *idPool) Fetch() uint16 {
	return <-p.prepared
}

// Return returns an id back to the pool.
func (p *idPool) Return(id uint16) {
	p.returns <- id
}
