package dns8

import (
	"math/rand"
	"time"
)

// TODO: use a better entropy source
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func shuffleAppend(ret, list []*NameServer) []*NameServer {
	n := len(list)
	if n == 0 {
		return make([]*NameServer, 0)
	}
	order := random.Perm(n)
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
	order := random.Perm(n)
	for i := 0; i < n; i++ {
		ret[i] = list[order[i]]
	}
	return ret
}
