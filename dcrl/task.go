package dcrl

import (
	"bytes"

	"lonnie.io/dig8/dns8"
)

// task is a query task arround one single domain
type task struct {
	domain *dns8.Domain
	client *dns8.Client

	res string // result
	out string // output
	log string // log
	err string // error
}

func (t *task) run() {
	logBuf := new(bytes.Buffer)
	tm := dns8.NewTerm(t.client)
	tm.Log = logBuf

	info := dns8.NewInfo(t.domain)
	_, err := tm.T(info)

	if err == nil {
		t.out = info.Out()
		t.res = info.Result()
	} else {
		t.err = err.Error()
	}

	t.log = logBuf.String()
}
