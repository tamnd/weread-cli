package weread

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "weread" {
		t.Errorf("Scheme = %q, want weread", info.Scheme)
	}
	found := false
	for _, h := range info.Hosts {
		if h == "weread.qq.com" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Hosts = %v, want to contain weread.qq.com", info.Hosts)
	}
	if info.Identity.Binary != "weread" {
		t.Errorf("Identity.Binary = %q, want weread", info.Identity.Binary)
	}
}

func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	domains := h.Domains()
	found := false
	for _, d := range domains {
		if d == "weread" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("weread domain not registered; got %v", domains)
	}
}
