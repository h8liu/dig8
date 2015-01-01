package dig8

import (
	"math/rand"
	"time"
)

// TODO: use a better entropy source
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func shuffleAppend(ret, list []*NameServer) []*NameServer {
	n := len(list)
	order := random.Perm(n)
	for i := 0; i < n; i++ {
		ret = append(ret, list[order[i]])
	}
	return ret
}

func shuffleList(list []*NameServer) []*NameServer {
	n := len(list)
	ret := make([]*NameServer, n)
	order := random.Perm(n)
	for i := 0; i < n; i++ {
		ret[i] = list[order[i]]
	}
	return ret
}
