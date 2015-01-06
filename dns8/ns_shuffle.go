package dns8

import (
	"math/rand"
	"sync"
	"time"
)

// TODO: use a better entropy source
var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var randLock sync.Mutex

func shuffleAppend(ret, list []*NameServer) []*NameServer {
	n := len(list)
	if n == 0 {
		return make([]*NameServer, 0)
	}
	randLock.Lock()
	order := random.Perm(n)
	randLock.Unlock()
	for i := 0; i < n; i++ {
		ret = append(ret, list[order[i]])
	}
	return ret
}

func shuffleList(list []*NameServer) []*NameServer {
	n := len(list)
	if n == 0 {
		return make([]*NameServer, 0)
	}
	ret := make([]*NameServer, n)
	randLock.Lock()
	order := random.Perm(n)
	randLock.Unlock()
	for i := 0; i < n; i++ {
		ret[i] = list[order[i]]
	}
	return ret
}
