package dcrl

import (
	"bufio"
	"os"
	"strings"

	"lonnie.io/dig8/dns8"
)

// ReadDomains reads a domain list file into a list of domains
func ReadDomains(f string) ([]*dns8.Domain, error) {
	fin, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	defer fin.Close()

	s := bufio.NewScanner(fin)
	var ret []*dns8.Domain

	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		d, e := dns8.ParseDomain(line)
		if e != nil {
			return nil, e
		}
		ret = append(ret, d)
	}

	e = s.Err()
	if e != nil {
		return nil, e
	}

	return ret, nil
}

// ReadDomainStrings reads a domain list file into a list of strings
func ReadDomainStrings(f string) ([]string, error) {
	doms, e := ReadDomains(f)
	if e != nil {
		return nil, e
	}

	ret := make([]string, len(doms))
	for i, d := range doms {
		ret[i] = d.String()
	}

	return ret, nil
}
