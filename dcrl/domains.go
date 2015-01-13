package dcrl

import (
	"bufio"
	"fmt"
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

// WriteDomains writes a list of domains into a file
func WriteDomains(f string, doms []*dns8.Domain) error {
	fout, e := os.Create(f)
	if e != nil {
		return e
	}

	for _, d := range doms {
		_, e = fmt.Fprintln(fout, d.String())
		if e != nil {
			fout.Close()
			return e
		}
	}

	return fout.Close()
}

// WriteDomainStrings writes a list of domain strings to a file
func WriteDomainStrings(f string, doms []string) error {
	fout, e := os.Create(f)
	if e != nil {
		return e
	}

	for _, d := range doms {
		_, e = fmt.Fprintln(fout, d)
		if e != nil {
			fout.Close()
			return e
		}
	}

	return fout.Close()
}
