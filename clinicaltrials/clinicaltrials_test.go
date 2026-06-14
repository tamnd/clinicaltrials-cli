package clinicaltrials_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tamnd/clinicaltrials-cli/clinicaltrials"
)

func newTestClient(ts *httptest.Server) *clinicaltrials.Client {
	cfg := clinicaltrials.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return clinicaltrials.NewClient(cfg)
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
      "completionDateStruct": { "date": "2025-12-31" }
    },
    "conditionsModule": { "conditions": ["Diabetes", "Hypertension"] },
    "designModule": { "phases": ["PHASE3"], "studyType": "INTERVENTIONAL", "enrollmentInfo": { "count": 100 } },
    "sponsorCollaboratorsModule": {
      "leadSponsor": { "name": "NIH", "class": "NIH" }
    },
    "descriptionModule": { "briefSummary": "A study summary." }
  }
}`

func TestGetSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(`{"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "test", "", "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchReturnsStudies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"studies":[` + minimalStudyJSON + `]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	studies, err := c.Search(context.Background(), "diabetes", "", "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(studies) != 1 {
		t.Fatalf("got %d studies, want 1", len(studies))
	}
	if studies[0].NCTID != "NCT12345678" {
		t.Errorf("NCTID = %q, want NCT12345678", studies[0].NCTID)
	}
	if studies[0].OverallStatus != "RECRUITING" {
		t.Errorf("OverallStatus = %q, want RECRUITING", studies[0].OverallStatus)
	}
	if studies[0].Phase != "PHASE3" {
		t.Errorf("Phase = %q, want PHASE3", studies[0].Phase)
	}
	if studies[0].Enrollment != 100 {
		t.Errorf("Enrollment = %d, want 100", studies[0].Enrollment)
	}
	if studies[0].LeadSponsor != "NIH" {
		t.Errorf("LeadSponsor = %q, want NIH", studies[0].LeadSponsor)
	}
}

func TestSearchFollowsPagination(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = w.Write([]byte(`{"nextPageToken":"page2","studies":[` + minimalStudyJSON + `]}`))
		} else {
			s := strings.ReplaceAll(minimalStudyJSON, "NCT12345678", "NCT87654321")
			_, _ = w.Write([]byte(`{"studies":[` + s + `]}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	studies, err := c.Search(context.Background(), "test", "", "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(studies) != 2 {
		t.Fatalf("got %d studies, want 2", len(studies))
	}
	if calls < 2 {
		t.Errorf("expected at least 2 HTTP calls (pagination), got %d", calls)
	}
	if studies[1].NCTID != "NCT87654321" {
		t.Errorf("second study NCTID = %q, want NCT87654321", studies[1].NCTID)
	}
}

func TestGetStudyByID(t *testing.T) {
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
	study, err := c.GetStudy(context.Background(), "NCT04368728")
	if err != nil {
		t.Fatal(err)
	}
	if study.NCTID != "NCT04368728" {
		t.Errorf("NCTID = %q, want NCT04368728", study.NCTID)
	}
	if study.OverallStatus != "COMPLETED" {
		t.Errorf("OverallStatus = %q, want COMPLETED", study.OverallStatus)
	}
}

func TestGetStudyNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetStudy(context.Background(), "NCT00000000")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestSearchStatusFilter(t *testing.T) {
	var gotStatus string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotStatus = r.URL.Query().Get("filter.overallStatus")
		_, _ = w.Write([]byte(`{"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "", "", "", "RECRUITING", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotStatus != "RECRUITING" {
		t.Errorf("filter.overallStatus = %q, want RECRUITING", gotStatus)
	}
}

func TestSearchConditionParam(t *testing.T) {
	var gotCond string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCond = r.URL.Query().Get("query.cond")
		_, _ = w.Write([]byte(`{"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "cancer", "", "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotCond != "cancer" {
		t.Errorf("query.cond = %q, want cancer", gotCond)
	}
}

func TestSearchInterventionParam(t *testing.T) {
	var gotIntr string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIntr = r.URL.Query().Get("query.intr")
		_, _ = w.Write([]byte(`{"studies":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "", "insulin", "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotIntr != "insulin" {
		t.Errorf("query.intr = %q, want insulin", gotIntr)
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
		_, _ = w.Write([]byte(`{"studies":[]}`))
	}))
	defer srv.Close()

	cfg := clinicaltrials.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := clinicaltrials.NewClient(cfg)

	start := time.Now()
	_, err := c.Search(context.Background(), "test", "", "", "", 5)
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
