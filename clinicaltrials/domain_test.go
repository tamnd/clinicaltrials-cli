package clinicaltrials

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "clinicaltrials" {
		t.Errorf("scheme = %q, want clinicaltrials", info.Scheme)
	}
	if info.Identity.Binary != "clinicaltrials" {
		t.Errorf("binary = %q, want clinicaltrials", info.Identity.Binary)
	}
	found := false
	for _, h := range info.Hosts {
		if h == Host {
			found = true
		}
	}
	if !found {
		t.Errorf("hosts = %v, want to contain %q", info.Hosts, Host)
	}
}

func TestClassify(t *testing.T) {
	d := Domain{}

	typ, id, err := d.Classify("NCT05608876")
	if err != nil || typ != "nctid" || id != "NCT05608876" {
		t.Errorf("Classify(NCT05608876) = %q/%q/%v, want nctid/NCT05608876/nil", typ, id, err)
	}

	typ, id, err = d.Classify("nct05608876")
	if err != nil || typ != "nctid" || id != "NCT05608876" {
		t.Errorf("Classify(lowercase nct) = %q/%q/%v, want nctid/NCT05608876/nil", typ, id, err)
	}

	typ, id, err = d.Classify("cancer")
	if err != nil || typ != "query" || id != "cancer" {
		t.Errorf("Classify(cancer) = %q/%q/%v, want query/cancer/nil", typ, id, err)
	}

	_, _, err = d.Classify("")
	if err == nil {
		t.Error("Classify('') = nil error, want error")
	}
}

func TestLocate(t *testing.T) {
	d := Domain{}

	u, err := d.Locate("nctid", "NCT05608876")
	if err != nil || u != "https://clinicaltrials.gov/study/NCT05608876" {
		t.Errorf("Locate(nctid) = %q/%v", u, err)
	}

	u, err = d.Locate("query", "cancer")
	if err != nil || u != "https://clinicaltrials.gov/search?query.term=cancer" {
		t.Errorf("Locate(query) = %q/%v", u, err)
	}

	_, err = d.Locate("unknown", "foo")
	if err == nil {
		t.Error("Locate(unknown) = nil error, want error")
	}
}

func TestDomainRegistered(t *testing.T) {
	// init() registered the domain; kit.Open should find it.
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Domain("clinicaltrials"); !ok {
		t.Fatal("clinicaltrials domain not registered")
	}
}
