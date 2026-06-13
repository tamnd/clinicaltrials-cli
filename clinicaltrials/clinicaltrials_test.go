package clinicaltrials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(ts *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return NewClient(cfg)
}

const minimalStudyJSON = `{
  "protocolSection": {
    "identificationModule": {
      "nctId": "NCT12345678",
      "briefTitle": "Test Trial",
      "officialTitle": "Full Official Test Trial Title"
    },
    "statusModule": {
      "overallStatus": "RECRUITING",
      "startDateStruct": { "date": "2023-01-01" },
      "completionDateStruct": { "date": "2025-12-31" },
      "lastUpdatePostDateStruct": { "date": "2024-06-01" }
    },
    "conditionsModule": { "conditions": ["Diabetes", "Hypertension"] },
    "designModule": { "phases": ["PHASE3"], "studyType": "INTERVENTIONAL" },
    "sponsorCollaboratorsModule": {
      "leadSponsor": { "name": "NIH", "class": "NIH" }
    },
    "descriptionModule": { "briefSummary": "A study summary." },
    "eligibilityModule": { "sex": "ALL", "minimumAge": "18 Years", "maximumAge": "75 Years" },
    "contactsLocationsModule": {
      "locations": [{ "country": "United States", "facility": "Test Center" }]
    }
  }
}`

func TestGetSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(`{"totalCount":0,"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "test", "", 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchReturnsTrials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"totalCount":1,"studies":[` + minimalStudyJSON + `]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	trials, err := c.Search(context.Background(), "diabetes", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(trials) != 1 {
		t.Fatalf("got %d trials, want 1", len(trials))
	}
	if trials[0].NCTID != "NCT12345678" {
		t.Errorf("NCTID = %q, want NCT12345678", trials[0].NCTID)
	}
	if trials[0].Rank != 1 {
		t.Errorf("Rank = %d, want 1", trials[0].Rank)
	}
	if trials[0].Status != "RECRUITING" {
		t.Errorf("Status = %q, want RECRUITING", trials[0].Status)
	}
}

func TestSearchFollowsPagination(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("pageToken") == "" {
			// first page
			_, _ = w.Write([]byte(`{"totalCount":2,"nextPageToken":"page2","studies":[` + minimalStudyJSON + `]}`))
		} else {
			// second page — return a study with a different NCT ID
			s := strings.ReplaceAll(minimalStudyJSON, "NCT12345678", "NCT87654321")
			_, _ = w.Write([]byte(`{"totalCount":2,"studies":[` + s + `]}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	trials, err := c.Search(context.Background(), "test", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(trials) != 2 {
		t.Fatalf("got %d trials, want 2", len(trials))
	}
	if calls < 2 {
		t.Errorf("expected at least 2 HTTP calls (pagination), got %d", calls)
	}
	if trials[1].NCTID != "NCT87654321" {
		t.Errorf("second trial NCTID = %q, want NCT87654321", trials[1].NCTID)
	}
}

func TestTrialByID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "NCT04368728") {
			http.NotFound(w, r)
			return
		}
		s := strings.ReplaceAll(minimalStudyJSON, "NCT12345678", "NCT04368728")
		s = strings.ReplaceAll(s, "RECRUITING", "COMPLETED")
		_, _ = w.Write([]byte(s))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	detail, err := c.Trial(context.Background(), "NCT04368728")
	if err != nil {
		t.Fatal(err)
	}
	if detail.NCTID != "NCT04368728" {
		t.Errorf("NCTID = %q, want NCT04368728", detail.NCTID)
	}
	if detail.Status != "COMPLETED" {
		t.Errorf("Status = %q, want COMPLETED", detail.Status)
	}
	if detail.URL != "https://clinicaltrials.gov/study/NCT04368728" {
		t.Errorf("URL = %q", detail.URL)
	}
}

func TestRecruitingFilter(t *testing.T) {
	var gotStatus string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotStatus = r.URL.Query().Get("filter.overallStatus")
		_, _ = w.Write([]byte(`{"totalCount":0,"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Recruiting(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotStatus != "RECRUITING" {
		t.Errorf("filter.overallStatus = %q, want RECRUITING", gotStatus)
	}
}

func TestNormalizeNCT(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"nct04368728", "NCT04368728"},
		{"NCT04368728", "NCT04368728"},
		{"04368728", "NCT04368728"},
	}
	for _, tc := range cases {
		got := normalizeNCT(tc.in)
		if got != tc.want {
			t.Errorf("normalizeNCT(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"totalCount":0,"studies":[]}`))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClient(cfg)

	start := time.Now()
	_, err := c.Search(context.Background(), "test", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}
