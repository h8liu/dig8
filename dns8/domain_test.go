package dns8

import (
	"testing"
)

func sameLabels(labs1, labs2 []string) bool {
	if len(labs1) != len(labs2) {
		return false
	}
	for i, lab := range labs1 {
		if labs2[i] != lab {
			return false
		}
	}
	return true
}

func TestDomainParse(t *testing.T) {
	v := func(input, name string, labels []string) {
		d, e := Parse(input)
		if e != nil {
			t.Error("err on valid domain:", input, e)
			return
		}
		if d.name != name {
			t.Errorf("name of %s, expect: %s, got %s",
				input, name, d.name)
		}
		if !sameLabels(d.labels, labels) {
			t.Errorf("labels of %s, expect: %s, got %s",
				name, labels, d.labels)
		}
	}

	v("", "", []string{})
	v(".", "", []string{})
	v("a.b.", "a.b", []string{"a", "b"})
	v("a.b", "a.b", []string{"a", "b"})
	v("A.b", "a.b", []string{"a", "b"})
	v("A.B.", "a.b", []string{"a", "b"})
	v("www.google.com.", "www.google.com", []string{"www", "google", "com"})
	v("liulonnie.net", "liulonnie.net", []string{"liulonnie", "net"})
	v("3721.net", "3721.net", []string{"3721", "net"})

	iv := func(input string) {
		d, e := Parse(input)
		if d != nil || e == nil {
			t.Error("silent on invalid domain:", input)
		}
	}

	iv("asdf-.net")
	iv("-asdf.net")
	iv("飞雪.org")
	iv("-.-")
	iv("liulonnie@gmail.com")
	iv("192.168.7.232")
}

func TestDomainReg(t *testing.T) {
	r := func(input, reged, regtr *Domain) {
		_reged, _regtr := input.RegParts()
		if !reged.Equal(_reged) {
			t.Errorf("reged of %s: expect %s, got %s",
				input, reged, _reged)
		}

		if !regtr.Equal(_regtr) {
			t.Errorf("regtr of %s: expect %s, got %s",
				input, regtr, _regtr)
		}
	}

	r(D("www.google.com"), D("google.com"), D("com"))
	r(D("liulonnie.net"), D("liulonnie.net"), D("net"))
	r(D("www.yahoo.com.cn"), D("yahoo.com.cn"), D("com.cn"))
	r(D("www.yahoo.edu.ru"), D("yahoo.edu.ru"), D("edu.ru"))
	r(D("www.yahoo.edu.ru"), D("yahoo.edu.ru"), D("edu.ru"))
	r(D("co"), nil, D("co"))
	r(D("au"), nil, D("au"))
	r(D("bd"), nil, D("bd"))
	r(D("cn"), nil, D("cn"))
	r(D("uba.ar"), nil, D("uba.ar"))
	r(D("t.uba.ar"), D("t.uba.ar"), D("uba.ar"))

	o := func(c, p *Domain) {
		if !c.IsChildOf(p) || !p.IsParentOf(c) {
			t.Errorf("failed on relative: %v %v", p, c)
		}
	}
	o(D("www.google.com"), D("google.com"))
	o(D("www.google.com"), D("com"))
	o(D("www.google.com"), D("."))
}
