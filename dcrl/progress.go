package dcrl

import (
	"bytes"
	"fmt"
)

// Progress reports how much a job is crawled
type Progress struct {
	Name    string
	Crawled int
	Total   int
	Done    bool
	Error   string
}

func (p *Progress) String() string {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "[%s] ", p.Name)
	if p.Error != "" {
		fmt.Fprintf(buf, "error: %s", p.Error)
	} else if p.Done {
		fmt.Fprintf(buf, "done (%d domains)", p.Total)
	} else {
		fmt.Fprintf(buf, "%d/%d", p.Crawled, p.Total)
	}

	return buf.String()
}
