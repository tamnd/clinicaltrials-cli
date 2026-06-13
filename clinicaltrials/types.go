package clinicaltrials

import "strings"

// Trial is the record emitted for list commands (search, recruiting, conditions).
type Trial struct {
	Rank       int    `json:"rank"`
	NCTID      string `json:"nct_id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Phase      string `json:"phase"`
	Conditions string `json:"conditions"`
	Sponsor    string `json:"sponsor"`
	StartDate  string `json:"start_date"`
	URL        string `json:"url"`
}

// TrialDetail is the richer record emitted by the trial command.
type TrialDetail struct {
	Rank           int    `json:"rank"`
	NCTID          string `json:"nct_id"`
	Title          string `json:"title"`
	OfficialTitle  string `json:"official_title"`
	Status         string `json:"status"`
	Phase          string `json:"phase"`
	StudyType      string `json:"study_type"`
	Conditions     string `json:"conditions"`
	Sponsor        string `json:"sponsor"`
	SponsorClass   string `json:"sponsor_class"`
	StartDate      string `json:"start_date"`
	CompletionDate string `json:"completion_date"`
	LastUpdate     string `json:"last_update"`
	Summary        string `json:"summary"`
	Sex            string `json:"sex"`
	MinAge         string `json:"min_age"`
	MaxAge         string `json:"max_age"`
	Locations      string `json:"locations"`
	URL            string `json:"url"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireResponse struct {
	TotalCount    int         `json:"totalCount"`
	NextPageToken string      `json:"nextPageToken"`
	Studies       []wireStudy `json:"studies"`
}

type wireStudy struct {
	ProtocolSection wireProtocol `json:"protocolSection"`
}

type wireProtocol struct {
	IdentificationModule       wireIdentification    `json:"identificationModule"`
	StatusModule               wireStatus            `json:"statusModule"`
	ConditionsModule           wireConditions        `json:"conditionsModule"`
	DesignModule               wireDesign            `json:"designModule"`
	SponsorCollaboratorsModule wireSponsorCollabs    `json:"sponsorCollaboratorsModule"`
	DescriptionModule          wireDescription       `json:"descriptionModule"`
	EligibilityModule          wireEligibility       `json:"eligibilityModule"`
	ContactsLocationsModule    wireContactsLocations `json:"contactsLocationsModule"`
}

type wireIdentification struct {
	NctId         string `json:"nctId"`
	BriefTitle    string `json:"briefTitle"`
	OfficialTitle string `json:"officialTitle"`
}

type wireStatus struct {
	OverallStatus            string         `json:"overallStatus"`
	StartDateStruct          wireDateStruct `json:"startDateStruct"`
	CompletionDateStruct     wireDateStruct `json:"completionDateStruct"`
	LastUpdatePostDateStruct wireDateStruct `json:"lastUpdatePostDateStruct"`
}

type wireDateStruct struct {
	Date string `json:"date"`
}

type wireConditions struct {
	Conditions []string `json:"conditions"`
}

type wireDesign struct {
	Phases    []string `json:"phases"`
	StudyType string   `json:"studyType"`
}

type wireSponsorCollabs struct {
	LeadSponsor   wireSponsorEntry   `json:"leadSponsor"`
	Collaborators []wireSponsorEntry `json:"collaborators"`
}

type wireSponsorEntry struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}

type wireDescription struct {
	BriefSummary string `json:"briefSummary"`
}

type wireEligibility struct {
	Sex        string `json:"sex"`
	MinimumAge string `json:"minimumAge"`
	MaximumAge string `json:"maximumAge"`
}

type wireContactsLocations struct {
	Locations []wireLocation `json:"locations"`
}

type wireLocation struct {
	Country  string `json:"country"`
	Facility string `json:"facility"`
	City     string `json:"city"`
	State    string `json:"state"`
}

// ─── mapping helpers ─────────────────────────────────────────────────────────

func wireToTrial(s wireStudy, rank int) Trial {
	p := s.ProtocolSection
	return Trial{
		Rank:       rank,
		NCTID:      p.IdentificationModule.NctId,
		Title:      p.IdentificationModule.BriefTitle,
		Status:     p.StatusModule.OverallStatus,
		Phase:      joinPhases(p.DesignModule.Phases),
		Conditions: joinFirst(p.ConditionsModule.Conditions, 3, "; "),
		Sponsor:    p.SponsorCollaboratorsModule.LeadSponsor.Name,
		StartDate:  p.StatusModule.StartDateStruct.Date,
		URL:        studyURL(p.IdentificationModule.NctId),
	}
}

func wireToTrialDetail(s wireStudy) TrialDetail {
	p := s.ProtocolSection
	return TrialDetail{
		Rank:           1,
		NCTID:          p.IdentificationModule.NctId,
		Title:          p.IdentificationModule.BriefTitle,
		OfficialTitle:  p.IdentificationModule.OfficialTitle,
		Status:         p.StatusModule.OverallStatus,
		Phase:          joinPhases(p.DesignModule.Phases),
		StudyType:      p.DesignModule.StudyType,
		Conditions:     joinFirst(p.ConditionsModule.Conditions, 3, "; "),
		Sponsor:        p.SponsorCollaboratorsModule.LeadSponsor.Name,
		SponsorClass:   p.SponsorCollaboratorsModule.LeadSponsor.Class,
		StartDate:      p.StatusModule.StartDateStruct.Date,
		CompletionDate: p.StatusModule.CompletionDateStruct.Date,
		LastUpdate:     p.StatusModule.LastUpdatePostDateStruct.Date,
		Summary:        truncateSummary(p.DescriptionModule.BriefSummary, 200),
		Sex:            p.EligibilityModule.Sex,
		MinAge:         p.EligibilityModule.MinimumAge,
		MaxAge:         p.EligibilityModule.MaximumAge,
		Locations:      firstCountries(p.ContactsLocationsModule.Locations, 3),
		URL:            studyURL(p.IdentificationModule.NctId),
	}
}

func studyURL(nctID string) string {
	return "https://clinicaltrials.gov/study/" + nctID
}

func joinPhases(phases []string) string {
	return strings.Join(phases, "/")
}

func joinFirst(ss []string, n int, sep string) string {
	if len(ss) > n {
		ss = ss[:n]
	}
	return strings.Join(ss, sep)
}

func firstCountries(locs []wireLocation, n int) string {
	seen := map[string]bool{}
	var out []string
	for _, l := range locs {
		if l.Country == "" || seen[l.Country] {
			continue
		}
		seen[l.Country] = true
		out = append(out, l.Country)
		if len(out) >= n {
			break
		}
	}
	return strings.Join(out, "; ")
}

func truncateSummary(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// normalizeNCT uppercases and ensures the NCT prefix.
func normalizeNCT(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	if !strings.HasPrefix(id, "NCT") {
		id = "NCT" + id
	}
	return id
}
